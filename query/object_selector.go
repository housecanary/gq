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

package query

import (
	"context"
	"fmt"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

// An objectSelector runs the resolver for each selected field, and then passes the
// resolved value to the proper selector for that field to further select fields from
// the resolved value.
type objectSelector struct {
	defaultSelector
	Fields []*objectSelectorField
}

type objectSelectorField struct {
	AstField      *ast.Field
	Sel           selector
	Field         *schema.FieldDescriptor
	ArgValues     map[string]ast.Value
	ArgResolvers  map[string]argumentResolver
	DefaultValues map[string]schema.LiteralValue
	Row           int
	Col           int
}

type argumentResolver func(context.Context, schema.LiteralValue) (interface{}, error)

func buildObjectSelector(cc *compileContext, typ *schema.ObjectType, selections ast.SelectionSet) (selector, error) {
	os := objectSelector{defaultSelector: cc.newDefaultSelector()}
	for _, sel := range selections {
		switch v := sel.(type) {
		case *ast.FieldSelection:
			if err := os.addField(cc, typ, v.Field); err != nil {
				return nil, err
			}
		case *ast.FragmentSpreadSelection:
			fragDef := cc.LookupFragmentDefinition(v.FragmentName)
			if fragDef == nil {
				// TODO locate error
				return nil, fmt.Errorf("Unknown fragment %s", v.FragmentName)
			}

			fragFields, err := cc.expandFragment(fragDef.OnType, fragDef.SelectionSet, typ)
			if err != nil {
				return nil, err
			}
			for _, field := range fragFields {
				if err := os.addField(cc, typ, field); err != nil {
					return nil, err
				}
			}

		case *ast.InlineFragmentSelection:
			fragFields, err := cc.expandFragment(v.OnType, v.SelectionSet, typ)
			if err != nil {
				return nil, err
			}
			for _, field := range fragFields {
				if err := os.addField(cc, typ, field); err != nil {
					return nil, err
				}
			}
		}
	}
	return &os, nil
}

func (s *objectSelector) addField(cc *compileContext, typ *schema.ObjectType, astField ast.Field) error {
	schemaField := typ.Field(astField.Name)
	if schemaField == nil {
		// TODO located error
		return fmt.Errorf("Unknown field %s", astField.Name)
	}
	fieldType := schemaField.Type()
	childSelector, err := buildSelector(cc.withLocation(astField.Row, astField.Col), fieldType, astField.SelectionSet)
	if err != nil {
		return err
	}
	argValues := make(map[string]ast.Value)
	for _, arg := range astField.Arguments {
		argValues[arg.Name] = arg.Value
	}
	defaultValues := make(map[string]schema.LiteralValue)
	for _, arg := range schemaField.Arguments() {
		defaultValues[arg.Name()] = arg.DefaultValue()
	}
	argResolvers := make(map[string]argumentResolver)
	for _, arg := range schemaField.Arguments() {
		ar, err := cc.makeArgumentResolver(arg.Type().(schema.InputableType))
		if err != nil {
			return err
		}
		argResolvers[arg.Name()] = ar
	}
	s.Fields = append(s.Fields, &objectSelectorField{
		AstField:      &astField,
		Sel:           childSelector,
		Field:         schemaField,
		ArgValues:     argValues,
		ArgResolvers:  argResolvers,
		DefaultValues: defaultValues,
		Row:           astField.Row,
		Col:           astField.Col,
	})

	return nil
}

func maybeNotifyCb(v interface{}, err error, cb ResolveCompleteCallback) (interface{}, error) {
	if cb == nil {
		return v, err
	}

	if _, ok := v.(schema.AsyncValue); ok {
		return v, err
	}

	err = cb(v, err)

	return v, err
}

func safeResolve(ctx exeContext, value interface{}, f *objectSelectorField, cb ResolveCompleteCallback) (fieldValue interface{}, err error) {
	resolver := f.Field.Resolver()
	var resolverContext context.Context = ctx
	if resolver.NeedsFullContext() {
		resolverContext = &resolverContextImpl{ctx, fieldWalker{f.Sel, f.AstField}, f}
	}

	if sr, ok := resolver.(schema.SafeResolver); ok {
		fieldValue, err = sr.ResolveSafe(resolverContext, value)
		fieldValue, err = maybeNotifyCb(fieldValue, err, cb)
		if err != nil {
			ctx.listener.NotifyError(err)
		}
	} else {
		defer func() {
			if r := recover(); r != nil {
				fieldValue = nil
				if re, ok := r.(error); ok {
					err = re
				} else {
					err = fmt.Errorf("%v", r)
				}
			}

			fieldValue, err = maybeNotifyCb(fieldValue, err, cb)

			if err != nil {
				ctx.listener.NotifyError(err)
			}
		}()
		fieldValue, err = resolver.Resolve(resolverContext, value)
	}
	return
}

func safeAsync(ctx exeContext, async schema.AsyncValue, f *objectSelectorField, fieldCollector collector, cb ResolveCompleteCallback) contFunc {
	return func() (ret contFunc) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				if re, ok := r.(error); ok {
					err = re
				} else {
					err = fmt.Errorf("%v", r)
				}

				if cb != nil {
					err = cb(nil, err)
				}

				if err != nil {
					ctx.listener.NotifyError(err)
					fieldCollector.Error(err, f.Row, f.Col)
				}

				ret = nil
			}
		}()

		value, err := async.Await(ctx)
		value, err = maybeNotifyCb(value, err, cb)

		if err != nil {
			ctx.listener.NotifyError(err)
			fieldCollector.Error(err, f.Row, f.Col)
			return nil
		}

		// The returned value was still async.  Schedule it to run again
		if async, ok := value.(schema.AsyncValue); ok {
			return safeAsync(ctx, async, f, fieldCollector, cb)
		}
		ret = f.Sel.apply(ctx, value, fieldCollector)
		return
	}
}

func (s *objectSelector) apply(ctx exeContext, value interface{}, collector collector) contFunc {
	if value == nil {
		return nil
	}

	valueCollector := collector.Object(len(s.Fields))
	var deferred worklist
	for _, f := range s.Fields {
		currentField := f
		fieldCollector := valueCollector.Field(currentField.AstField.Alias)
		currentField.Sel.prepareCollector(fieldCollector)

		cb, err := ctx.listener.NotifyResolve(currentField.AstField, currentField.Field)
		if err != nil {
			fieldCollector.Error(err, currentField.Row, currentField.Col)
			continue
		}

		value, err := safeResolve(ctx, value, currentField, cb)
		if err != nil {
			fieldCollector.Error(err, currentField.Row, currentField.Col)
			continue
		}
		if async, ok := value.(schema.AsyncValue); ok {
			deferred.Add(safeAsync(ctx, async, currentField, fieldCollector, cb))
		} else {
			deferred.Add(currentField.Sel.apply(ctx, value, fieldCollector))
		}

		// FUTURE: If and when we support mutations, instead of adding contFunc
		// to the worklist, we should immediately NotifyIdle() and call contFunc
		// to ensure serial execution order
	}

	if len(deferred) > 0 {
		return deferred.Continue
	}

	return nil
}

type resolverContextImpl struct {
	exeContext
	fieldWalker
	f *objectSelectorField
}

type fieldWalker struct {
	sel   selector
	field *ast.Field
}

func (c *resolverContextImpl) GetArgumentValue(name string) (interface{}, error) {
	argResolver, ok := c.f.ArgResolvers[name]
	if !ok {
		return nil, fmt.Errorf("Invalid argument %s", name)
	}

	val, ok := c.f.ArgValues[name]
	if !ok {
		rv, err := argResolver(c, c.f.DefaultValues[name])
		if err != nil {
			err = fmt.Errorf("Error in argument %s: %v", name, err)
		}
		return rv, err
	}

	rv, err := argResolver(c, c.astValueToLiteralValue(val))
	if err != nil {
		err = fmt.Errorf("Error in argument %s: %v", name, err)
	}
	return rv, err
}

func (f fieldWalker) WalkChildSelections(cb schema.FieldWalkCB) bool {
	walkObjectSelections := func(o *objectSelector) bool {
		for _, c := range o.Fields {
			abort := cb(c.AstField, c.Field, fieldWalker{c.Sel, c.AstField})
			if abort {
				return true
			}
		}
		return false
	}
	sel := f.sel
	for {
		switch t := sel.(type) {
		case listSelector:
			sel = t.ElementSelector
			continue
		case notNilSelector:
			sel = t.Delegate
			continue
		}
		break
	}

	switch t := sel.(type) {
	case *objectSelector:
		return walkObjectSelections(t)
	case interfaceSelector:
		for _, e := range t.Elements {
			abort := walkObjectSelections(e.(*objectSelector))
			if abort {
				return true
			}
		}
	case unionSelector:
		for _, e := range t.Elements {
			abort := walkObjectSelections(e.(*objectSelector))
			if abort {
				return true
			}
		}
	}
	return false
}

func (c *resolverContextImpl) astValueToLiteralValue(val ast.Value) schema.LiteralValue {
	switch v := val.(type) {
	case ast.StringValue:
		return schema.LiteralString(v.V)
	case ast.IntValue:
		return schema.LiteralNumber(v.V)
	case ast.FloatValue:
		return schema.LiteralNumber(v.V)
	case ast.BooleanValue:
		return schema.LiteralBool(v.V)
	case ast.NilValue:
		return nil
	case ast.EnumValue:
		return schema.LiteralString(v.V)
	case ast.ArrayValue:
		ary := make(schema.LiteralArray, len(v.V))
		for i, e := range v.V {
			ary[i] = c.astValueToLiteralValue(e)
		}
		return ary
	case ast.ObjectValue:
		m := make(schema.LiteralObject)
		for k, e := range v.V {
			m[k] = c.astValueToLiteralValue(e)
		}
		return m
	case ast.ReferenceValue:
		return c.variables[v.Name]
	}
	panic(fmt.Errorf("Unknown ast value %v", val))
}
