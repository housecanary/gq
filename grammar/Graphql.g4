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

@lexer::header {
    func isDigit(c int) bool {
        return c >= '0' && c <= '9';
    }
    func isNameStart(c int) bool {
        return '_' == c ||
          (c >= 'A' && c <= 'Z') ||
          (c >= 'a' && c <= 'z');
    }
    func isDot(c int) bool {
        return '.' == c;
    }
}

/* Common */
operationType : SUBSCRIPTION | MUTATION | QUERY;

description : StringValue;

enumValue : enumValueName ;


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

baseName: NAME | FRAGMENT | QUERY | MUTATION | SUBSCRIPTION | SCHEMA | SCALAR | TYPE | INTERFACE | IMPLEMENTS | ENUM | UNION | INPUT | EXTEND | DIRECTIVE | REPEATABLE;
fragmentName: baseName | BooleanValue | NullValue;
enumValueName: baseName | ON_KEYWORD;

name: baseName | BooleanValue | NullValue | ON_KEYWORD;

value :
StringValue |
IntValue |
FloatValue |
BooleanValue |
NullValue |
enumValue |
arrayValue |
objectValue;


valueWithVariable :
variable |
StringValue |
IntValue |
FloatValue |
BooleanValue |
NullValue |
enumValue |
arrayValueWithVariable |
objectValueWithVariable;


variable : '$' name;

defaultValue : '=' value;

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
ON_KEYWORD: 'on';
REPEATABLE: 'repeatable';
NAME: [_A-Za-z][_0-9A-Za-z]*;


// Int Value
IntValue :  IntegerPart { !isDigit(p.GetInputStream().LA(1)) && !isDot(p.GetInputStream().LA(1)) && !isNameStart(p.GetInputStream().LA(1))  }?;
fragment IntegerPart : NegativeSign? '0' | NegativeSign? NonZeroDigit Digit*;
fragment NegativeSign : '-';
fragment NonZeroDigit: '1'..'9';

// Float Value
FloatValue : ((IntegerPart FractionalPart ExponentPart) { !isDigit(p.GetInputStream().LA(1)) && !isDot(p.GetInputStream().LA(1)) && !isNameStart(p.GetInputStream().LA(1))  }?) |
    ((IntegerPart FractionalPart ) { !isDigit(p.GetInputStream().LA(1)) && !isDot(p.GetInputStream().LA(1)) && !isNameStart(p.GetInputStream().LA(1))  }?) |
    ((IntegerPart ExponentPart) { !isDigit(p.GetInputStream().LA(1)) && !isDot(p.GetInputStream().LA(1)) && !isNameStart(p.GetInputStream().LA(1))  }?);
fragment FractionalPart: '.' Digit+;
fragment ExponentPart :  ExponentIndicator Sign? Digit+;
fragment ExponentIndicator: 'e' | 'E';
fragment Sign: '+'|'-';
fragment Digit : '0'..'9';


// StringValue
StringValue:
'""'  { p.GetInputStream().LA(1) != '"'}? |
'"' StringCharacter+ '"' |
'"""' BlockStringCharacter*? '"""';

fragment BlockStringCharacter:
'\\"""'|
SourceCharacter;


fragment StringCharacter:
[\u0009\u0020\u0021\u0023-\u005B\u005D-\uFFFF] |
'\\u' EscapedUnicode  |
'\\' EscapedCharacter;

fragment EscapedCharacter :  ["\\/bfnrt];
fragment EscapedUnicode : Hex Hex Hex Hex;
fragment Hex : [0-9a-fA-F];
fragment SourceCharacter :[\u0009\u000A\u000D\u0020-\uFFFF];
Ignored: (UnicodeBOM|WhiteSpace|LineTerminator|Comment|Comma) -> skip;
fragment UnicodeBOM : [\ufeff];
fragment WhiteSpace : [\u0009\u0020];
fragment LineTerminator: '\r\n' | [\n\r];
fragment CommentChar : [\u0009\u0020-\uFFFF];
fragment Comment: '#' CommentChar*;
fragment Comma : ',';

/* Operation */
operationDefinition:
selectionSet |
operationType  name? variableDefinitions? directives? selectionSet;

variableDefinitions : '(' variableDefinition+ ')';

variableDefinition : variable ':' gqlType defaultValue? directives?;


selectionSet :  '{' selection+ '}';

selection :
field |
fragmentSpread |
inlineFragment;

field : alias? name arguments? directives? selectionSet?;

alias : name ':';




fragmentSpread : '...' fragmentName directives?;

inlineFragment : '...' typeCondition? directives? selectionSet;

fragmentDefinition : FRAGMENT fragmentName typeCondition directives? selectionSet;

typeCondition : ON_KEYWORD typeName;

/* Document */
document : (operationDefinition | fragmentDefinition)+;

/* Schema */
typeSystemDefinition:
schemaDefinition |
typeDefinition |
directiveDefinition
;

typeSystemExtension :
schemaExtension |
typeExtension
;

schemaDefinition : description? SCHEMA directives? '{' operationTypeDefinition+ '}';

schemaExtension :
    EXTEND SCHEMA directives? '{' operationTypeDefinition+ '}' |
    EXTEND SCHEMA directives+
;

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

emptyParentheses : '{' '}';

scalarTypeDefinition : description? SCALAR name directives?;

scalarTypeExtensionDefinition : EXTEND SCALAR name directives?;

objectTypeDefinition : description? TYPE name implementsInterfaces? directives? fieldsDefinition?;

objectTypeExtensionDefinition :
    EXTEND TYPE name implementsInterfaces? directives? extensionFieldsDefinition |
    EXTEND TYPE name implementsInterfaces? directives emptyParentheses? |
    EXTEND TYPE name implementsInterfaces
;

implementsInterfaces :
    IMPLEMENTS '&'? typeName+ |
    implementsInterfaces '&' typeName ;

fieldsDefinition : '{' fieldDefinition* '}';

extensionFieldsDefinition : '{' fieldDefinition+ '}';

fieldDefinition : description? name argumentsDefinition? ':' gqlType directives?;

argumentsDefinition : '(' inputValueDefinition+ ')';

inputValueDefinition : description? name ':' gqlType defaultValue? directives?;

interfaceTypeDefinition : description? INTERFACE name directives? fieldsDefinition?;

interfaceTypeExtensionDefinition :
    EXTEND INTERFACE name implementsInterfaces? directives? extensionFieldsDefinition |
    EXTEND INTERFACE name implementsInterfaces? directives emptyParentheses? |
    EXTEND INTERFACE name implementsInterfaces
;


unionTypeDefinition : description? UNION name directives? unionMembership;

unionTypeExtensionDefinition :
    EXTEND UNION name directives? unionMembership |
    EXTEND UNION name directives
;

unionMembership : '=' unionMembers;

unionMembers:
'|'? typeName |
unionMembers '|' typeName
;

enumTypeDefinition : description? ENUM name directives? enumValueDefinitions?;

enumTypeExtensionDefinition :
    EXTEND ENUM name directives? extensionEnumValueDefinitions |
    EXTEND ENUM name directives emptyParentheses?
;

enumValueDefinitions : '{' enumValueDefinition* '}';

extensionEnumValueDefinitions : '{' enumValueDefinition+ '}';

enumValueDefinition : description? enumValue directives?;


inputObjectTypeDefinition : description? INPUT name directives? inputObjectValueDefinitions?;

inputObjectTypeExtensionDefinition :
    EXTEND INPUT name directives? extensionInputObjectValueDefinitions |
    EXTEND INPUT name directives emptyParentheses?
;

inputObjectValueDefinitions : '{' inputValueDefinition* '}';

extensionInputObjectValueDefinitions : '{' inputValueDefinition+ '}';


directiveDefinition : description? DIRECTIVE '@' name argumentsDefinition? REPEATABLE? ON_KEYWORD directiveLocations;

directiveLocation : name;

directiveLocations :
'|'? directiveLocation |
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
    description? ':' gqlType defaultValue? directives? |
    description? defaultValue directives? |
    description? directives;

partialEnumTypeDefinition :
    description? ENUM name directives? enumValueDefinitions |
    description? name directives? enumValueDefinitions |
    description? directives? enumValueDefinitions?;

partialInterfaceTypeDefinition :
    description? INTERFACE name directives? fieldsDefinition |
    description? name directives? fieldsDefinition |
    description? directives? fieldsDefinition;

partialUnionTypeDefinition :
    description? UNION name directives? unionMembership? |
    description? name directives? unionMembership? |
    description? directives? unionMembership?;

partialScalarTypeDefinition : description? name? directives?;

/* TS Definitions */
tsResolverFieldDefinition : 
    description? name directives? |
    description? name ':' gqlType directives?;

tsTypeDefinition :
    description? name? directives?;
