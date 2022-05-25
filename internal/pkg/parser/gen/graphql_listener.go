// Code generated from /Users/mpoindexter/dev/gq/grammar/Graphql.g4 by ANTLR 4.12.0. DO NOT EDIT.

package gen // Graphql
import "github.com/antlr/antlr4/runtime/Go/antlr/v4"

// GraphqlListener is a complete listener for a parse tree produced by GraphqlParser.
type GraphqlListener interface {
	antlr.ParseTreeListener

	// EnterOperationType is called when entering the operationType production.
	EnterOperationType(c *OperationTypeContext)

	// EnterDescription is called when entering the description production.
	EnterDescription(c *DescriptionContext)

	// EnterEnumValue is called when entering the enumValue production.
	EnterEnumValue(c *EnumValueContext)

	// EnterArrayValue is called when entering the arrayValue production.
	EnterArrayValue(c *ArrayValueContext)

	// EnterArrayValueWithVariable is called when entering the arrayValueWithVariable production.
	EnterArrayValueWithVariable(c *ArrayValueWithVariableContext)

	// EnterObjectValue is called when entering the objectValue production.
	EnterObjectValue(c *ObjectValueContext)

	// EnterObjectValueWithVariable is called when entering the objectValueWithVariable production.
	EnterObjectValueWithVariable(c *ObjectValueWithVariableContext)

	// EnterObjectField is called when entering the objectField production.
	EnterObjectField(c *ObjectFieldContext)

	// EnterObjectFieldWithVariable is called when entering the objectFieldWithVariable production.
	EnterObjectFieldWithVariable(c *ObjectFieldWithVariableContext)

	// EnterDirectives is called when entering the directives production.
	EnterDirectives(c *DirectivesContext)

	// EnterDirective is called when entering the directive production.
	EnterDirective(c *DirectiveContext)

	// EnterArguments is called when entering the arguments production.
	EnterArguments(c *ArgumentsContext)

	// EnterArgument is called when entering the argument production.
	EnterArgument(c *ArgumentContext)

	// EnterBaseName is called when entering the baseName production.
	EnterBaseName(c *BaseNameContext)

	// EnterFragmentName is called when entering the fragmentName production.
	EnterFragmentName(c *FragmentNameContext)

	// EnterEnumValueName is called when entering the enumValueName production.
	EnterEnumValueName(c *EnumValueNameContext)

	// EnterName is called when entering the name production.
	EnterName(c *NameContext)

	// EnterValue is called when entering the value production.
	EnterValue(c *ValueContext)

	// EnterValueWithVariable is called when entering the valueWithVariable production.
	EnterValueWithVariable(c *ValueWithVariableContext)

	// EnterVariable is called when entering the variable production.
	EnterVariable(c *VariableContext)

	// EnterDefaultValue is called when entering the defaultValue production.
	EnterDefaultValue(c *DefaultValueContext)

	// EnterGqlType is called when entering the gqlType production.
	EnterGqlType(c *GqlTypeContext)

	// EnterTypeName is called when entering the typeName production.
	EnterTypeName(c *TypeNameContext)

	// EnterListType is called when entering the listType production.
	EnterListType(c *ListTypeContext)

	// EnterNonNullType is called when entering the nonNullType production.
	EnterNonNullType(c *NonNullTypeContext)

	// EnterOperationDefinition is called when entering the operationDefinition production.
	EnterOperationDefinition(c *OperationDefinitionContext)

	// EnterVariableDefinitions is called when entering the variableDefinitions production.
	EnterVariableDefinitions(c *VariableDefinitionsContext)

	// EnterVariableDefinition is called when entering the variableDefinition production.
	EnterVariableDefinition(c *VariableDefinitionContext)

	// EnterSelectionSet is called when entering the selectionSet production.
	EnterSelectionSet(c *SelectionSetContext)

	// EnterSelection is called when entering the selection production.
	EnterSelection(c *SelectionContext)

	// EnterField is called when entering the field production.
	EnterField(c *FieldContext)

	// EnterAlias is called when entering the alias production.
	EnterAlias(c *AliasContext)

	// EnterFragmentSpread is called when entering the fragmentSpread production.
	EnterFragmentSpread(c *FragmentSpreadContext)

	// EnterInlineFragment is called when entering the inlineFragment production.
	EnterInlineFragment(c *InlineFragmentContext)

	// EnterFragmentDefinition is called when entering the fragmentDefinition production.
	EnterFragmentDefinition(c *FragmentDefinitionContext)

	// EnterTypeCondition is called when entering the typeCondition production.
	EnterTypeCondition(c *TypeConditionContext)

	// EnterDocument is called when entering the document production.
	EnterDocument(c *DocumentContext)

	// EnterTypeSystemDefinition is called when entering the typeSystemDefinition production.
	EnterTypeSystemDefinition(c *TypeSystemDefinitionContext)

	// EnterTypeSystemExtension is called when entering the typeSystemExtension production.
	EnterTypeSystemExtension(c *TypeSystemExtensionContext)

	// EnterSchemaDefinition is called when entering the schemaDefinition production.
	EnterSchemaDefinition(c *SchemaDefinitionContext)

	// EnterSchemaExtension is called when entering the schemaExtension production.
	EnterSchemaExtension(c *SchemaExtensionContext)

	// EnterOperationTypeDefinition is called when entering the operationTypeDefinition production.
	EnterOperationTypeDefinition(c *OperationTypeDefinitionContext)

	// EnterTypeDefinition is called when entering the typeDefinition production.
	EnterTypeDefinition(c *TypeDefinitionContext)

	// EnterTypeExtension is called when entering the typeExtension production.
	EnterTypeExtension(c *TypeExtensionContext)

	// EnterEmptyParentheses is called when entering the emptyParentheses production.
	EnterEmptyParentheses(c *EmptyParenthesesContext)

	// EnterScalarTypeDefinition is called when entering the scalarTypeDefinition production.
	EnterScalarTypeDefinition(c *ScalarTypeDefinitionContext)

	// EnterScalarTypeExtensionDefinition is called when entering the scalarTypeExtensionDefinition production.
	EnterScalarTypeExtensionDefinition(c *ScalarTypeExtensionDefinitionContext)

	// EnterObjectTypeDefinition is called when entering the objectTypeDefinition production.
	EnterObjectTypeDefinition(c *ObjectTypeDefinitionContext)

	// EnterObjectTypeExtensionDefinition is called when entering the objectTypeExtensionDefinition production.
	EnterObjectTypeExtensionDefinition(c *ObjectTypeExtensionDefinitionContext)

	// EnterImplementsInterfaces is called when entering the implementsInterfaces production.
	EnterImplementsInterfaces(c *ImplementsInterfacesContext)

	// EnterFieldsDefinition is called when entering the fieldsDefinition production.
	EnterFieldsDefinition(c *FieldsDefinitionContext)

	// EnterExtensionFieldsDefinition is called when entering the extensionFieldsDefinition production.
	EnterExtensionFieldsDefinition(c *ExtensionFieldsDefinitionContext)

	// EnterFieldDefinition is called when entering the fieldDefinition production.
	EnterFieldDefinition(c *FieldDefinitionContext)

	// EnterArgumentsDefinition is called when entering the argumentsDefinition production.
	EnterArgumentsDefinition(c *ArgumentsDefinitionContext)

	// EnterInputValueDefinition is called when entering the inputValueDefinition production.
	EnterInputValueDefinition(c *InputValueDefinitionContext)

	// EnterInterfaceTypeDefinition is called when entering the interfaceTypeDefinition production.
	EnterInterfaceTypeDefinition(c *InterfaceTypeDefinitionContext)

	// EnterInterfaceTypeExtensionDefinition is called when entering the interfaceTypeExtensionDefinition production.
	EnterInterfaceTypeExtensionDefinition(c *InterfaceTypeExtensionDefinitionContext)

	// EnterUnionTypeDefinition is called when entering the unionTypeDefinition production.
	EnterUnionTypeDefinition(c *UnionTypeDefinitionContext)

	// EnterUnionTypeExtensionDefinition is called when entering the unionTypeExtensionDefinition production.
	EnterUnionTypeExtensionDefinition(c *UnionTypeExtensionDefinitionContext)

	// EnterUnionMembership is called when entering the unionMembership production.
	EnterUnionMembership(c *UnionMembershipContext)

	// EnterUnionMembers is called when entering the unionMembers production.
	EnterUnionMembers(c *UnionMembersContext)

	// EnterEnumTypeDefinition is called when entering the enumTypeDefinition production.
	EnterEnumTypeDefinition(c *EnumTypeDefinitionContext)

	// EnterEnumTypeExtensionDefinition is called when entering the enumTypeExtensionDefinition production.
	EnterEnumTypeExtensionDefinition(c *EnumTypeExtensionDefinitionContext)

	// EnterEnumValueDefinitions is called when entering the enumValueDefinitions production.
	EnterEnumValueDefinitions(c *EnumValueDefinitionsContext)

	// EnterExtensionEnumValueDefinitions is called when entering the extensionEnumValueDefinitions production.
	EnterExtensionEnumValueDefinitions(c *ExtensionEnumValueDefinitionsContext)

	// EnterEnumValueDefinition is called when entering the enumValueDefinition production.
	EnterEnumValueDefinition(c *EnumValueDefinitionContext)

	// EnterInputObjectTypeDefinition is called when entering the inputObjectTypeDefinition production.
	EnterInputObjectTypeDefinition(c *InputObjectTypeDefinitionContext)

	// EnterInputObjectTypeExtensionDefinition is called when entering the inputObjectTypeExtensionDefinition production.
	EnterInputObjectTypeExtensionDefinition(c *InputObjectTypeExtensionDefinitionContext)

	// EnterInputObjectValueDefinitions is called when entering the inputObjectValueDefinitions production.
	EnterInputObjectValueDefinitions(c *InputObjectValueDefinitionsContext)

	// EnterExtensionInputObjectValueDefinitions is called when entering the extensionInputObjectValueDefinitions production.
	EnterExtensionInputObjectValueDefinitions(c *ExtensionInputObjectValueDefinitionsContext)

	// EnterDirectiveDefinition is called when entering the directiveDefinition production.
	EnterDirectiveDefinition(c *DirectiveDefinitionContext)

	// EnterDirectiveLocation is called when entering the directiveLocation production.
	EnterDirectiveLocation(c *DirectiveLocationContext)

	// EnterDirectiveLocations is called when entering the directiveLocations production.
	EnterDirectiveLocations(c *DirectiveLocationsContext)

	// EnterPartialFieldDefinition is called when entering the partialFieldDefinition production.
	EnterPartialFieldDefinition(c *PartialFieldDefinitionContext)

	// EnterPartialObjectTypeDefinition is called when entering the partialObjectTypeDefinition production.
	EnterPartialObjectTypeDefinition(c *PartialObjectTypeDefinitionContext)

	// EnterPartialInputObjectTypeDefinition is called when entering the partialInputObjectTypeDefinition production.
	EnterPartialInputObjectTypeDefinition(c *PartialInputObjectTypeDefinitionContext)

	// EnterPartialInputValueDefinition is called when entering the partialInputValueDefinition production.
	EnterPartialInputValueDefinition(c *PartialInputValueDefinitionContext)

	// EnterPartialEnumTypeDefinition is called when entering the partialEnumTypeDefinition production.
	EnterPartialEnumTypeDefinition(c *PartialEnumTypeDefinitionContext)

	// EnterPartialInterfaceTypeDefinition is called when entering the partialInterfaceTypeDefinition production.
	EnterPartialInterfaceTypeDefinition(c *PartialInterfaceTypeDefinitionContext)

	// EnterPartialUnionTypeDefinition is called when entering the partialUnionTypeDefinition production.
	EnterPartialUnionTypeDefinition(c *PartialUnionTypeDefinitionContext)

	// EnterPartialScalarTypeDefinition is called when entering the partialScalarTypeDefinition production.
	EnterPartialScalarTypeDefinition(c *PartialScalarTypeDefinitionContext)

	// EnterTsResolverFieldDefinition is called when entering the tsResolverFieldDefinition production.
	EnterTsResolverFieldDefinition(c *TsResolverFieldDefinitionContext)

	// ExitOperationType is called when exiting the operationType production.
	ExitOperationType(c *OperationTypeContext)

	// ExitDescription is called when exiting the description production.
	ExitDescription(c *DescriptionContext)

	// ExitEnumValue is called when exiting the enumValue production.
	ExitEnumValue(c *EnumValueContext)

	// ExitArrayValue is called when exiting the arrayValue production.
	ExitArrayValue(c *ArrayValueContext)

	// ExitArrayValueWithVariable is called when exiting the arrayValueWithVariable production.
	ExitArrayValueWithVariable(c *ArrayValueWithVariableContext)

	// ExitObjectValue is called when exiting the objectValue production.
	ExitObjectValue(c *ObjectValueContext)

	// ExitObjectValueWithVariable is called when exiting the objectValueWithVariable production.
	ExitObjectValueWithVariable(c *ObjectValueWithVariableContext)

	// ExitObjectField is called when exiting the objectField production.
	ExitObjectField(c *ObjectFieldContext)

	// ExitObjectFieldWithVariable is called when exiting the objectFieldWithVariable production.
	ExitObjectFieldWithVariable(c *ObjectFieldWithVariableContext)

	// ExitDirectives is called when exiting the directives production.
	ExitDirectives(c *DirectivesContext)

	// ExitDirective is called when exiting the directive production.
	ExitDirective(c *DirectiveContext)

	// ExitArguments is called when exiting the arguments production.
	ExitArguments(c *ArgumentsContext)

	// ExitArgument is called when exiting the argument production.
	ExitArgument(c *ArgumentContext)

	// ExitBaseName is called when exiting the baseName production.
	ExitBaseName(c *BaseNameContext)

	// ExitFragmentName is called when exiting the fragmentName production.
	ExitFragmentName(c *FragmentNameContext)

	// ExitEnumValueName is called when exiting the enumValueName production.
	ExitEnumValueName(c *EnumValueNameContext)

	// ExitName is called when exiting the name production.
	ExitName(c *NameContext)

	// ExitValue is called when exiting the value production.
	ExitValue(c *ValueContext)

	// ExitValueWithVariable is called when exiting the valueWithVariable production.
	ExitValueWithVariable(c *ValueWithVariableContext)

	// ExitVariable is called when exiting the variable production.
	ExitVariable(c *VariableContext)

	// ExitDefaultValue is called when exiting the defaultValue production.
	ExitDefaultValue(c *DefaultValueContext)

	// ExitGqlType is called when exiting the gqlType production.
	ExitGqlType(c *GqlTypeContext)

	// ExitTypeName is called when exiting the typeName production.
	ExitTypeName(c *TypeNameContext)

	// ExitListType is called when exiting the listType production.
	ExitListType(c *ListTypeContext)

	// ExitNonNullType is called when exiting the nonNullType production.
	ExitNonNullType(c *NonNullTypeContext)

	// ExitOperationDefinition is called when exiting the operationDefinition production.
	ExitOperationDefinition(c *OperationDefinitionContext)

	// ExitVariableDefinitions is called when exiting the variableDefinitions production.
	ExitVariableDefinitions(c *VariableDefinitionsContext)

	// ExitVariableDefinition is called when exiting the variableDefinition production.
	ExitVariableDefinition(c *VariableDefinitionContext)

	// ExitSelectionSet is called when exiting the selectionSet production.
	ExitSelectionSet(c *SelectionSetContext)

	// ExitSelection is called when exiting the selection production.
	ExitSelection(c *SelectionContext)

	// ExitField is called when exiting the field production.
	ExitField(c *FieldContext)

	// ExitAlias is called when exiting the alias production.
	ExitAlias(c *AliasContext)

	// ExitFragmentSpread is called when exiting the fragmentSpread production.
	ExitFragmentSpread(c *FragmentSpreadContext)

	// ExitInlineFragment is called when exiting the inlineFragment production.
	ExitInlineFragment(c *InlineFragmentContext)

	// ExitFragmentDefinition is called when exiting the fragmentDefinition production.
	ExitFragmentDefinition(c *FragmentDefinitionContext)

	// ExitTypeCondition is called when exiting the typeCondition production.
	ExitTypeCondition(c *TypeConditionContext)

	// ExitDocument is called when exiting the document production.
	ExitDocument(c *DocumentContext)

	// ExitTypeSystemDefinition is called when exiting the typeSystemDefinition production.
	ExitTypeSystemDefinition(c *TypeSystemDefinitionContext)

	// ExitTypeSystemExtension is called when exiting the typeSystemExtension production.
	ExitTypeSystemExtension(c *TypeSystemExtensionContext)

	// ExitSchemaDefinition is called when exiting the schemaDefinition production.
	ExitSchemaDefinition(c *SchemaDefinitionContext)

	// ExitSchemaExtension is called when exiting the schemaExtension production.
	ExitSchemaExtension(c *SchemaExtensionContext)

	// ExitOperationTypeDefinition is called when exiting the operationTypeDefinition production.
	ExitOperationTypeDefinition(c *OperationTypeDefinitionContext)

	// ExitTypeDefinition is called when exiting the typeDefinition production.
	ExitTypeDefinition(c *TypeDefinitionContext)

	// ExitTypeExtension is called when exiting the typeExtension production.
	ExitTypeExtension(c *TypeExtensionContext)

	// ExitEmptyParentheses is called when exiting the emptyParentheses production.
	ExitEmptyParentheses(c *EmptyParenthesesContext)

	// ExitScalarTypeDefinition is called when exiting the scalarTypeDefinition production.
	ExitScalarTypeDefinition(c *ScalarTypeDefinitionContext)

	// ExitScalarTypeExtensionDefinition is called when exiting the scalarTypeExtensionDefinition production.
	ExitScalarTypeExtensionDefinition(c *ScalarTypeExtensionDefinitionContext)

	// ExitObjectTypeDefinition is called when exiting the objectTypeDefinition production.
	ExitObjectTypeDefinition(c *ObjectTypeDefinitionContext)

	// ExitObjectTypeExtensionDefinition is called when exiting the objectTypeExtensionDefinition production.
	ExitObjectTypeExtensionDefinition(c *ObjectTypeExtensionDefinitionContext)

	// ExitImplementsInterfaces is called when exiting the implementsInterfaces production.
	ExitImplementsInterfaces(c *ImplementsInterfacesContext)

	// ExitFieldsDefinition is called when exiting the fieldsDefinition production.
	ExitFieldsDefinition(c *FieldsDefinitionContext)

	// ExitExtensionFieldsDefinition is called when exiting the extensionFieldsDefinition production.
	ExitExtensionFieldsDefinition(c *ExtensionFieldsDefinitionContext)

	// ExitFieldDefinition is called when exiting the fieldDefinition production.
	ExitFieldDefinition(c *FieldDefinitionContext)

	// ExitArgumentsDefinition is called when exiting the argumentsDefinition production.
	ExitArgumentsDefinition(c *ArgumentsDefinitionContext)

	// ExitInputValueDefinition is called when exiting the inputValueDefinition production.
	ExitInputValueDefinition(c *InputValueDefinitionContext)

	// ExitInterfaceTypeDefinition is called when exiting the interfaceTypeDefinition production.
	ExitInterfaceTypeDefinition(c *InterfaceTypeDefinitionContext)

	// ExitInterfaceTypeExtensionDefinition is called when exiting the interfaceTypeExtensionDefinition production.
	ExitInterfaceTypeExtensionDefinition(c *InterfaceTypeExtensionDefinitionContext)

	// ExitUnionTypeDefinition is called when exiting the unionTypeDefinition production.
	ExitUnionTypeDefinition(c *UnionTypeDefinitionContext)

	// ExitUnionTypeExtensionDefinition is called when exiting the unionTypeExtensionDefinition production.
	ExitUnionTypeExtensionDefinition(c *UnionTypeExtensionDefinitionContext)

	// ExitUnionMembership is called when exiting the unionMembership production.
	ExitUnionMembership(c *UnionMembershipContext)

	// ExitUnionMembers is called when exiting the unionMembers production.
	ExitUnionMembers(c *UnionMembersContext)

	// ExitEnumTypeDefinition is called when exiting the enumTypeDefinition production.
	ExitEnumTypeDefinition(c *EnumTypeDefinitionContext)

	// ExitEnumTypeExtensionDefinition is called when exiting the enumTypeExtensionDefinition production.
	ExitEnumTypeExtensionDefinition(c *EnumTypeExtensionDefinitionContext)

	// ExitEnumValueDefinitions is called when exiting the enumValueDefinitions production.
	ExitEnumValueDefinitions(c *EnumValueDefinitionsContext)

	// ExitExtensionEnumValueDefinitions is called when exiting the extensionEnumValueDefinitions production.
	ExitExtensionEnumValueDefinitions(c *ExtensionEnumValueDefinitionsContext)

	// ExitEnumValueDefinition is called when exiting the enumValueDefinition production.
	ExitEnumValueDefinition(c *EnumValueDefinitionContext)

	// ExitInputObjectTypeDefinition is called when exiting the inputObjectTypeDefinition production.
	ExitInputObjectTypeDefinition(c *InputObjectTypeDefinitionContext)

	// ExitInputObjectTypeExtensionDefinition is called when exiting the inputObjectTypeExtensionDefinition production.
	ExitInputObjectTypeExtensionDefinition(c *InputObjectTypeExtensionDefinitionContext)

	// ExitInputObjectValueDefinitions is called when exiting the inputObjectValueDefinitions production.
	ExitInputObjectValueDefinitions(c *InputObjectValueDefinitionsContext)

	// ExitExtensionInputObjectValueDefinitions is called when exiting the extensionInputObjectValueDefinitions production.
	ExitExtensionInputObjectValueDefinitions(c *ExtensionInputObjectValueDefinitionsContext)

	// ExitDirectiveDefinition is called when exiting the directiveDefinition production.
	ExitDirectiveDefinition(c *DirectiveDefinitionContext)

	// ExitDirectiveLocation is called when exiting the directiveLocation production.
	ExitDirectiveLocation(c *DirectiveLocationContext)

	// ExitDirectiveLocations is called when exiting the directiveLocations production.
	ExitDirectiveLocations(c *DirectiveLocationsContext)

	// ExitPartialFieldDefinition is called when exiting the partialFieldDefinition production.
	ExitPartialFieldDefinition(c *PartialFieldDefinitionContext)

	// ExitPartialObjectTypeDefinition is called when exiting the partialObjectTypeDefinition production.
	ExitPartialObjectTypeDefinition(c *PartialObjectTypeDefinitionContext)

	// ExitPartialInputObjectTypeDefinition is called when exiting the partialInputObjectTypeDefinition production.
	ExitPartialInputObjectTypeDefinition(c *PartialInputObjectTypeDefinitionContext)

	// ExitPartialInputValueDefinition is called when exiting the partialInputValueDefinition production.
	ExitPartialInputValueDefinition(c *PartialInputValueDefinitionContext)

	// ExitPartialEnumTypeDefinition is called when exiting the partialEnumTypeDefinition production.
	ExitPartialEnumTypeDefinition(c *PartialEnumTypeDefinitionContext)

	// ExitPartialInterfaceTypeDefinition is called when exiting the partialInterfaceTypeDefinition production.
	ExitPartialInterfaceTypeDefinition(c *PartialInterfaceTypeDefinitionContext)

	// ExitPartialUnionTypeDefinition is called when exiting the partialUnionTypeDefinition production.
	ExitPartialUnionTypeDefinition(c *PartialUnionTypeDefinitionContext)

	// ExitPartialScalarTypeDefinition is called when exiting the partialScalarTypeDefinition production.
	ExitPartialScalarTypeDefinition(c *PartialScalarTypeDefinitionContext)

	// ExitTsResolverFieldDefinition is called when exiting the tsResolverFieldDefinition production.
	ExitTsResolverFieldDefinition(c *TsResolverFieldDefinitionContext)
}
