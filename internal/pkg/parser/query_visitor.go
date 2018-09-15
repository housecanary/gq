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
	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser/gen"
)

type queryDocumentVisitor struct {
	BaseASTBuilderVisitor
}

func (v *queryDocumentVisitor) VisitOperationType(ctx *gen.OperationTypeContext) interface{} {
	return ast.OperationType(ctx.GetText())
}

func (v *queryDocumentVisitor) VisitOperationDefinition(ctx *gen.OperationDefinitionContext) interface{} {
	var op ast.OperationDefinition

	if c := ctx.OperationType(); c != nil {
		op.OperationType = c.Accept(v).(ast.OperationType)
	} else {
		op.OperationType = ast.OperationTypeQuery
	}

	if c := ctx.Name(); c != nil {
		op.Name = c.Accept(v).(string)
	}

	if c := ctx.VariableDefinitions(); c != nil {
		op.VariableDefinitions = c.Accept(v).(ast.VariableDefinitions)
	}

	if c := ctx.Directives(); c != nil {
		op.Directives = c.Accept(v).(ast.Directives)
	}

	op.SelectionSet = ctx.SelectionSet().Accept(v).(ast.SelectionSet)

	return &op
}

func (v *queryDocumentVisitor) VisitVariableDefinitions(ctx *gen.VariableDefinitionsContext) interface{} {
	var defs ast.VariableDefinitions
	for _, c := range ctx.AllVariableDefinition() {
		defs = append(defs, c.Accept(v).(*ast.VariableDefinition))
	}
	return defs
}

func (v *queryDocumentVisitor) VisitVariableDefinition(ctx *gen.VariableDefinitionContext) interface{} {
	var def ast.VariableDefinition
	def.VariableName = ctx.Variable().Accept(v).(string)
	def.Type = ctx.GqlType().Accept(v).(ast.Type)
	if c := ctx.DefaultValue(); c != nil {
		def.DefaultValue = c.Accept(v).(ast.Value)
	}
	return &def
}

func (v *queryDocumentVisitor) VisitSelectionSet(ctx *gen.SelectionSetContext) interface{} {
	children := ctx.AllSelection()
	selections := make([]ast.Selection, len(children))
	for i, c := range children {
		selections[i] = c.Accept(v).(ast.Selection)
	}
	return ast.SelectionSet(selections)
}

func (v *queryDocumentVisitor) VisitSelection(ctx *gen.SelectionContext) interface{} {
	if c := ctx.Field(); c != nil {
		return c.Accept(v)
	}
	if c := ctx.FragmentSpread(); c != nil {
		return c.Accept(v)
	}
	if c := ctx.InlineFragment(); c != nil {
		return c.Accept(v)
	}

	panic("Invalid selection")
}

func (v *queryDocumentVisitor) VisitField(ctx *gen.FieldContext) interface{} {
	var f ast.Field
	f.Name = ctx.Name().Accept(v).(string)
	if c := ctx.Alias(); c != nil {
		f.Alias = c.Accept(v).(string)
	} else {
		f.Alias = f.Name
	}

	if c := ctx.Arguments(); c != nil {
		f.Arguments = c.Accept(v).(ast.Arguments)
	}
	if c := ctx.Directives(); c != nil {
		f.Directives = c.Accept(v).(ast.Directives)
	}
	if c := ctx.SelectionSet(); c != nil {
		f.SelectionSet = c.Accept(v).(ast.SelectionSet)
	}
	f.Row = ctx.GetStart().GetLine() + 1
	f.Col = ctx.GetStart().GetColumn() + 1
	return &ast.FieldSelection{Field: f}
}

func (v *queryDocumentVisitor) VisitAlias(ctx *gen.AliasContext) interface{} {
	return ctx.Name().Accept(v)
}

func (v *queryDocumentVisitor) VisitFragmentSpread(ctx *gen.FragmentSpreadContext) interface{} {
	var f ast.FragmentSpreadSelection
	f.FragmentName = ctx.FragmentName().Accept(v).(string)
	if c := ctx.Directives(); c != nil {
		f.Directives = c.Accept(v).(ast.Directives)
	}
	return &f
}

func (v *queryDocumentVisitor) VisitInlineFragment(ctx *gen.InlineFragmentContext) interface{} {
	var f ast.InlineFragmentSelection
	if c := ctx.TypeCondition(); c != nil {
		f.OnType = c.Accept(v).(string)
	}
	if c := ctx.Directives(); c != nil {
		f.Directives = c.Accept(v).(ast.Directives)
	}
	f.SelectionSet = ctx.SelectionSet().Accept(v).(ast.SelectionSet)
	return &f
}

func (v *queryDocumentVisitor) VisitFragmentDefinition(ctx *gen.FragmentDefinitionContext) interface{} {
	var f ast.FragmentDefinition
	f.Name = ctx.FragmentName().Accept(v).(string)
	f.OnType = ctx.TypeCondition().Accept(v).(string)
	if c := ctx.Directives(); c != nil {
		f.Directives = c.Accept(v).(ast.Directives)
	}
	f.SelectionSet = ctx.SelectionSet().Accept(v).(ast.SelectionSet)
	return &f
}

func (v *queryDocumentVisitor) VisitFragmentName(ctx *gen.FragmentNameContext) interface{} {
	return ctx.Name().Accept(v)
}

func (v *queryDocumentVisitor) VisitTypeCondition(ctx *gen.TypeConditionContext) interface{} {
	return ctx.TypeName().Accept(v)
}

func (v *queryDocumentVisitor) VisitDocument(ctx *gen.DocumentContext) interface{} {
	var doc ast.Document
	for _, opCtx := range ctx.AllOperationDefinition() {
		op := opCtx.Accept(v).(*ast.OperationDefinition)
		doc.AddOperationDefinition(op)
	}

	for _, fragCtx := range ctx.AllFragmentDefinition() {
		frag := fragCtx.Accept(v).(*ast.FragmentDefinition)
		doc.AddFragmentDefinition(frag)
	}

	return &doc
}

var _ gen.GraphqlVisitor = (*queryDocumentVisitor)(nil)
