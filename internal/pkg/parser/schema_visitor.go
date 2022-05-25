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

type schemaVisitor struct {
	BaseASTBuilderVisitor
}

// Currently type system extensions are unsupported so do not visit any
func (v *schemaVisitor) VisitTypeSystemExtension(ctx *gen.TypeSystemExtensionContext) interface{} {
	return nil
}

// Currently directive definitions are not processed
func (v *schemaVisitor) VisitDirectiveDefinition(ctx *gen.DirectiveDefinitionContext) interface{} {
	return nil
}

func (v *schemaVisitor) VisitDescription(ctx *gen.DescriptionContext) interface{} {
	return parseString(ctx.StringValue().GetText())
}

func (v *schemaVisitor) VisitScalarTypeDefinition(ctx *gen.ScalarTypeDefinitionContext) interface{} {
	var d ast.ScalarTypeDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}
	d.Name = ctx.Name().Accept(v).(string)

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}

	return &d
}

func (v *schemaVisitor) VisitObjectTypeDefinition(ctx *gen.ObjectTypeDefinitionContext) interface{} {
	var d ast.ObjectTypeDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}
	d.Name = ctx.Name().Accept(v).(string)

	if c := ctx.ImplementsInterfaces(); c != nil {
		d.ImplementsInterfaces = c.Accept(v).(ast.ImplementsInterfaces)
	}

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}

	if c := ctx.FieldsDefinition(); c != nil {
		d.FieldsDefinition = c.Accept(v).(ast.FieldsDefinition)
	}

	return &d
}

func (v *schemaVisitor) VisitImplementsInterfaces(ctx *gen.ImplementsInterfacesContext) interface{} {
	var d ast.ImplementsInterfaces
	for _, c := range ctx.AllTypeName() {
		d = append(d, c.Accept(v).(string))
	}

	if c := ctx.ImplementsInterfaces(); c != nil {
		d = append(d, c.Accept(v).(ast.ImplementsInterfaces)...)
	}
	return d
}

func (v *schemaVisitor) VisitFieldsDefinition(ctx *gen.FieldsDefinitionContext) interface{} {
	var d ast.FieldsDefinition
	for _, c := range ctx.AllFieldDefinition() {
		d = append(d, c.Accept(v).(*ast.FieldDefinition))
	}
	return d
}

func (v *schemaVisitor) VisitFieldDefinition(ctx *gen.FieldDefinitionContext) interface{} {
	var d ast.FieldDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}
	d.Name = ctx.Name().Accept(v).(string)

	if c := ctx.ArgumentsDefinition(); c != nil {
		d.ArgumentsDefinition = c.Accept(v).(ast.ArgumentsDefinition)
	}

	d.Type = ctx.GqlType().Accept(v).(ast.Type)

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}
	return &d
}

func (v *schemaVisitor) VisitArgumentsDefinition(ctx *gen.ArgumentsDefinitionContext) interface{} {
	var d ast.ArgumentsDefinition
	for _, c := range ctx.AllInputValueDefinition() {
		d = append(d, c.Accept(v).(*ast.InputValueDefinition))
	}
	return d
}

func (v *schemaVisitor) VisitInputValueDefinition(ctx *gen.InputValueDefinitionContext) interface{} {
	var d ast.InputValueDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}
	d.Name = ctx.Name().Accept(v).(string)
	d.Type = ctx.GqlType().Accept(v).(ast.Type)
	if c := ctx.DefaultValue(); c != nil {
		d.DefaultValue = c.Accept(v).(ast.Value)
	}
	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}
	return &d
}

func (v *schemaVisitor) VisitInterfaceTypeDefinition(ctx *gen.InterfaceTypeDefinitionContext) interface{} {
	var d ast.InterfaceTypeDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}
	d.Name = ctx.Name().Accept(v).(string)

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}

	if c := ctx.FieldsDefinition(); c != nil {
		d.FieldsDefinition = c.Accept(v).(ast.FieldsDefinition)
	}

	return &d
}

func (v *schemaVisitor) VisitUnionTypeDefinition(ctx *gen.UnionTypeDefinitionContext) interface{} {
	var d ast.UnionTypeDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}
	d.Name = ctx.Name().Accept(v).(string)

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}

	d.UnionMembership = ctx.UnionMembership().Accept(v).(ast.UnionMembership)
	return &d
}

func (v *schemaVisitor) VisitUnionMembership(ctx *gen.UnionMembershipContext) interface{} {
	return ctx.UnionMembers().Accept(v)
}

func (v *schemaVisitor) VisitUnionMembers(ctx *gen.UnionMembersContext) interface{} {
	var d ast.UnionMembership

	if c := ctx.UnionMembers(); c != nil {
		d = append(d, c.Accept(v).(ast.UnionMembership)...)
	}
	d = append(d, ctx.TypeName().Accept(v).(string))
	return d
}

func (v *schemaVisitor) VisitEnumTypeDefinition(ctx *gen.EnumTypeDefinitionContext) interface{} {
	var d ast.EnumTypeDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}
	d.Name = ctx.Name().Accept(v).(string)

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}

	if c := ctx.EnumValueDefinitions(); c != nil {
		d.EnumValueDefinitions = c.Accept(v).(ast.EnumValueDefinitions)
	}

	return &d
}

func (v *schemaVisitor) VisitEnumValueDefinitions(ctx *gen.EnumValueDefinitionsContext) interface{} {
	var d ast.EnumValueDefinitions
	for _, c := range ctx.AllEnumValueDefinition() {
		d = append(d, c.Accept(v).(*ast.EnumValueDefinition))
	}
	return d
}

func (v *schemaVisitor) VisitEnumValueDefinition(ctx *gen.EnumValueDefinitionContext) interface{} {
	var d ast.EnumValueDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}
	d.Value = ctx.EnumValue().Accept(v).(ast.EnumValue).V

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}
	return &d
}

func (v *schemaVisitor) VisitInputObjectTypeDefinition(ctx *gen.InputObjectTypeDefinitionContext) interface{} {
	var d ast.InputObjectTypeDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}
	d.Name = ctx.Name().Accept(v).(string)

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}

	if c := ctx.InputObjectValueDefinitions(); c != nil {
		d.InputObjectValueDefinitions = c.Accept(v).(ast.InputObjectValueDefinitions)
	}
	return &d
}

func (v *schemaVisitor) VisitInputObjectValueDefinitions(ctx *gen.InputObjectValueDefinitionsContext) interface{} {
	var d ast.InputObjectValueDefinitions
	for _, c := range ctx.AllInputValueDefinition() {
		d = append(d, c.Accept(v).(*ast.InputValueDefinition))
	}
	return d
}

func (v *schemaVisitor) VisitPartialFieldDefinition(ctx *gen.PartialFieldDefinitionContext) interface{} {
	var d ast.FieldDefinition
	if c := ctx.Name(); c != nil {
		d.Name = c.Accept(v).(string)
	}

	if c := ctx.ArgumentsDefinition(); c != nil {
		d.ArgumentsDefinition = c.Accept(v).(ast.ArgumentsDefinition)
	}

	if c := ctx.GqlType(); c != nil {
		d.Type = c.Accept(v).(ast.Type)
	}

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}
	return &d
}

func (v *schemaVisitor) VisitTsResolverFieldDefinition(ctx *gen.TsResolverFieldDefinitionContext) interface{} {
	var d ast.FieldDefinition

	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}

	d.Name = ctx.Name().Accept(v).(string)

	if c := ctx.GqlType(); c != nil {
		d.Type = c.Accept(v).(ast.Type)
	}

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}
	return &d
}

func (v *schemaVisitor) VisitPartialObjectTypeDefinition(ctx *gen.PartialObjectTypeDefinitionContext) interface{} {
	var d ast.ObjectTypeDefinition
	if c := ctx.Name(); c != nil {
		d.Name = c.Accept(v).(string)
	}

	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}

	if c := ctx.ImplementsInterfaces(); c != nil {
		d.ImplementsInterfaces = c.Accept(v).(ast.ImplementsInterfaces)
	}

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}

	if c := ctx.FieldsDefinition(); c != nil {
		d.FieldsDefinition = c.Accept(v).(ast.FieldsDefinition)
	}

	return &d
}

func (v *schemaVisitor) VisitPartialInputObjectTypeDefinition(ctx *gen.PartialInputObjectTypeDefinitionContext) interface{} {
	var d ast.InputObjectTypeDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}

	if c := ctx.Name(); c != nil {
		d.Name = c.Accept(v).(string)
	}

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}

	if c := ctx.InputObjectValueDefinitions(); c != nil {
		d.InputObjectValueDefinitions = c.Accept(v).(ast.InputObjectValueDefinitions)
	}
	return &d
}

func (v *schemaVisitor) VisitPartialInputValueDefinition(ctx *gen.PartialInputValueDefinitionContext) interface{} {
	var d ast.InputValueDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}
	if c := ctx.Name(); c != nil {
		d.Name = c.Accept(v).(string)
	}
	if c := ctx.GqlType(); c != nil {
		d.Type = c.Accept(v).(ast.Type)
	}
	if c := ctx.DefaultValue(); c != nil {
		d.DefaultValue = c.Accept(v).(ast.Value)
	}
	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}
	return &d
}

func (v *schemaVisitor) VisitPartialEnumTypeDefinition(ctx *gen.PartialEnumTypeDefinitionContext) interface{} {
	var d ast.EnumTypeDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}
	if c := ctx.Name(); c != nil {
		d.Name = c.Accept(v).(string)
	}

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}

	if c := ctx.EnumValueDefinitions(); c != nil {
		d.EnumValueDefinitions = c.Accept(v).(ast.EnumValueDefinitions)
	}

	return &d
}

func (v *schemaVisitor) VisitPartialInterfaceTypeDefinition(ctx *gen.PartialInterfaceTypeDefinitionContext) interface{} {
	var d ast.InterfaceTypeDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}

	if c := ctx.Name(); c != nil {
		d.Name = c.Accept(v).(string)
	}

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}

	if c := ctx.FieldsDefinition(); c != nil {
		d.FieldsDefinition = c.Accept(v).(ast.FieldsDefinition)
	}

	return &d
}

func (v *schemaVisitor) VisitPartialUnionTypeDefinition(ctx *gen.PartialUnionTypeDefinitionContext) interface{} {
	var d ast.UnionTypeDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}
	if c := ctx.Name(); c != nil {
		d.Name = c.Accept(v).(string)
	}

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}

	if c := ctx.UnionMembership(); c != nil {
		d.UnionMembership = c.Accept(v).(ast.UnionMembership)
	}
	return &d
}

func (v *schemaVisitor) VisitPartialScalarTypeDefinition(ctx *gen.PartialScalarTypeDefinitionContext) interface{} {
	var d ast.ScalarTypeDefinition
	if c := ctx.Description(); c != nil {
		d.Description = c.Accept(v).(string)
	}

	if c := ctx.Name(); c != nil {
		d.Name = c.Accept(v).(string)
	}

	if c := ctx.Directives(); c != nil {
		d.Directives = c.Accept(v).(ast.Directives)
	}

	return &d
}
