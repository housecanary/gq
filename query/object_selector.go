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
	ArgValues     map[string]interface{}
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
	argValues := make(map[string]interface{})
	for _, arg := range astField.Arguments {
		lv, converted := concreteAstValueToLiteralValue(arg.Value)
		if converted {
			argValues[arg.Name] = lv
		} else {
			argValues[arg.Name] = arg.Value
		}
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

func notifyCb(v interface{}, err error, cb ResolveCompleteCallback) (interface{}, error) {
	if _, ok := v.(schema.AsyncValue); ok {
		return v, err
	}

	err = cb(v, err)

	return v, err
}

func safeResolve(ctx *exeContext, value interface{}, f *objectSelectorField, cb ResolveCompleteCallback) (fieldValue interface{}, err error) {
	prevField := ctx.currentField
	ctx.currentField = f
	defer func() { ctx.currentField = prevField }()
	resolver := f.Field.Resolver()

	if sr, ok := resolver.(schema.SafeResolver); ok {
		fieldValue, err = sr.ResolveSafe(ctx, value)
		if cb != nil {
			fieldValue, err = notifyCb(fieldValue, err, cb)
		}
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
			if cb != nil {
				fieldValue, err = notifyCb(fieldValue, err, cb)
			}

			if err != nil {
				ctx.listener.NotifyError(err)
			}
		}()
		fieldValue, err = resolver.Resolve(ctx, value)
	}
	return
}

func safeAsync(ctx *exeContext, async schema.AsyncValue, f *objectSelectorField, fieldCollector collector, cb ResolveCompleteCallback) contFunc {
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
		if cb != nil {
			value, err = notifyCb(value, err, cb)
		}

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

func (s *objectSelector) apply(ctx *exeContext, value interface{}, collector collector) contFunc {
	if value == nil {
		return nil
	}

	valueCollector := collector.Object(len(s.Fields))
	var deferred []contFunc
	for _, f := range s.Fields {
		currentField := f
		fieldCollector := valueCollector.Field(currentField.AstField.Alias)

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

		var cf contFunc
		if async, ok := value.(schema.AsyncValue); ok {
			cf = safeAsync(ctx, async, currentField, fieldCollector, cb)
		} else {
			cf = currentField.Sel.apply(ctx, value, fieldCollector)
		}

		if cf != nil {
			deferred = append(deferred, cf)
		}

		// FUTURE: If and when we support mutations, instead of adding contFunc
		// to the worklist, we should immediately NotifyIdle() and call contFunc
		// to ensure serial execution order
	}

	if len(deferred) > 0 {
		wl := worklist(deferred)
		return wl.Continue
	}

	return nil
}

func (c *exeContext) GetArgumentValue(name string) (interface{}, error) {
	argResolver, ok := c.currentField.ArgResolvers[name]
	if !ok {
		return nil, fmt.Errorf("Invalid argument %s", name)
	}

	val, ok := c.currentField.ArgValues[name]
	if !ok {
		rv, err := argResolver(c, c.currentField.DefaultValues[name])
		if err != nil {
			err = fmt.Errorf("Error in argument %s: %v", name, err)
		}
		return rv, err
	}

	var lv schema.LiteralValue
	if converted, ok := val.(schema.LiteralValue); ok {
		lv = converted
	} else {
		lv = c.astValueToLiteralValue(val.(ast.Value))
	}

	rv, err := argResolver(c, lv)
	if err != nil {
		err = fmt.Errorf("Error in argument %s: %v", name, err)
	}
	return rv, err
}

func (c *exeContext) GetRawArgumentValue(name string) (schema.LiteralValue, error) {
	_, ok := c.currentField.ArgResolvers[name]
	if !ok {
		return nil, fmt.Errorf("Invalid argument %s", name)
	}

	val, ok := c.currentField.ArgValues[name]
	if !ok {
		val = c.currentField.DefaultValues[name]
	}

	if val == nil {
		return nil, nil
	}

	var lv schema.LiteralValue
	if converted, ok := val.(schema.LiteralValue); ok {
		lv = converted
	} else {
		lv = c.astValueToLiteralValue(val.(ast.Value))
	}

	return lv, nil
}

func (c *exeContext) ChildFieldsIterator() schema.FieldSelectionIterator {
	return newSelectionIterator(c.currentField.Sel)
}

func (c *exeContext) WalkChildSelections(cb schema.FieldWalkCB) bool {
	return walkChildSelectionsAdapter{c.currentField.Sel}.WalkChildSelections(cb)
}

type walkChildSelectionsAdapter struct {
	sel selector
}

func (a walkChildSelectionsAdapter) WalkChildSelections(cb schema.FieldWalkCB) bool {
	itr := newSelectionIterator(a.sel)
	for itr.Next() {
		abort := cb(itr.Selection(), itr.SchemaField(), walkChildSelectionsAdapter{itr.current.Sel})
		if abort {
			return true
		}
	}
	return false
}

func (c *exeContext) astValueToLiteralValue(val ast.Value) schema.LiteralValue {
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
		m := make(schema.LiteralObject, len(v.V))
		for k, e := range v.V {
			m[k] = c.astValueToLiteralValue(e)
		}
		return m
	case ast.ReferenceValue:
		return c.variables[v.Name]
	}
	panic(fmt.Errorf("Unknown ast value %v", val))
}

func concreteAstValueToLiteralValue(val ast.Value) (schema.LiteralValue, bool) {
	switch v := val.(type) {
	case ast.StringValue:
		return schema.LiteralString(v.V), true
	case ast.IntValue:
		return schema.LiteralNumber(v.V), true
	case ast.FloatValue:
		return schema.LiteralNumber(v.V), true
	case ast.BooleanValue:
		return schema.LiteralBool(v.V), true
	case ast.NilValue:
		return nil, true
	case ast.EnumValue:
		return schema.LiteralString(v.V), true
	case ast.ArrayValue:
		ary := make(schema.LiteralArray, len(v.V))
		var concrete bool
		for i, e := range v.V {
			ary[i], concrete = concreteAstValueToLiteralValue(e)
			if !concrete {
				return nil, false
			}
		}
		return ary, true
	case ast.ObjectValue:
		m := make(schema.LiteralObject, len(v.V))
		var concrete bool
		for k, e := range v.V {
			m[k], concrete = concreteAstValueToLiteralValue(e)
			if !concrete {
				return nil, false
			}
		}
		return m, true
	}
	return nil, false
}
