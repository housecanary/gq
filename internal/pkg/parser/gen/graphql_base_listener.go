// Code generated from /Users/mpoindexter/dev/gq/grammar/Graphql.g4 by ANTLR 4.12.0. DO NOT EDIT.

package gen // Graphql
import "github.com/antlr/antlr4/runtime/Go/antlr/v4"

// BaseGraphqlListener is a complete listener for a parse tree produced by GraphqlParser.
type BaseGraphqlListener struct{}

var _ GraphqlListener = &BaseGraphqlListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseGraphqlListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseGraphqlListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseGraphqlListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseGraphqlListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterOperationType is called when production operationType is entered.
func (s *BaseGraphqlListener) EnterOperationType(ctx *OperationTypeContext) {}

// ExitOperationType is called when production operationType is exited.
func (s *BaseGraphqlListener) ExitOperationType(ctx *OperationTypeContext) {}

// EnterDescription is called when production description is entered.
func (s *BaseGraphqlListener) EnterDescription(ctx *DescriptionContext) {}

// ExitDescription is called when production description is exited.
func (s *BaseGraphqlListener) ExitDescription(ctx *DescriptionContext) {}

// EnterEnumValue is called when production enumValue is entered.
func (s *BaseGraphqlListener) EnterEnumValue(ctx *EnumValueContext) {}

// ExitEnumValue is called when production enumValue is exited.
func (s *BaseGraphqlListener) ExitEnumValue(ctx *EnumValueContext) {}

// EnterArrayValue is called when production arrayValue is entered.
func (s *BaseGraphqlListener) EnterArrayValue(ctx *ArrayValueContext) {}

// ExitArrayValue is called when production arrayValue is exited.
func (s *BaseGraphqlListener) ExitArrayValue(ctx *ArrayValueContext) {}

// EnterArrayValueWithVariable is called when production arrayValueWithVariable is entered.
func (s *BaseGraphqlListener) EnterArrayValueWithVariable(ctx *ArrayValueWithVariableContext) {}

// ExitArrayValueWithVariable is called when production arrayValueWithVariable is exited.
func (s *BaseGraphqlListener) ExitArrayValueWithVariable(ctx *ArrayValueWithVariableContext) {}

// EnterObjectValue is called when production objectValue is entered.
func (s *BaseGraphqlListener) EnterObjectValue(ctx *ObjectValueContext) {}

// ExitObjectValue is called when production objectValue is exited.
func (s *BaseGraphqlListener) ExitObjectValue(ctx *ObjectValueContext) {}

// EnterObjectValueWithVariable is called when production objectValueWithVariable is entered.
func (s *BaseGraphqlListener) EnterObjectValueWithVariable(ctx *ObjectValueWithVariableContext) {}

// ExitObjectValueWithVariable is called when production objectValueWithVariable is exited.
func (s *BaseGraphqlListener) ExitObjectValueWithVariable(ctx *ObjectValueWithVariableContext) {}

// EnterObjectField is called when production objectField is entered.
func (s *BaseGraphqlListener) EnterObjectField(ctx *ObjectFieldContext) {}

// ExitObjectField is called when production objectField is exited.
func (s *BaseGraphqlListener) ExitObjectField(ctx *ObjectFieldContext) {}

// EnterObjectFieldWithVariable is called when production objectFieldWithVariable is entered.
func (s *BaseGraphqlListener) EnterObjectFieldWithVariable(ctx *ObjectFieldWithVariableContext) {}

// ExitObjectFieldWithVariable is called when production objectFieldWithVariable is exited.
func (s *BaseGraphqlListener) ExitObjectFieldWithVariable(ctx *ObjectFieldWithVariableContext) {}

// EnterDirectives is called when production directives is entered.
func (s *BaseGraphqlListener) EnterDirectives(ctx *DirectivesContext) {}

// ExitDirectives is called when production directives is exited.
func (s *BaseGraphqlListener) ExitDirectives(ctx *DirectivesContext) {}

// EnterDirective is called when production directive is entered.
func (s *BaseGraphqlListener) EnterDirective(ctx *DirectiveContext) {}

// ExitDirective is called when production directive is exited.
func (s *BaseGraphqlListener) ExitDirective(ctx *DirectiveContext) {}

// EnterArguments is called when production arguments is entered.
func (s *BaseGraphqlListener) EnterArguments(ctx *ArgumentsContext) {}

// ExitArguments is called when production arguments is exited.
func (s *BaseGraphqlListener) ExitArguments(ctx *ArgumentsContext) {}

// EnterArgument is called when production argument is entered.
func (s *BaseGraphqlListener) EnterArgument(ctx *ArgumentContext) {}

// ExitArgument is called when production argument is exited.
func (s *BaseGraphqlListener) ExitArgument(ctx *ArgumentContext) {}

// EnterBaseName is called when production baseName is entered.
func (s *BaseGraphqlListener) EnterBaseName(ctx *BaseNameContext) {}

// ExitBaseName is called when production baseName is exited.
func (s *BaseGraphqlListener) ExitBaseName(ctx *BaseNameContext) {}

// EnterFragmentName is called when production fragmentName is entered.
func (s *BaseGraphqlListener) EnterFragmentName(ctx *FragmentNameContext) {}

// ExitFragmentName is called when production fragmentName is exited.
func (s *BaseGraphqlListener) ExitFragmentName(ctx *FragmentNameContext) {}

// EnterEnumValueName is called when production enumValueName is entered.
func (s *BaseGraphqlListener) EnterEnumValueName(ctx *EnumValueNameContext) {}

// ExitEnumValueName is called when production enumValueName is exited.
func (s *BaseGraphqlListener) ExitEnumValueName(ctx *EnumValueNameContext) {}

// EnterName is called when production name is entered.
func (s *BaseGraphqlListener) EnterName(ctx *NameContext) {}

// ExitName is called when production name is exited.
func (s *BaseGraphqlListener) ExitName(ctx *NameContext) {}

// EnterValue is called when production value is entered.
func (s *BaseGraphqlListener) EnterValue(ctx *ValueContext) {}

// ExitValue is called when production value is exited.
func (s *BaseGraphqlListener) ExitValue(ctx *ValueContext) {}

// EnterValueWithVariable is called when production valueWithVariable is entered.
func (s *BaseGraphqlListener) EnterValueWithVariable(ctx *ValueWithVariableContext) {}

// ExitValueWithVariable is called when production valueWithVariable is exited.
func (s *BaseGraphqlListener) ExitValueWithVariable(ctx *ValueWithVariableContext) {}

// EnterVariable is called when production variable is entered.
func (s *BaseGraphqlListener) EnterVariable(ctx *VariableContext) {}

// ExitVariable is called when production variable is exited.
func (s *BaseGraphqlListener) ExitVariable(ctx *VariableContext) {}

// EnterDefaultValue is called when production defaultValue is entered.
func (s *BaseGraphqlListener) EnterDefaultValue(ctx *DefaultValueContext) {}

// ExitDefaultValue is called when production defaultValue is exited.
func (s *BaseGraphqlListener) ExitDefaultValue(ctx *DefaultValueContext) {}

// EnterGqlType is called when production gqlType is entered.
func (s *BaseGraphqlListener) EnterGqlType(ctx *GqlTypeContext) {}

// ExitGqlType is called when production gqlType is exited.
func (s *BaseGraphqlListener) ExitGqlType(ctx *GqlTypeContext) {}

// EnterTypeName is called when production typeName is entered.
func (s *BaseGraphqlListener) EnterTypeName(ctx *TypeNameContext) {}

// ExitTypeName is called when production typeName is exited.
func (s *BaseGraphqlListener) ExitTypeName(ctx *TypeNameContext) {}

// EnterListType is called when production listType is entered.
func (s *BaseGraphqlListener) EnterListType(ctx *ListTypeContext) {}

// ExitListType is called when production listType is exited.
func (s *BaseGraphqlListener) ExitListType(ctx *ListTypeContext) {}

// EnterNonNullType is called when production nonNullType is entered.
func (s *BaseGraphqlListener) EnterNonNullType(ctx *NonNullTypeContext) {}

// ExitNonNullType is called when production nonNullType is exited.
func (s *BaseGraphqlListener) ExitNonNullType(ctx *NonNullTypeContext) {}

// EnterOperationDefinition is called when production operationDefinition is entered.
func (s *BaseGraphqlListener) EnterOperationDefinition(ctx *OperationDefinitionContext) {}

// ExitOperationDefinition is called when production operationDefinition is exited.
func (s *BaseGraphqlListener) ExitOperationDefinition(ctx *OperationDefinitionContext) {}

// EnterVariableDefinitions is called when production variableDefinitions is entered.
func (s *BaseGraphqlListener) EnterVariableDefinitions(ctx *VariableDefinitionsContext) {}

// ExitVariableDefinitions is called when production variableDefinitions is exited.
func (s *BaseGraphqlListener) ExitVariableDefinitions(ctx *VariableDefinitionsContext) {}

// EnterVariableDefinition is called when production variableDefinition is entered.
func (s *BaseGraphqlListener) EnterVariableDefinition(ctx *VariableDefinitionContext) {}

// ExitVariableDefinition is called when production variableDefinition is exited.
func (s *BaseGraphqlListener) ExitVariableDefinition(ctx *VariableDefinitionContext) {}

// EnterSelectionSet is called when production selectionSet is entered.
func (s *BaseGraphqlListener) EnterSelectionSet(ctx *SelectionSetContext) {}

// ExitSelectionSet is called when production selectionSet is exited.
func (s *BaseGraphqlListener) ExitSelectionSet(ctx *SelectionSetContext) {}

// EnterSelection is called when production selection is entered.
func (s *BaseGraphqlListener) EnterSelection(ctx *SelectionContext) {}

// ExitSelection is called when production selection is exited.
func (s *BaseGraphqlListener) ExitSelection(ctx *SelectionContext) {}

// EnterField is called when production field is entered.
func (s *BaseGraphqlListener) EnterField(ctx *FieldContext) {}

// ExitField is called when production field is exited.
func (s *BaseGraphqlListener) ExitField(ctx *FieldContext) {}

// EnterAlias is called when production alias is entered.
func (s *BaseGraphqlListener) EnterAlias(ctx *AliasContext) {}

// ExitAlias is called when production alias is exited.
func (s *BaseGraphqlListener) ExitAlias(ctx *AliasContext) {}

// EnterFragmentSpread is called when production fragmentSpread is entered.
func (s *BaseGraphqlListener) EnterFragmentSpread(ctx *FragmentSpreadContext) {}

// ExitFragmentSpread is called when production fragmentSpread is exited.
func (s *BaseGraphqlListener) ExitFragmentSpread(ctx *FragmentSpreadContext) {}

// EnterInlineFragment is called when production inlineFragment is entered.
func (s *BaseGraphqlListener) EnterInlineFragment(ctx *InlineFragmentContext) {}

// ExitInlineFragment is called when production inlineFragment is exited.
func (s *BaseGraphqlListener) ExitInlineFragment(ctx *InlineFragmentContext) {}

// EnterFragmentDefinition is called when production fragmentDefinition is entered.
func (s *BaseGraphqlListener) EnterFragmentDefinition(ctx *FragmentDefinitionContext) {}

// ExitFragmentDefinition is called when production fragmentDefinition is exited.
func (s *BaseGraphqlListener) ExitFragmentDefinition(ctx *FragmentDefinitionContext) {}

// EnterTypeCondition is called when production typeCondition is entered.
func (s *BaseGraphqlListener) EnterTypeCondition(ctx *TypeConditionContext) {}

// ExitTypeCondition is called when production typeCondition is exited.
func (s *BaseGraphqlListener) ExitTypeCondition(ctx *TypeConditionContext) {}

// EnterDocument is called when production document is entered.
func (s *BaseGraphqlListener) EnterDocument(ctx *DocumentContext) {}

// ExitDocument is called when production document is exited.
func (s *BaseGraphqlListener) ExitDocument(ctx *DocumentContext) {}

// EnterTypeSystemDefinition is called when production typeSystemDefinition is entered.
func (s *BaseGraphqlListener) EnterTypeSystemDefinition(ctx *TypeSystemDefinitionContext) {}

// ExitTypeSystemDefinition is called when production typeSystemDefinition is exited.
func (s *BaseGraphqlListener) ExitTypeSystemDefinition(ctx *TypeSystemDefinitionContext) {}

// EnterTypeSystemExtension is called when production typeSystemExtension is entered.
func (s *BaseGraphqlListener) EnterTypeSystemExtension(ctx *TypeSystemExtensionContext) {}

// ExitTypeSystemExtension is called when production typeSystemExtension is exited.
func (s *BaseGraphqlListener) ExitTypeSystemExtension(ctx *TypeSystemExtensionContext) {}

// EnterSchemaDefinition is called when production schemaDefinition is entered.
func (s *BaseGraphqlListener) EnterSchemaDefinition(ctx *SchemaDefinitionContext) {}

// ExitSchemaDefinition is called when production schemaDefinition is exited.
func (s *BaseGraphqlListener) ExitSchemaDefinition(ctx *SchemaDefinitionContext) {}

// EnterSchemaExtension is called when production schemaExtension is entered.
func (s *BaseGraphqlListener) EnterSchemaExtension(ctx *SchemaExtensionContext) {}

// ExitSchemaExtension is called when production schemaExtension is exited.
func (s *BaseGraphqlListener) ExitSchemaExtension(ctx *SchemaExtensionContext) {}

// EnterOperationTypeDefinition is called when production operationTypeDefinition is entered.
func (s *BaseGraphqlListener) EnterOperationTypeDefinition(ctx *OperationTypeDefinitionContext) {}

// ExitOperationTypeDefinition is called when production operationTypeDefinition is exited.
func (s *BaseGraphqlListener) ExitOperationTypeDefinition(ctx *OperationTypeDefinitionContext) {}

// EnterTypeDefinition is called when production typeDefinition is entered.
func (s *BaseGraphqlListener) EnterTypeDefinition(ctx *TypeDefinitionContext) {}

// ExitTypeDefinition is called when production typeDefinition is exited.
func (s *BaseGraphqlListener) ExitTypeDefinition(ctx *TypeDefinitionContext) {}

// EnterTypeExtension is called when production typeExtension is entered.
func (s *BaseGraphqlListener) EnterTypeExtension(ctx *TypeExtensionContext) {}

// ExitTypeExtension is called when production typeExtension is exited.
func (s *BaseGraphqlListener) ExitTypeExtension(ctx *TypeExtensionContext) {}

// EnterEmptyParentheses is called when production emptyParentheses is entered.
func (s *BaseGraphqlListener) EnterEmptyParentheses(ctx *EmptyParenthesesContext) {}

// ExitEmptyParentheses is called when production emptyParentheses is exited.
func (s *BaseGraphqlListener) ExitEmptyParentheses(ctx *EmptyParenthesesContext) {}

// EnterScalarTypeDefinition is called when production scalarTypeDefinition is entered.
func (s *BaseGraphqlListener) EnterScalarTypeDefinition(ctx *ScalarTypeDefinitionContext) {}

// ExitScalarTypeDefinition is called when production scalarTypeDefinition is exited.
func (s *BaseGraphqlListener) ExitScalarTypeDefinition(ctx *ScalarTypeDefinitionContext) {}

// EnterScalarTypeExtensionDefinition is called when production scalarTypeExtensionDefinition is entered.
func (s *BaseGraphqlListener) EnterScalarTypeExtensionDefinition(ctx *ScalarTypeExtensionDefinitionContext) {
}

// ExitScalarTypeExtensionDefinition is called when production scalarTypeExtensionDefinition is exited.
func (s *BaseGraphqlListener) ExitScalarTypeExtensionDefinition(ctx *ScalarTypeExtensionDefinitionContext) {
}

// EnterObjectTypeDefinition is called when production objectTypeDefinition is entered.
func (s *BaseGraphqlListener) EnterObjectTypeDefinition(ctx *ObjectTypeDefinitionContext) {}

// ExitObjectTypeDefinition is called when production objectTypeDefinition is exited.
func (s *BaseGraphqlListener) ExitObjectTypeDefinition(ctx *ObjectTypeDefinitionContext) {}

// EnterObjectTypeExtensionDefinition is called when production objectTypeExtensionDefinition is entered.
func (s *BaseGraphqlListener) EnterObjectTypeExtensionDefinition(ctx *ObjectTypeExtensionDefinitionContext) {
}

// ExitObjectTypeExtensionDefinition is called when production objectTypeExtensionDefinition is exited.
func (s *BaseGraphqlListener) ExitObjectTypeExtensionDefinition(ctx *ObjectTypeExtensionDefinitionContext) {
}

// EnterImplementsInterfaces is called when production implementsInterfaces is entered.
func (s *BaseGraphqlListener) EnterImplementsInterfaces(ctx *ImplementsInterfacesContext) {}

// ExitImplementsInterfaces is called when production implementsInterfaces is exited.
func (s *BaseGraphqlListener) ExitImplementsInterfaces(ctx *ImplementsInterfacesContext) {}

// EnterFieldsDefinition is called when production fieldsDefinition is entered.
func (s *BaseGraphqlListener) EnterFieldsDefinition(ctx *FieldsDefinitionContext) {}

// ExitFieldsDefinition is called when production fieldsDefinition is exited.
func (s *BaseGraphqlListener) ExitFieldsDefinition(ctx *FieldsDefinitionContext) {}

// EnterExtensionFieldsDefinition is called when production extensionFieldsDefinition is entered.
func (s *BaseGraphqlListener) EnterExtensionFieldsDefinition(ctx *ExtensionFieldsDefinitionContext) {}

// ExitExtensionFieldsDefinition is called when production extensionFieldsDefinition is exited.
func (s *BaseGraphqlListener) ExitExtensionFieldsDefinition(ctx *ExtensionFieldsDefinitionContext) {}

// EnterFieldDefinition is called when production fieldDefinition is entered.
func (s *BaseGraphqlListener) EnterFieldDefinition(ctx *FieldDefinitionContext) {}

// ExitFieldDefinition is called when production fieldDefinition is exited.
func (s *BaseGraphqlListener) ExitFieldDefinition(ctx *FieldDefinitionContext) {}

// EnterArgumentsDefinition is called when production argumentsDefinition is entered.
func (s *BaseGraphqlListener) EnterArgumentsDefinition(ctx *ArgumentsDefinitionContext) {}

// ExitArgumentsDefinition is called when production argumentsDefinition is exited.
func (s *BaseGraphqlListener) ExitArgumentsDefinition(ctx *ArgumentsDefinitionContext) {}

// EnterInputValueDefinition is called when production inputValueDefinition is entered.
func (s *BaseGraphqlListener) EnterInputValueDefinition(ctx *InputValueDefinitionContext) {}

// ExitInputValueDefinition is called when production inputValueDefinition is exited.
func (s *BaseGraphqlListener) ExitInputValueDefinition(ctx *InputValueDefinitionContext) {}

// EnterInterfaceTypeDefinition is called when production interfaceTypeDefinition is entered.
func (s *BaseGraphqlListener) EnterInterfaceTypeDefinition(ctx *InterfaceTypeDefinitionContext) {}

// ExitInterfaceTypeDefinition is called when production interfaceTypeDefinition is exited.
func (s *BaseGraphqlListener) ExitInterfaceTypeDefinition(ctx *InterfaceTypeDefinitionContext) {}

// EnterInterfaceTypeExtensionDefinition is called when production interfaceTypeExtensionDefinition is entered.
func (s *BaseGraphqlListener) EnterInterfaceTypeExtensionDefinition(ctx *InterfaceTypeExtensionDefinitionContext) {
}

// ExitInterfaceTypeExtensionDefinition is called when production interfaceTypeExtensionDefinition is exited.
func (s *BaseGraphqlListener) ExitInterfaceTypeExtensionDefinition(ctx *InterfaceTypeExtensionDefinitionContext) {
}

// EnterUnionTypeDefinition is called when production unionTypeDefinition is entered.
func (s *BaseGraphqlListener) EnterUnionTypeDefinition(ctx *UnionTypeDefinitionContext) {}

// ExitUnionTypeDefinition is called when production unionTypeDefinition is exited.
func (s *BaseGraphqlListener) ExitUnionTypeDefinition(ctx *UnionTypeDefinitionContext) {}

// EnterUnionTypeExtensionDefinition is called when production unionTypeExtensionDefinition is entered.
func (s *BaseGraphqlListener) EnterUnionTypeExtensionDefinition(ctx *UnionTypeExtensionDefinitionContext) {
}

// ExitUnionTypeExtensionDefinition is called when production unionTypeExtensionDefinition is exited.
func (s *BaseGraphqlListener) ExitUnionTypeExtensionDefinition(ctx *UnionTypeExtensionDefinitionContext) {
}

// EnterUnionMembership is called when production unionMembership is entered.
func (s *BaseGraphqlListener) EnterUnionMembership(ctx *UnionMembershipContext) {}

// ExitUnionMembership is called when production unionMembership is exited.
func (s *BaseGraphqlListener) ExitUnionMembership(ctx *UnionMembershipContext) {}

// EnterUnionMembers is called when production unionMembers is entered.
func (s *BaseGraphqlListener) EnterUnionMembers(ctx *UnionMembersContext) {}

// ExitUnionMembers is called when production unionMembers is exited.
func (s *BaseGraphqlListener) ExitUnionMembers(ctx *UnionMembersContext) {}

// EnterEnumTypeDefinition is called when production enumTypeDefinition is entered.
func (s *BaseGraphqlListener) EnterEnumTypeDefinition(ctx *EnumTypeDefinitionContext) {}

// ExitEnumTypeDefinition is called when production enumTypeDefinition is exited.
func (s *BaseGraphqlListener) ExitEnumTypeDefinition(ctx *EnumTypeDefinitionContext) {}

// EnterEnumTypeExtensionDefinition is called when production enumTypeExtensionDefinition is entered.
func (s *BaseGraphqlListener) EnterEnumTypeExtensionDefinition(ctx *EnumTypeExtensionDefinitionContext) {
}

// ExitEnumTypeExtensionDefinition is called when production enumTypeExtensionDefinition is exited.
func (s *BaseGraphqlListener) ExitEnumTypeExtensionDefinition(ctx *EnumTypeExtensionDefinitionContext) {
}

// EnterEnumValueDefinitions is called when production enumValueDefinitions is entered.
func (s *BaseGraphqlListener) EnterEnumValueDefinitions(ctx *EnumValueDefinitionsContext) {}

// ExitEnumValueDefinitions is called when production enumValueDefinitions is exited.
func (s *BaseGraphqlListener) ExitEnumValueDefinitions(ctx *EnumValueDefinitionsContext) {}

// EnterExtensionEnumValueDefinitions is called when production extensionEnumValueDefinitions is entered.
func (s *BaseGraphqlListener) EnterExtensionEnumValueDefinitions(ctx *ExtensionEnumValueDefinitionsContext) {
}

// ExitExtensionEnumValueDefinitions is called when production extensionEnumValueDefinitions is exited.
func (s *BaseGraphqlListener) ExitExtensionEnumValueDefinitions(ctx *ExtensionEnumValueDefinitionsContext) {
}

// EnterEnumValueDefinition is called when production enumValueDefinition is entered.
func (s *BaseGraphqlListener) EnterEnumValueDefinition(ctx *EnumValueDefinitionContext) {}

// ExitEnumValueDefinition is called when production enumValueDefinition is exited.
func (s *BaseGraphqlListener) ExitEnumValueDefinition(ctx *EnumValueDefinitionContext) {}

// EnterInputObjectTypeDefinition is called when production inputObjectTypeDefinition is entered.
func (s *BaseGraphqlListener) EnterInputObjectTypeDefinition(ctx *InputObjectTypeDefinitionContext) {}

// ExitInputObjectTypeDefinition is called when production inputObjectTypeDefinition is exited.
func (s *BaseGraphqlListener) ExitInputObjectTypeDefinition(ctx *InputObjectTypeDefinitionContext) {}

// EnterInputObjectTypeExtensionDefinition is called when production inputObjectTypeExtensionDefinition is entered.
func (s *BaseGraphqlListener) EnterInputObjectTypeExtensionDefinition(ctx *InputObjectTypeExtensionDefinitionContext) {
}

// ExitInputObjectTypeExtensionDefinition is called when production inputObjectTypeExtensionDefinition is exited.
func (s *BaseGraphqlListener) ExitInputObjectTypeExtensionDefinition(ctx *InputObjectTypeExtensionDefinitionContext) {
}

// EnterInputObjectValueDefinitions is called when production inputObjectValueDefinitions is entered.
func (s *BaseGraphqlListener) EnterInputObjectValueDefinitions(ctx *InputObjectValueDefinitionsContext) {
}

// ExitInputObjectValueDefinitions is called when production inputObjectValueDefinitions is exited.
func (s *BaseGraphqlListener) ExitInputObjectValueDefinitions(ctx *InputObjectValueDefinitionsContext) {
}

// EnterExtensionInputObjectValueDefinitions is called when production extensionInputObjectValueDefinitions is entered.
func (s *BaseGraphqlListener) EnterExtensionInputObjectValueDefinitions(ctx *ExtensionInputObjectValueDefinitionsContext) {
}

// ExitExtensionInputObjectValueDefinitions is called when production extensionInputObjectValueDefinitions is exited.
func (s *BaseGraphqlListener) ExitExtensionInputObjectValueDefinitions(ctx *ExtensionInputObjectValueDefinitionsContext) {
}

// EnterDirectiveDefinition is called when production directiveDefinition is entered.
func (s *BaseGraphqlListener) EnterDirectiveDefinition(ctx *DirectiveDefinitionContext) {}

// ExitDirectiveDefinition is called when production directiveDefinition is exited.
func (s *BaseGraphqlListener) ExitDirectiveDefinition(ctx *DirectiveDefinitionContext) {}

// EnterDirectiveLocation is called when production directiveLocation is entered.
func (s *BaseGraphqlListener) EnterDirectiveLocation(ctx *DirectiveLocationContext) {}

// ExitDirectiveLocation is called when production directiveLocation is exited.
func (s *BaseGraphqlListener) ExitDirectiveLocation(ctx *DirectiveLocationContext) {}

// EnterDirectiveLocations is called when production directiveLocations is entered.
func (s *BaseGraphqlListener) EnterDirectiveLocations(ctx *DirectiveLocationsContext) {}

// ExitDirectiveLocations is called when production directiveLocations is exited.
func (s *BaseGraphqlListener) ExitDirectiveLocations(ctx *DirectiveLocationsContext) {}

// EnterPartialFieldDefinition is called when production partialFieldDefinition is entered.
func (s *BaseGraphqlListener) EnterPartialFieldDefinition(ctx *PartialFieldDefinitionContext) {}

// ExitPartialFieldDefinition is called when production partialFieldDefinition is exited.
func (s *BaseGraphqlListener) ExitPartialFieldDefinition(ctx *PartialFieldDefinitionContext) {}

// EnterPartialObjectTypeDefinition is called when production partialObjectTypeDefinition is entered.
func (s *BaseGraphqlListener) EnterPartialObjectTypeDefinition(ctx *PartialObjectTypeDefinitionContext) {
}

// ExitPartialObjectTypeDefinition is called when production partialObjectTypeDefinition is exited.
func (s *BaseGraphqlListener) ExitPartialObjectTypeDefinition(ctx *PartialObjectTypeDefinitionContext) {
}

// EnterPartialInputObjectTypeDefinition is called when production partialInputObjectTypeDefinition is entered.
func (s *BaseGraphqlListener) EnterPartialInputObjectTypeDefinition(ctx *PartialInputObjectTypeDefinitionContext) {
}

// ExitPartialInputObjectTypeDefinition is called when production partialInputObjectTypeDefinition is exited.
func (s *BaseGraphqlListener) ExitPartialInputObjectTypeDefinition(ctx *PartialInputObjectTypeDefinitionContext) {
}

// EnterPartialInputValueDefinition is called when production partialInputValueDefinition is entered.
func (s *BaseGraphqlListener) EnterPartialInputValueDefinition(ctx *PartialInputValueDefinitionContext) {
}

// ExitPartialInputValueDefinition is called when production partialInputValueDefinition is exited.
func (s *BaseGraphqlListener) ExitPartialInputValueDefinition(ctx *PartialInputValueDefinitionContext) {
}

// EnterPartialEnumTypeDefinition is called when production partialEnumTypeDefinition is entered.
func (s *BaseGraphqlListener) EnterPartialEnumTypeDefinition(ctx *PartialEnumTypeDefinitionContext) {}

// ExitPartialEnumTypeDefinition is called when production partialEnumTypeDefinition is exited.
func (s *BaseGraphqlListener) ExitPartialEnumTypeDefinition(ctx *PartialEnumTypeDefinitionContext) {}

// EnterPartialInterfaceTypeDefinition is called when production partialInterfaceTypeDefinition is entered.
func (s *BaseGraphqlListener) EnterPartialInterfaceTypeDefinition(ctx *PartialInterfaceTypeDefinitionContext) {
}

// ExitPartialInterfaceTypeDefinition is called when production partialInterfaceTypeDefinition is exited.
func (s *BaseGraphqlListener) ExitPartialInterfaceTypeDefinition(ctx *PartialInterfaceTypeDefinitionContext) {
}

// EnterPartialUnionTypeDefinition is called when production partialUnionTypeDefinition is entered.
func (s *BaseGraphqlListener) EnterPartialUnionTypeDefinition(ctx *PartialUnionTypeDefinitionContext) {
}

// ExitPartialUnionTypeDefinition is called when production partialUnionTypeDefinition is exited.
func (s *BaseGraphqlListener) ExitPartialUnionTypeDefinition(ctx *PartialUnionTypeDefinitionContext) {
}

// EnterPartialScalarTypeDefinition is called when production partialScalarTypeDefinition is entered.
func (s *BaseGraphqlListener) EnterPartialScalarTypeDefinition(ctx *PartialScalarTypeDefinitionContext) {
}

// ExitPartialScalarTypeDefinition is called when production partialScalarTypeDefinition is exited.
func (s *BaseGraphqlListener) ExitPartialScalarTypeDefinition(ctx *PartialScalarTypeDefinitionContext) {
}

// EnterTsResolverFieldDefinition is called when production tsResolverFieldDefinition is entered.
func (s *BaseGraphqlListener) EnterTsResolverFieldDefinition(ctx *TsResolverFieldDefinitionContext) {}

// ExitTsResolverFieldDefinition is called when production tsResolverFieldDefinition is exited.
func (s *BaseGraphqlListener) ExitTsResolverFieldDefinition(ctx *TsResolverFieldDefinitionContext) {}
