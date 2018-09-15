// Copyright 2018 HouseCanary, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package structschema

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

func areTypesEqual(a, b ast.Type) bool {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	switch t := a.(type) {
	case *ast.SimpleType:
		return t.Name == (b.(*ast.SimpleType)).Name
	case *ast.ListType:
		return areTypesEqual(t.Of, (b.(*ast.ListType)).Of)
	case *ast.NotNilType:
		return areTypesEqual(t.Of, (b.(*ast.ListType)).Of)
	}

	panic("Unknown type")
}

func (m *fieldMeta) validateType(b *Builder, returnType reflect.Type) error {
	schemaReturnType, err := b.goTypeToSchemaType(returnType)
	if err != nil {
		return err
	}

	expectedType := stripNotNil(m.GqlField.Type)

	if !areTypesEqual(schemaReturnType, expectedType) {
		return fmt.Errorf("Returned type of field %s is not compatible with schema type (got: %v, expected: %v)", m.Name, schemaReturnType.Signature(), expectedType.Signature())
	}
	return nil
}

func stripNotNil(typ ast.Type) ast.Type {
	switch t := typ.(type) {
	case *ast.NotNilType:
		return stripNotNil(t.ContainedType())
	case *ast.ListType:
		return &ast.ListType{Of: stripNotNil(t.ContainedType())}
	default:
		return typ
	}
}

func (m *fieldMeta) buildFieldResolver(b *Builder, structTyp reflect.Type, name string) error {
	typeName := structTyp.Name()

	pointerType := reflect.PtrTo(structTyp)
	field, ok := structTyp.FieldByName(name)
	if !ok {
		return fmt.Errorf("Struct %s does not have field %s", structTyp.Name(), name)
	}
	index := field.Index
	if err := m.validateType(b, field.Type); err != nil {
		return fmt.Errorf("Error creating field resolver for %s:%s - %v", typeName, name, err)
	}

	isListType := field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array

	m.Resolver = schema.ContextResolver(func(ctx context.Context, v interface{}) (interface{}, error) {
		rv := reflect.ValueOf(v)
		if rv.Type() == pointerType {
			rv = rv.Elem()
		}
		if rv.Type() != structTyp {
			return nil, fmt.Errorf("%v(%T) was not of type %v", v, v, structTyp)
		}
		fv := rv.FieldByIndex(index)
		kind := fv.Kind()
		if fv.CanAddr() && kind != reflect.Ptr && !isListType {
			fv = fv.Addr()
		}
		if isListType {
			return toList(fv), nil
		}
		return fv.Interface(), nil
	})
	return nil
}

func (m *fieldMeta) buildMethodResolver(b *Builder, structTyp reflect.Type, name string) error {
	typeName := structTyp.Name()

	pointerType := reflect.PtrTo(structTyp)
	var method *reflect.Method
	for i := 0; i < pointerType.NumMethod(); i++ {
		m := pointerType.Method(i)
		if strings.EqualFold(m.Name, "resolve"+name) {
			method = &m
		}
	}

	if method == nil {
		return fmt.Errorf("Cannot find a resolver method for field %s on struct %s: Expected a method of the form Resolve[Field Name]", name, structTyp.Name())
	}

	mTyp := method.Type
	methodName := method.Name

	argStart := 1 // Start at 1, the first param is the receiver
	var argResolvers []argResolver
	var needsResolverContext bool

	for i := argStart; i < mTyp.NumIn(); i++ {
		argType := mTyp.In(i)

		if argType == resolverContextType {
			argStart++
			needsResolverContext = true
			argResolvers = append(argResolvers, func(ctx context.Context) (reflect.Value, error) {
				rc := ctx.(schema.ResolverContext)
				return reflect.ValueOf(rc), nil
			})
		} else if argType == contextType {
			argStart++
			argResolvers = append(argResolvers, func(ctx context.Context) (reflect.Value, error) {
				return reflect.ValueOf(ctx), nil
			})
		} else {
			argSig := argType.String()
			if argProvider, ok := b.argProviders[argSig]; ok {
				argStart++
				argResolvers = append(argResolvers, func(ctx context.Context) (reflect.Value, error) {
					argVal := argProvider(ctx)
					return reflect.ValueOf(argVal), nil
				})
			} else {
				break
			}
		}
	}

	numNamedArgs := mTyp.NumIn() - argStart
	if numNamedArgs != len(m.GqlField.ArgumentsDefinition) {
		if !(numNamedArgs == 0 && needsResolverContext) { // We'll allow a mismatched arg signature if the resolver func takes a ResolverContext arg and no graphql named args
			return fmt.Errorf("In type %s, field %s defines %v arguments, but method %s receives %v", typeName, name, len(m.GqlField.ArgumentsDefinition), methodName, numNamedArgs)
		}
	}

	namedArgIndex := 0
	for i := argStart; i < mTyp.NumIn(); i++ {
		argTyp := mTyp.In(i)
		_, err := b.goTypeToSchemaType(argTyp)
		if err != nil {
			return fmt.Errorf("Error resolving argument %d of resolver method for %s:%s: %v", i+1, typeName, name, err)
		}
		argDef := m.GqlField.ArgumentsDefinition[namedArgIndex]
		argName := argDef.Name
		argHasPointerType := argTyp.Kind() == reflect.Ptr
		argResolvers = append(argResolvers, func(ctx context.Context) (reflect.Value, error) {
			rc := ctx.(schema.ResolverContext)
			v, err := rc.GetArgumentValue(argName)
			if err != nil {
				return reflect.ValueOf(nil), fmt.Errorf("Error resolving argument %s: %v", argName, err)
			}

			if v == nil {
				return reflect.Zero(argTyp), nil
			}

			rv := reflect.ValueOf(v)

			// If arg doesn't have a pointer type, need to indirect
			if !argHasPointerType {
				rv = reflect.Indirect(rv)
			}

			return rv, nil
		})
		namedArgIndex++
	}

	rh, err := m.makeResultHandler(b, mTyp, typeName, name)

	if err != nil {
		return fmt.Errorf("Error creating result handler on resolver method for %s:%s: %v", typeName, name, err)
	}

	receiverTyp := mTyp.In(0)

	if len(argResolvers) == 0 {
		// No args to resolve.  Use a simple resolver
		m.Resolver = schema.ContextResolver(func(ctx context.Context, v interface{}) (interface{}, error) {
			receiver := coerceReceiver(reflect.ValueOf(v), receiverTyp)
			fun := method.Func
			args := []reflect.Value{receiver}
			return rh(ctx, fun.Call(args)...)
		})
	} else if needsResolverContext || numNamedArgs > 0 {
		m.Resolver = schema.FullResolver(func(ctx schema.ResolverContext, v interface{}) (interface{}, error) {
			fun := method.Func
			args := make([]reflect.Value, len(argResolvers)+1) // +1 to reserve space for receiver
			receiver := coerceReceiver(reflect.ValueOf(v), receiverTyp)
			args[0] = receiver
			for i, ar := range argResolvers {
				if a, err := ar(ctx); err == nil {
					args[i+1] = a
				} else {
					return nil, err
				}
			}
			return rh(ctx, fun.Call(args)...)
		})
	} else {
		m.Resolver = schema.ContextResolver(func(ctx context.Context, v interface{}) (interface{}, error) {
			fun := method.Func
			receiver := coerceReceiver(reflect.ValueOf(v), receiverTyp)
			args := make([]reflect.Value, len(argResolvers)+1) // +1 to reserve space for receiver
			args[0] = receiver
			for i, ar := range argResolvers {
				if a, err := ar(ctx); err == nil {
					args[i+1] = a
				} else {
					return nil, err
				}
			}
			return rh(ctx, fun.Call(args)...)
		})
	}

	return nil
}

func coerceReceiver(receiver reflect.Value, expected reflect.Type) reflect.Value {
	rTyp := receiver.Type()
	if rTyp != expected {
		if expected.Kind() == reflect.Ptr && rTyp == expected.Elem() {
			if receiver.CanAddr() {
				receiver = receiver.Addr()
			} else {
				t := reflect.New(rTyp)
				t.Elem().Set(receiver)
				receiver = t
			}
		} else if rTyp.Kind() == reflect.Ptr && rTyp.Elem() == expected {
			receiver = receiver.Elem()
		} else {
			panic(fmt.Errorf("Value not coercible to receiver: %v", receiver))
		}
	}

	return receiver
}

func (m *fieldMeta) makeResultHandler(b *Builder, typ reflect.Type, typeName, fieldName string) (resultHandler, error) {
	switch typ.NumOut() {
	case 1:
		out := typ.Out(0)
		if out.Kind() == reflect.Chan {
			if err := m.validateType(b, out.Elem()); err != nil {
				return nil, err
			}
			return buildAsyncResultHandler(out.Elem(), typeName, fieldName), nil
		} else if out.Kind() == reflect.Func {
			asyncResultHandler, err := m.makeResultHandler(b, out, typeName, fieldName)
			if err != nil {
				return nil, err
			}
			return buildAsyncFuncResultHandler(asyncResultHandler, typeName, fieldName), nil
		} else {
			if err := m.validateType(b, out); err != nil {
				return nil, err
			}
			return buildSimpleResultHandler(out), nil
		}
	case 2:
		out := typ.Out(0)
		if out.Kind() == reflect.Chan && typ.Out(1).Kind() == reflect.Chan {
			if err := m.validateType(b, out.Elem()); err != nil {
				return nil, err
			}
			return buildAsyncErrorResultHandler(out.Elem(), typeName, fieldName), nil
		} else if typ.Out(1).AssignableTo(errorType) {
			if err := m.validateType(b, out); err != nil {
				return nil, err
			}
			return buildErrorResultHandler(out), nil
		}
	}
	return nil, fmt.Errorf("Expected method return value to be one of (<value>) | (chan <value>) | (<value>, error) | (<-chan <value>, <-chan error) | (func () <value>) | (func () (<value>, error))")
}

type argResolver func(context.Context) (reflect.Value, error)

type funcAsyncValue struct {
	f                  reflect.Value
	asyncResultHandler resultHandler
	ctx                context.Context
}

func (v *funcAsyncValue) Await(ctx context.Context) (interface{}, error) {
	return v.asyncResultHandler(v.ctx, v.f.Call(nil)...)
}

type chanAsyncValue struct {
	typeName  string
	fieldName string
	c         reflect.Value
	wrap      func(reflect.Value) interface{}
}

func (v *chanAsyncValue) Await(ctx context.Context) (interface{}, error) {
	rv, ok := v.c.Recv()
	if !ok {
		return nil, fmt.Errorf("Channel receive failed, closed prematurely")
	}
	if v.wrap != nil {
		return v.wrap(rv), nil
	}
	return fixNil(rv), nil
}

type chanErrorAsyncValue struct {
	typeName  string
	fieldName string
	c         reflect.Value
	e         reflect.Value
	wrap      func(reflect.Value) interface{}
}

func (v *chanErrorAsyncValue) Await(ctx context.Context) (interface{}, error) {
	chosen, rv, ok := reflect.Select([]reflect.SelectCase{
		reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: v.c,
		},
		reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: v.e,
		},
	})

	switch chosen {
	case 0:
		if !ok {
			errValue, errOk := v.e.Recv()
			if !errOk {
				return nil, fmt.Errorf("Channel receive failed: result closed and err closed")
			}
			return nil, fixNilE(errValue)
		}
		if v.wrap != nil {
			return v.wrap(rv), nil
		}
		return fixNil(rv), nil
	case 1:
		if !ok {
			resultValue, resultOk := v.e.Recv()
			if !resultOk {
				return nil, fmt.Errorf("Channel receive failed: err closed and result closed")
			}
			return fixNil(resultValue), nil
		}
		return nil, fixNilE(rv)
	}

	// Unreachable code
	panic("Invalid selection")
}

type resultHandler func(ctx context.Context, result ...reflect.Value) (interface{}, error)

func fixNil(v reflect.Value) interface{} {
	switch v.Kind() {
	case reflect.Chan:
		fallthrough
	case reflect.Func:
		fallthrough
	case reflect.Interface:
		fallthrough
	case reflect.Map:
		fallthrough
	case reflect.Ptr:
		fallthrough
	case reflect.Slice:
		if v.IsNil() {
			return nil
		}
	}
	return v.Interface()
}

func fixNilE(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Chan:
		fallthrough
	case reflect.Func:
		fallthrough
	case reflect.Interface:
		fallthrough
	case reflect.Map:
		fallthrough
	case reflect.Ptr:
		fallthrough
	case reflect.Slice:
		if v.IsNil() {
			return nil
		}
	}
	return v.Interface().(error)
}

type reflectListValue struct {
	reflect.Value
}

func (v reflectListValue) ForEachElement(cb schema.ListValueCallback) {
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		if item.Kind() == reflect.Ptr && item.IsNil() {
			cb(nil)
		} else {
			cb(item.Interface())
		}
	}
}

func toList(v reflect.Value) schema.ListValue {
	if v.IsNil() {
		return nil
	}

	return reflectListValue{v}
}

func buildSimpleResultHandler(resultTyp reflect.Type) resultHandler {
	if resultTyp.Kind() == reflect.Slice || resultTyp.Kind() == reflect.Array {
		return func(ctx context.Context, result ...reflect.Value) (interface{}, error) {
			rv := result[0]
			return toList(rv), nil
		}
	}
	return func(ctx context.Context, result ...reflect.Value) (interface{}, error) {
		return fixNil(result[0]), nil
	}
}

func buildErrorResultHandler(resultTyp reflect.Type) resultHandler {
	if resultTyp.Kind() == reflect.Slice || resultTyp.Kind() == reflect.Array {
		return func(ctx context.Context, result ...reflect.Value) (interface{}, error) {
			return toList(result[0]), fixNilE(result[1])
		}
	}
	return func(ctx context.Context, result ...reflect.Value) (interface{}, error) {
		return fixNil(result[0]), fixNilE(result[1])
	}
}

func buildAsyncResultHandler(resultTyp reflect.Type, typeName, fieldName string) resultHandler {
	var wrap func(v reflect.Value) interface{}
	if resultTyp.Kind() == reflect.Slice || resultTyp.Kind() == reflect.Array {
		wrap = func(v reflect.Value) interface{} {
			return toList(v)
		}
	}
	return func(ctx context.Context, result ...reflect.Value) (interface{}, error) {
		return &chanAsyncValue{typeName, fieldName, result[0], wrap}, nil
	}
}

func buildAsyncErrorResultHandler(resultTyp reflect.Type, typeName, fieldName string) resultHandler {
	var wrap func(v reflect.Value) interface{}
	if resultTyp.Kind() == reflect.Slice || resultTyp.Kind() == reflect.Array {
		wrap = func(v reflect.Value) interface{} {
			return toList(v)
		}
	}
	return func(ctx context.Context, result ...reflect.Value) (interface{}, error) {
		return &chanErrorAsyncValue{typeName, fieldName, result[0], result[1], wrap}, nil
	}
}

func buildAsyncFuncResultHandler(asyncResultHandler resultHandler, typeName, fieldName string) resultHandler {
	return func(ctx context.Context, result ...reflect.Value) (interface{}, error) {
		return &funcAsyncValue{result[0], asyncResultHandler, ctx}, nil
	}
}
