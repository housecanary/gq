// Code generated from grammar/Graphql.g4 by ANTLR 4.9.2. DO NOT EDIT.

package gen // Graphql
import "github.com/antlr/antlr4/runtime/Go/antlr"

type BaseGraphqlVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseGraphqlVisitor) VisitOperationType(ctx *OperationTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitEnumValue(ctx *EnumValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitArrayValue(ctx *ArrayValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitArrayValueWithVariable(ctx *ArrayValueWithVariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitObjectValue(ctx *ObjectValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitObjectValueWithVariable(ctx *ObjectValueWithVariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitObjectField(ctx *ObjectFieldContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitObjectFieldWithVariable(ctx *ObjectFieldWithVariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitDirectives(ctx *DirectivesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitDirective(ctx *DirectiveContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitArguments(ctx *ArgumentsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitArgument(ctx *ArgumentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitName(ctx *NameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitValue(ctx *ValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitValueWithVariable(ctx *ValueWithVariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitVariable(ctx *VariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitDefaultValue(ctx *DefaultValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitStringValue(ctx *StringValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitGqlType(ctx *GqlTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitTypeName(ctx *TypeNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitListType(ctx *ListTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitNonNullType(ctx *NonNullTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitOperationDefinition(ctx *OperationDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitVariableDefinitions(ctx *VariableDefinitionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitVariableDefinition(ctx *VariableDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitSelectionSet(ctx *SelectionSetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitSelection(ctx *SelectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitField(ctx *FieldContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitAlias(ctx *AliasContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitFragmentSpread(ctx *FragmentSpreadContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitInlineFragment(ctx *InlineFragmentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitFragmentDefinition(ctx *FragmentDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitFragmentName(ctx *FragmentNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitTypeCondition(ctx *TypeConditionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitDocument(ctx *DocumentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitDescription(ctx *DescriptionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitTypeSystemDefinition(ctx *TypeSystemDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitSchemaDefinition(ctx *SchemaDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitOperationTypeDefinition(ctx *OperationTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitTypeDefinition(ctx *TypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitTypeExtension(ctx *TypeExtensionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitScalarTypeDefinition(ctx *ScalarTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitScalarTypeExtensionDefinition(ctx *ScalarTypeExtensionDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitObjectTypeDefinition(ctx *ObjectTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitObjectTypeExtensionDefinition(ctx *ObjectTypeExtensionDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitImplementsInterfaces(ctx *ImplementsInterfacesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitFieldsDefinition(ctx *FieldsDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitFieldDefinition(ctx *FieldDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitArgumentsDefinition(ctx *ArgumentsDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitInputValueDefinition(ctx *InputValueDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitInterfaceTypeDefinition(ctx *InterfaceTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitInterfaceTypeExtensionDefinition(ctx *InterfaceTypeExtensionDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitUnionTypeDefinition(ctx *UnionTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitUnionTypeExtensionDefinition(ctx *UnionTypeExtensionDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitUnionMembership(ctx *UnionMembershipContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitUnionMembers(ctx *UnionMembersContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitEnumTypeDefinition(ctx *EnumTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitEnumTypeExtensionDefinition(ctx *EnumTypeExtensionDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitEnumValueDefinitions(ctx *EnumValueDefinitionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitEnumValueDefinition(ctx *EnumValueDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitInputObjectTypeDefinition(ctx *InputObjectTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitInputObjectTypeExtensionDefinition(ctx *InputObjectTypeExtensionDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitInputObjectValueDefinitions(ctx *InputObjectValueDefinitionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitDirectiveDefinition(ctx *DirectiveDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitDirectiveLocation(ctx *DirectiveLocationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitDirectiveLocations(ctx *DirectiveLocationsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitPartialFieldDefinition(ctx *PartialFieldDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitPartialObjectTypeDefinition(ctx *PartialObjectTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitPartialInputObjectTypeDefinition(ctx *PartialInputObjectTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitPartialInputValueDefinition(ctx *PartialInputValueDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitPartialEnumTypeDefinition(ctx *PartialEnumTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitPartialInterfaceTypeDefinition(ctx *PartialInterfaceTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitPartialUnionTypeDefinition(ctx *PartialUnionTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGraphqlVisitor) VisitPartialScalarTypeDefinition(ctx *PartialScalarTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}
