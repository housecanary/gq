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

package parser

import (
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser/gen"
)

type BaseASTBuilderVisitor struct {
	gen.BaseGraphqlVisitor
}

func (v *BaseASTBuilderVisitor) VisitEnumValue(ctx *gen.EnumValueContext) interface{} {
	s := ctx.Name().Accept(v).(string)
	return ast.EnumValue{V: s}
}

func (v *BaseASTBuilderVisitor) VisitArrayValue(ctx *gen.ArrayValueContext) interface{} {
	children := ctx.AllValue()
	typedChildren := make([]antlr.ParserRuleContext, len(children))
	for i, c := range children {
		typedChildren[i] = c
	}
	return v.visitArrayValue(typedChildren)
}

func (v *BaseASTBuilderVisitor) VisitArrayValueWithVariable(ctx *gen.ArrayValueWithVariableContext) interface{} {
	children := ctx.AllValueWithVariable()
	typedChildren := make([]antlr.ParserRuleContext, len(children))
	for i, c := range children {
		typedChildren[i] = c
	}
	return v.visitArrayValue(typedChildren)
}

func (v *BaseASTBuilderVisitor) visitArrayValue(children []antlr.ParserRuleContext) ast.ArrayValue {
	values := make([]ast.Value, len(children))
	for i, c := range children {
		values[i] = c.Accept(v).(ast.Value)
	}
	return ast.ArrayValue{V: values}
}

func (v *BaseASTBuilderVisitor) VisitObjectValue(ctx *gen.ObjectValueContext) interface{} {
	children := ctx.AllObjectField()
	m := make(map[string]ast.Value)
	for _, c := range children {
		f := c.Accept(v).(*objectField)
		m[f.Name] = f.Value
	}
	return ast.ObjectValue{V: m}
}

func (v *BaseASTBuilderVisitor) VisitObjectValueWithVariable(ctx *gen.ObjectValueWithVariableContext) interface{} {
	children := ctx.AllObjectFieldWithVariable()
	m := make(map[string]ast.Value)
	for _, c := range children {
		f := c.Accept(v).(*objectField)
		m[f.Name] = f.Value
	}
	return ast.ObjectValue{V: m}
}

type objectField struct {
	Name  string
	Value ast.Value
}

func (v *BaseASTBuilderVisitor) VisitObjectField(ctx *gen.ObjectFieldContext) interface{} {
	var f objectField
	f.Name = ctx.Name().Accept(v).(string)
	f.Value = ctx.Value().Accept(v).(ast.Value)
	return &f
}

func (v *BaseASTBuilderVisitor) VisitObjectFieldWithVariable(ctx *gen.ObjectFieldWithVariableContext) interface{} {
	var f objectField
	f.Name = ctx.Name().Accept(v).(string)
	f.Value = ctx.ValueWithVariable().Accept(v).(ast.Value)
	return &f
}

func (v *BaseASTBuilderVisitor) VisitDirectives(ctx *gen.DirectivesContext) interface{} {
	var directives ast.Directives
	for _, c := range ctx.AllDirective() {
		directives = append(directives, c.Accept(v).(*ast.Directive))
	}
	return directives
}

func (v *BaseASTBuilderVisitor) VisitDirective(ctx *gen.DirectiveContext) interface{} {
	var directive ast.Directive
	directive.Name = ctx.Name().Accept(v).(string)
	if c := ctx.Arguments(); c != nil {
		directive.Arguments = c.Accept(v).(ast.Arguments)
	}
	return &directive
}

func (v *BaseASTBuilderVisitor) VisitArguments(ctx *gen.ArgumentsContext) interface{} {
	var arguments ast.Arguments
	for _, c := range ctx.AllArgument() {
		arguments = append(arguments, c.Accept(v).(*ast.Argument))
	}
	return arguments
}

func (v *BaseASTBuilderVisitor) VisitArgument(ctx *gen.ArgumentContext) interface{} {
	var argument ast.Argument
	argument.Name = ctx.Name().Accept(v).(string)
	argument.Value = ctx.ValueWithVariable().Accept(v).(ast.Value)
	return &argument
}

func (v *BaseASTBuilderVisitor) VisitName(ctx *gen.NameContext) interface{} {
	return ctx.GetText()
}

func (v *BaseASTBuilderVisitor) VisitValue(ctx *gen.ValueContext) interface{} {
	if c := ctx.StringValue(); c != nil {
		return c.Accept(v).(ast.StringValue)
	}

	if c := ctx.IntValue(); c != nil {
		repr := c.GetText()
		i, err := strconv.ParseInt(repr, 10, 64)
		if err != nil {
			panic(err)
		}
		return ast.IntValue{V: i}
	}

	if c := ctx.FloatValue(); c != nil {
		repr := c.GetText()
		i, err := strconv.ParseFloat(repr, 64)
		if err != nil {
			panic(err)
		}
		return ast.FloatValue{V: i}
	}

	if c := ctx.BooleanValue(); c != nil {
		repr := c.GetText()
		if repr == "true" {
			return ast.BooleanValue{V: true}
		} else if repr == "false" {
			return ast.BooleanValue{V: false}
		}
		panic("Invalid boolean value")
	}

	if c := ctx.NullValue(); c != nil {
		return ast.NilValue{}
	}

	if c := ctx.EnumValue(); c != nil {
		return c.Accept(v)
	}

	if c := ctx.ArrayValue(); c != nil {
		return c.Accept(v)
	}

	if c := ctx.ObjectValue(); c != nil {
		return c.Accept(v)
	}

	panic("Invalid value")
}

func (v *BaseASTBuilderVisitor) VisitValueWithVariable(ctx *gen.ValueWithVariableContext) interface{} {

	if c := ctx.Variable(); c != nil {
		return ast.ReferenceValue{Name: c.Accept(v).(string)}
	}

	if c := ctx.StringValue(); c != nil {
		return c.Accept(v).(ast.StringValue)
	}

	if c := ctx.IntValue(); c != nil {
		repr := c.GetText()
		i, err := strconv.ParseInt(repr, 10, 64)
		if err != nil {
			panic(err)
		}
		return ast.IntValue{V: i}
	}

	if c := ctx.FloatValue(); c != nil {
		repr := c.GetText()
		i, err := strconv.ParseFloat(repr, 64)
		if err != nil {
			panic(err)
		}
		return ast.FloatValue{V: i}
	}

	if c := ctx.BooleanValue(); c != nil {
		repr := c.GetText()
		if repr == "true" {
			return ast.BooleanValue{V: true}
		} else if repr == "false" {
			return ast.BooleanValue{V: false}
		}
		panic("Invalid boolean value")
	}

	if c := ctx.NullValue(); c != nil {
		return ast.NilValue{}
	}

	if c := ctx.EnumValue(); c != nil {
		return c.Accept(v)
	}

	if c := ctx.ArrayValueWithVariable(); c != nil {
		return c.Accept(v)
	}

	if c := ctx.ObjectValueWithVariable(); c != nil {
		return c.Accept(v)
	}

	panic("Invalid value")
}

func (v *BaseASTBuilderVisitor) VisitVariable(ctx *gen.VariableContext) interface{} {
	return ctx.Name().Accept(v).(string)
}

func (v *BaseASTBuilderVisitor) VisitDefaultValue(ctx *gen.DefaultValueContext) interface{} {
	return ctx.Value().Accept(v)
}

func (v *BaseASTBuilderVisitor) VisitStringValue(ctx *gen.StringValueContext) interface{} {
	if c := ctx.TripleQuotedStringValue(); c != nil {
		s := c.GetText()
		return ast.StringValue{V: s[3 : len(s)-3]}
	} else if c := ctx.StringValue(); c != nil {
		s := c.GetText()
		return ast.StringValue{V: s[1 : len(s)-1]}
	}
	panic("Invalid string value")
}

func (v *BaseASTBuilderVisitor) VisitGqlType(ctx *gen.GqlTypeContext) interface{} {
	if c := ctx.TypeName(); c != nil {
		s := c.Accept(v).(string)
		return &ast.SimpleType{Name: s}
	}

	if c := ctx.ListType(); c != nil {
		return c.Accept(v)
	}

	if c := ctx.NonNullType(); c != nil {
		return c.Accept(v)
	}

	panic("Invalid gql type")
}

func (v *BaseASTBuilderVisitor) VisitTypeName(ctx *gen.TypeNameContext) interface{} {
	return ctx.GetText()
}

func (v *BaseASTBuilderVisitor) VisitListType(ctx *gen.ListTypeContext) interface{} {
	return &ast.ListType{Of: ctx.GqlType().Accept(v).(ast.Type)}
}

func (v *BaseASTBuilderVisitor) VisitNonNullType(ctx *gen.NonNullTypeContext) interface{} {
	if c := ctx.TypeName(); c != nil {
		s := c.Accept(v).(string)
		return &ast.NotNilType{Of: &ast.SimpleType{Name: s}}
	}

	if c := ctx.ListType(); c != nil {
		return &ast.NotNilType{Of: c.Accept(v).(ast.Type)}
	}

	panic("Invalid parse tree:  No type arg to not nil")
}
