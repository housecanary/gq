/*
Derived from the GraphQL-Java implementation, according to MIT license.

The MIT License (MIT)

Copyright (c) 2015 Andreas Marek and Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy of 
this software and associated documentation files (the "Software"), to deal in 
the Software without restriction, including without limitation the rights to 
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies 
of the Software, and to permit persons to whom the Software is furnished to do 
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all 
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR 
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, 
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE 
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER 
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, 
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE 
SOFTWARE.
 */
 
grammar Graphql;

/* Common */
operationType : SUBSCRIPTION | MUTATION | QUERY;

enumValue : name ;

arrayValue: '[' value* ']';

arrayValueWithVariable: '[' valueWithVariable* ']';



objectValue: '{' objectField* '}';
objectValueWithVariable: '{' objectFieldWithVariable* '}';
objectField : name ':' value;
objectFieldWithVariable : name ':' valueWithVariable;


directives : directive+;

directive :'@' name arguments?;


arguments : '(' argument+ ')';

argument : name ':' valueWithVariable;

name: NAME | FRAGMENT | QUERY | MUTATION | SUBSCRIPTION | SCHEMA | SCALAR | TYPE | INTERFACE | IMPLEMENTS | ENUM | UNION | INPUT | EXTEND | DIRECTIVE;

value :
stringValue |
IntValue |
FloatValue |
BooleanValue |
NullValue |
enumValue |
arrayValue |
objectValue;


valueWithVariable :
variable |
stringValue |
IntValue |
FloatValue |
BooleanValue |
NullValue |
enumValue |
arrayValueWithVariable |
objectValueWithVariable;


variable : '$' name;

defaultValue : '=' value;

stringValue
 : TripleQuotedStringValue
 | StringValue
 ;
gqlType : typeName | listType | nonNullType;

typeName : name;
listType : '[' gqlType ']';
nonNullType: typeName '!' | listType '!';


BooleanValue: 'true' | 'false';

NullValue: 'null';

FRAGMENT: 'fragment';
QUERY: 'query';
MUTATION: 'mutation';
SUBSCRIPTION: 'subscription';
SCHEMA: 'schema';
SCALAR: 'scalar';
TYPE: 'type';
INTERFACE: 'interface';
IMPLEMENTS: 'implements';
ENUM: 'enum';
UNION: 'union';
INPUT: 'input';
EXTEND: 'extend';
DIRECTIVE: 'directive';
NAME: [_A-Za-z][_0-9A-Za-z]*;


IntValue : Sign? IntegerPart;

FloatValue : Sign? IntegerPart ('.' Digit+)? ExponentPart?;

Sign : '-';

IntegerPart : '0' | NonZeroDigit | NonZeroDigit Digit+;

NonZeroDigit: '1'.. '9';

ExponentPart : ('e'|'E') Sign? Digit+;

Digit : '0'..'9';


StringValue
 : '"' ( ~["\\\n\r\u2028\u2029] | EscapedChar )* '"'
 ;
TripleQuotedStringValue
 : '"""' TripleQuotedStringPart? '"""'
 ;
// Fragments never become a token of their own: they are only used inside other lexer rules
fragment TripleQuotedStringPart : ( EscapedTripleQuote | SourceCharacter )+?;
fragment EscapedTripleQuote : '\\"""';
fragment SourceCharacter :[\u0009\u000A\u000D\u0020-\uFFFF];
Comment: '#' ~[\n\r\u2028\u2029]* -> channel(2);
Ignored: (UnicodeBOM|Whitespace|LineTerminator|Comma) -> skip;
fragment EscapedChar :   '\\' (["\\/bfnrt] | Unicode) ;
fragment Unicode : 'u' Hex Hex Hex Hex ;
fragment Hex : [0-9a-fA-F] ;

fragment LineTerminator: [\n\r\u2028\u2029];

fragment Whitespace : [\u0009\u0020];
fragment Comma : ',';
fragment UnicodeBOM : [\ufeff];

/* Operation */
operationDefinition:
selectionSet |
operationType  name? variableDefinitions? directives? selectionSet;

variableDefinitions : '(' variableDefinition+ ')';

variableDefinition : variable ':' gqlType defaultValue?;


selectionSet :  '{' selection+ '}';

selection :
field |
fragmentSpread |
inlineFragment;

field : alias? name arguments? directives? selectionSet?;

alias : name ':';




fragmentSpread : '...' fragmentName directives?;

inlineFragment : '...' typeCondition? directives? selectionSet;

fragmentDefinition : 'fragment' fragmentName typeCondition directives? selectionSet;

fragmentName :  name;

typeCondition : 'on' typeName;

/* Document */
document : (operationDefinition | fragmentDefinition)+;

/* Schema */
description : stringValue;

typeSystemDefinition: description?
schemaDefinition |
typeDefinition |
typeExtension |
directiveDefinition
;

schemaDefinition : description? SCHEMA directives? '{' operationTypeDefinition+ '}';

operationTypeDefinition : description? operationType ':' typeName;

typeDefinition:
scalarTypeDefinition |
objectTypeDefinition |
interfaceTypeDefinition |
unionTypeDefinition |
enumTypeDefinition |
inputObjectTypeDefinition
;

//
// type extensions dont get "description" strings according to spec
// https://github.com/facebook/graphql/blob/master/spec/Appendix%20B%20--%20Grammar%20Summary.md
//

typeExtension :
    objectTypeExtensionDefinition |
    interfaceTypeExtensionDefinition |
    unionTypeExtensionDefinition |
    scalarTypeExtensionDefinition |
    enumTypeExtensionDefinition |
    inputObjectTypeExtensionDefinition
;


scalarTypeDefinition : description? SCALAR name directives?;

scalarTypeExtensionDefinition : EXTEND SCALAR name directives?;

objectTypeDefinition : description? TYPE name implementsInterfaces? directives? fieldsDefinition?;

objectTypeExtensionDefinition : EXTEND TYPE name implementsInterfaces? directives? fieldsDefinition?;

implementsInterfaces :
    IMPLEMENTS '&'? typeName+ |
    implementsInterfaces '&' typeName ;

fieldsDefinition : '{' fieldDefinition* '}';

fieldDefinition : description? name argumentsDefinition? ':' gqlType directives?;

argumentsDefinition : '(' inputValueDefinition+ ')';

inputValueDefinition : description? name ':' gqlType defaultValue? directives?;

interfaceTypeDefinition : description? INTERFACE name directives? fieldsDefinition?;

interfaceTypeExtensionDefinition : EXTEND INTERFACE name directives? fieldsDefinition?;


unionTypeDefinition : description? UNION name directives? unionMembership;

unionTypeExtensionDefinition : EXTEND UNION name directives? unionMembership?;

unionMembership : '=' unionMembers;

unionMembers:
'|'? typeName |
unionMembers '|' typeName
;

enumTypeDefinition : description? ENUM name directives? enumValueDefinitions;

enumTypeExtensionDefinition : EXTEND ENUM name directives? enumValueDefinitions?;

enumValueDefinitions : '{' enumValueDefinition+ '}';

enumValueDefinition : description? enumValue directives?;


inputObjectTypeDefinition : description? INPUT name directives? inputObjectValueDefinitions;

inputObjectTypeExtensionDefinition : EXTEND INPUT name directives? inputObjectValueDefinitions?;

inputObjectValueDefinitions : '{' inputValueDefinition+ '}';


directiveDefinition : description? DIRECTIVE '@' name argumentsDefinition? 'on' directiveLocations;

directiveLocation : name;

directiveLocations :
directiveLocation |
directiveLocations '|' directiveLocation
;

/* Partial definitions */
partialFieldDefinition : 
    name argumentsDefinition? ':' gqlType directives? |
    argumentsDefinition ':' gqlType directives? |
    name argumentsDefinition? directives? |
    ':' gqlType directives? |
    argumentsDefinition |
    directives;

partialObjectTypeDefinition : 
    description? TYPE name implementsInterfaces? directives? fieldsDefinition? |
    description? name implementsInterfaces? directives? fieldsDefinition? |
    description? implementsInterfaces? directives? fieldsDefinition?;

partialInputObjectTypeDefinition : 
    description? INPUT name directives? inputObjectValueDefinitions |
    description? name directives? inputObjectValueDefinitions |
    description? directives? inputObjectValueDefinitions |
    description? directives?;

partialInputValueDefinition : 
    description? name ':' gqlType defaultValue? directives? |
    description? name defaultValue? directives? |
    description? ':' gqlType defaultValue? directives?;

partialEnumTypeDefinition :
    description? ENUM name directives? enumValueDefinitions |
    description? name directives? enumValueDefinitions |
    description? directives? enumValueDefinitions;

partialInterfaceTypeDefinition :
    description? INTERFACE name directives? fieldsDefinition |
    description? name directives? fieldsDefinition |
    description? directives? fieldsDefinition;

partialUnionTypeDefinition :
    description? UNION name directives? unionMembership? |
    description? name directives? unionMembership? |
    description? directives? unionMembership?;

partialScalarTypeDefinition : description? name? directives?;