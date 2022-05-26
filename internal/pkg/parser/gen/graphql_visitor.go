// Code generated from /Users/mpoindexter/dev/gq/grammar/Graphql.g4 by ANTLR 4.12.0. DO NOT EDIT.

package gen // Graphql
import "github.com/antlr/antlr4/runtime/Go/antlr/v4"

// A complete Visitor for a parse tree produced by GraphqlParser.
type GraphqlVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by GraphqlParser#operationType.
	VisitOperationType(ctx *OperationTypeContext) interface{}

	// Visit a parse tree produced by GraphqlParser#description.
	VisitDescription(ctx *DescriptionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#enumValue.
	VisitEnumValue(ctx *EnumValueContext) interface{}

	// Visit a parse tree produced by GraphqlParser#arrayValue.
	VisitArrayValue(ctx *ArrayValueContext) interface{}

	// Visit a parse tree produced by GraphqlParser#arrayValueWithVariable.
	VisitArrayValueWithVariable(ctx *ArrayValueWithVariableContext) interface{}

	// Visit a parse tree produced by GraphqlParser#objectValue.
	VisitObjectValue(ctx *ObjectValueContext) interface{}

	// Visit a parse tree produced by GraphqlParser#objectValueWithVariable.
	VisitObjectValueWithVariable(ctx *ObjectValueWithVariableContext) interface{}

	// Visit a parse tree produced by GraphqlParser#objectField.
	VisitObjectField(ctx *ObjectFieldContext) interface{}

	// Visit a parse tree produced by GraphqlParser#objectFieldWithVariable.
	VisitObjectFieldWithVariable(ctx *ObjectFieldWithVariableContext) interface{}

	// Visit a parse tree produced by GraphqlParser#directives.
	VisitDirectives(ctx *DirectivesContext) interface{}

	// Visit a parse tree produced by GraphqlParser#directive.
	VisitDirective(ctx *DirectiveContext) interface{}

	// Visit a parse tree produced by GraphqlParser#arguments.
	VisitArguments(ctx *ArgumentsContext) interface{}

	// Visit a parse tree produced by GraphqlParser#argument.
	VisitArgument(ctx *ArgumentContext) interface{}

	// Visit a parse tree produced by GraphqlParser#baseName.
	VisitBaseName(ctx *BaseNameContext) interface{}

	// Visit a parse tree produced by GraphqlParser#fragmentName.
	VisitFragmentName(ctx *FragmentNameContext) interface{}

	// Visit a parse tree produced by GraphqlParser#enumValueName.
	VisitEnumValueName(ctx *EnumValueNameContext) interface{}

	// Visit a parse tree produced by GraphqlParser#name.
	VisitName(ctx *NameContext) interface{}

	// Visit a parse tree produced by GraphqlParser#value.
	VisitValue(ctx *ValueContext) interface{}

	// Visit a parse tree produced by GraphqlParser#valueWithVariable.
	VisitValueWithVariable(ctx *ValueWithVariableContext) interface{}

	// Visit a parse tree produced by GraphqlParser#variable.
	VisitVariable(ctx *VariableContext) interface{}

	// Visit a parse tree produced by GraphqlParser#defaultValue.
	VisitDefaultValue(ctx *DefaultValueContext) interface{}

	// Visit a parse tree produced by GraphqlParser#gqlType.
	VisitGqlType(ctx *GqlTypeContext) interface{}

	// Visit a parse tree produced by GraphqlParser#typeName.
	VisitTypeName(ctx *TypeNameContext) interface{}

	// Visit a parse tree produced by GraphqlParser#listType.
	VisitListType(ctx *ListTypeContext) interface{}

	// Visit a parse tree produced by GraphqlParser#nonNullType.
	VisitNonNullType(ctx *NonNullTypeContext) interface{}

	// Visit a parse tree produced by GraphqlParser#operationDefinition.
	VisitOperationDefinition(ctx *OperationDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#variableDefinitions.
	VisitVariableDefinitions(ctx *VariableDefinitionsContext) interface{}

	// Visit a parse tree produced by GraphqlParser#variableDefinition.
	VisitVariableDefinition(ctx *VariableDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#selectionSet.
	VisitSelectionSet(ctx *SelectionSetContext) interface{}

	// Visit a parse tree produced by GraphqlParser#selection.
	VisitSelection(ctx *SelectionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#field.
	VisitField(ctx *FieldContext) interface{}

	// Visit a parse tree produced by GraphqlParser#alias.
	VisitAlias(ctx *AliasContext) interface{}

	// Visit a parse tree produced by GraphqlParser#fragmentSpread.
	VisitFragmentSpread(ctx *FragmentSpreadContext) interface{}

	// Visit a parse tree produced by GraphqlParser#inlineFragment.
	VisitInlineFragment(ctx *InlineFragmentContext) interface{}

	// Visit a parse tree produced by GraphqlParser#fragmentDefinition.
	VisitFragmentDefinition(ctx *FragmentDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#typeCondition.
	VisitTypeCondition(ctx *TypeConditionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#document.
	VisitDocument(ctx *DocumentContext) interface{}

	// Visit a parse tree produced by GraphqlParser#typeSystemDefinition.
	VisitTypeSystemDefinition(ctx *TypeSystemDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#typeSystemExtension.
	VisitTypeSystemExtension(ctx *TypeSystemExtensionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#schemaDefinition.
	VisitSchemaDefinition(ctx *SchemaDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#schemaExtension.
	VisitSchemaExtension(ctx *SchemaExtensionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#operationTypeDefinition.
	VisitOperationTypeDefinition(ctx *OperationTypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#typeDefinition.
	VisitTypeDefinition(ctx *TypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#typeExtension.
	VisitTypeExtension(ctx *TypeExtensionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#emptyParentheses.
	VisitEmptyParentheses(ctx *EmptyParenthesesContext) interface{}

	// Visit a parse tree produced by GraphqlParser#scalarTypeDefinition.
	VisitScalarTypeDefinition(ctx *ScalarTypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#scalarTypeExtensionDefinition.
	VisitScalarTypeExtensionDefinition(ctx *ScalarTypeExtensionDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#objectTypeDefinition.
	VisitObjectTypeDefinition(ctx *ObjectTypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#objectTypeExtensionDefinition.
	VisitObjectTypeExtensionDefinition(ctx *ObjectTypeExtensionDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#implementsInterfaces.
	VisitImplementsInterfaces(ctx *ImplementsInterfacesContext) interface{}

	// Visit a parse tree produced by GraphqlParser#fieldsDefinition.
	VisitFieldsDefinition(ctx *FieldsDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#extensionFieldsDefinition.
	VisitExtensionFieldsDefinition(ctx *ExtensionFieldsDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#fieldDefinition.
	VisitFieldDefinition(ctx *FieldDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#argumentsDefinition.
	VisitArgumentsDefinition(ctx *ArgumentsDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#inputValueDefinition.
	VisitInputValueDefinition(ctx *InputValueDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#interfaceTypeDefinition.
	VisitInterfaceTypeDefinition(ctx *InterfaceTypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#interfaceTypeExtensionDefinition.
	VisitInterfaceTypeExtensionDefinition(ctx *InterfaceTypeExtensionDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#unionTypeDefinition.
	VisitUnionTypeDefinition(ctx *UnionTypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#unionTypeExtensionDefinition.
	VisitUnionTypeExtensionDefinition(ctx *UnionTypeExtensionDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#unionMembership.
	VisitUnionMembership(ctx *UnionMembershipContext) interface{}

	// Visit a parse tree produced by GraphqlParser#unionMembers.
	VisitUnionMembers(ctx *UnionMembersContext) interface{}

	// Visit a parse tree produced by GraphqlParser#enumTypeDefinition.
	VisitEnumTypeDefinition(ctx *EnumTypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#enumTypeExtensionDefinition.
	VisitEnumTypeExtensionDefinition(ctx *EnumTypeExtensionDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#enumValueDefinitions.
	VisitEnumValueDefinitions(ctx *EnumValueDefinitionsContext) interface{}

	// Visit a parse tree produced by GraphqlParser#extensionEnumValueDefinitions.
	VisitExtensionEnumValueDefinitions(ctx *ExtensionEnumValueDefinitionsContext) interface{}

	// Visit a parse tree produced by GraphqlParser#enumValueDefinition.
	VisitEnumValueDefinition(ctx *EnumValueDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#inputObjectTypeDefinition.
	VisitInputObjectTypeDefinition(ctx *InputObjectTypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#inputObjectTypeExtensionDefinition.
	VisitInputObjectTypeExtensionDefinition(ctx *InputObjectTypeExtensionDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#inputObjectValueDefinitions.
	VisitInputObjectValueDefinitions(ctx *InputObjectValueDefinitionsContext) interface{}

	// Visit a parse tree produced by GraphqlParser#extensionInputObjectValueDefinitions.
	VisitExtensionInputObjectValueDefinitions(ctx *ExtensionInputObjectValueDefinitionsContext) interface{}

	// Visit a parse tree produced by GraphqlParser#directiveDefinition.
	VisitDirectiveDefinition(ctx *DirectiveDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#directiveLocation.
	VisitDirectiveLocation(ctx *DirectiveLocationContext) interface{}

	// Visit a parse tree produced by GraphqlParser#directiveLocations.
	VisitDirectiveLocations(ctx *DirectiveLocationsContext) interface{}

	// Visit a parse tree produced by GraphqlParser#partialFieldDefinition.
	VisitPartialFieldDefinition(ctx *PartialFieldDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#partialObjectTypeDefinition.
	VisitPartialObjectTypeDefinition(ctx *PartialObjectTypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#partialInputObjectTypeDefinition.
	VisitPartialInputObjectTypeDefinition(ctx *PartialInputObjectTypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#partialInputValueDefinition.
	VisitPartialInputValueDefinition(ctx *PartialInputValueDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#partialEnumTypeDefinition.
	VisitPartialEnumTypeDefinition(ctx *PartialEnumTypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#partialInterfaceTypeDefinition.
	VisitPartialInterfaceTypeDefinition(ctx *PartialInterfaceTypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#partialUnionTypeDefinition.
	VisitPartialUnionTypeDefinition(ctx *PartialUnionTypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#partialScalarTypeDefinition.
	VisitPartialScalarTypeDefinition(ctx *PartialScalarTypeDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#tsResolverFieldDefinition.
	VisitTsResolverFieldDefinition(ctx *TsResolverFieldDefinitionContext) interface{}

	// Visit a parse tree produced by GraphqlParser#tsTypeDefinition.
	VisitTsTypeDefinition(ctx *TsTypeDefinitionContext) interface{}
}
