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

package schema

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/housecanary/gq/ast"
)

var introspectionStringType = &ScalarType{
	named: named{"String"},
	encode: func(ctx context.Context, v interface{}) (LiteralValue, error) {
		if v == nil {
			return nil, nil
		}
		return LiteralString(v.(string)), nil
	},
	decode: func(ctx context.Context, v LiteralValue) (interface{}, error) {
		if v == nil {
			return nil, nil
		}
		return string(v.(LiteralString)), nil
	},
}

var introspectionBoolType = &ScalarType{
	named: named{"Boolean"},
	encode: func(ctx context.Context, v interface{}) (LiteralValue, error) {
		return LiteralBool(v.(bool)), nil
	},
	decode: func(ctx context.Context, v LiteralValue) (interface{}, error) {
		return bool(v.(LiteralBool)), nil
	},
}

type iTypeKind string

var introspectionTypeKindType = &EnumType{
	named: named{"__TypeKind"},
	values: map[LiteralString]*enumValueDescriptor{
		"SCALAR":       {named: named{"SCALAR"}},
		"OBJECT":       {named: named{"OBJECT"}},
		"INTERFACE":    {named: named{"INTERFACE"}},
		"UNION":        {named: named{"UNION"}},
		"ENUM":         {named: named{"ENUM"}},
		"INPUT_OBJECT": {named: named{"INPUT_OBJECT"}},
		"LIST":         {named: named{"LIST"}},
		"NON_NULL":     {named: named{"NON_NULL"}},
	},
	encode: func(ctx context.Context, v interface{}) (LiteralValue, error) {
		if v == nil {
			return nil, nil
		}

		tk, ok := v.(iTypeKind)
		if !ok {
			return nil, fmt.Errorf("Value is not an iTypeKind")
		}
		if tk == "" {
			return nil, nil
		}

		return LiteralString(tk), nil
	},
}

var introspectionDirectiveLocationType = &EnumType{
	named: named{"__DirectiveLocation"},
	values: map[LiteralString]*enumValueDescriptor{
		"QUERY":                  {named: named{"QUERY"}},
		"MUTATION":               {named: named{"MUTATION"}},
		"SUBSCRIPTION":           {named: named{"SUBSCRIPTION"}},
		"FIELD":                  {named: named{"FIELD"}},
		"FRAGMENT_DEFINITION":    {named: named{"FRAGMENT_DEFINITION"}},
		"FRAGMENT_SPREAD":        {named: named{"FRAGMENT_SPREAD"}},
		"INLINE_FRAGMENT":        {named: named{"INLINE_FRAGMENT"}},
		"SCHEMA":                 {named: named{"SCHEMA"}},
		"SCALAR":                 {named: named{"SCALAR"}},
		"OBJECT":                 {named: named{"OBJECT"}},
		"FIELD_DEFINITION":       {named: named{"FIELD_DEFINITION"}},
		"ARGUMENT_DEFINITION":    {named: named{"ARGUMENT_DEFINITION"}},
		"INTERFACE":              {named: named{"INTERFACE"}},
		"UNION":                  {named: named{"UNION"}},
		"ENUM":                   {named: named{"ENUM"}},
		"ENUM_VALUE":             {named: named{"ENUM_VALUE"}},
		"INPUT_OBJECT":           {named: named{"INPUT_OBJECT"}},
		"INPUT_FIELD_DEFINITION": {named: named{"INPUT_FIELD_DEFINITION"}},
	},
	encode: func(ctx context.Context, v interface{}) (LiteralValue, error) {
		if v == nil {
			return nil, nil
		}

		tk, ok := v.(DirectiveLocation)
		if !ok {
			return nil, fmt.Errorf("Value is not a DirectiveLocation")
		}
		if tk == "" {
			return nil, nil
		}

		return LiteralString(tk), nil
	},
}

var introspectionSchemaType = &ObjectType{
	named: named{"__Schema"},
	fieldsByName: map[string]*FieldDescriptor{
		"types": {
			named: named{"types"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				s := v.(*Schema)
				types := make([]Type, 0, len(s.allTypes))
				for _, v := range s.allTypes {
					types = append(types, v)
				}
				return listOfTypes(types), nil
			}),
			typ: &ListType{&NotNilType{introspectionTypeType}},
		},
		"queryType": {
			named: named{"queryType"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return v.(*Schema).QueryType, nil
			}),
			typ: &NotNilType{introspectionTypeType},
		},
		"mutationType": {
			named: named{"mutationType"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return nil, nil // FUTURE: Mutations not supported
			}),
			typ: introspectionTypeType,
		},
		"subscriptionType": {
			named: named{"subscriptionType"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return nil, nil // FUTURE: Subscriptions not supported
			}),
			typ: introspectionTypeType,
		},
		"directives": {
			named: named{"directives"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return listOfDirectiveDefinitions(v.(*Schema).directives), nil
			}),
			typ: &NotNilType{&ListType{&NotNilType{introspectionDirectiveType}}},
		},
	},
}

var introspectionTypeType = &ObjectType{
	named: named{"__Type"},
}

var introspectionTypeTypeFields = map[string]*FieldDescriptor{
	"kind": {
		named: named{"kind"},
		r: SimpleResolver(func(v interface{}) (interface{}, error) {
			switch v.(type) {
			case *ScalarType:
				return iTypeKind("SCALAR"), nil
			case *ObjectType:
				return iTypeKind("OBJECT"), nil
			case *InterfaceType:
				return iTypeKind("INTERFACE"), nil
			case *UnionType:
				return iTypeKind("UNION"), nil
			case *EnumType:
				return iTypeKind("ENUM"), nil
			case *InputObjectType:
				return iTypeKind("INPUT_OBJECT"), nil
			case *NotNilType:
				return iTypeKind("NON_NULL"), nil
			case *ListType:
				return iTypeKind("LIST"), nil
			}
			panic("Unhandled type")
		}),
		typ: &NotNilType{introspectionTypeKindType},
	},
	"name": {
		named: named{"name"},
		r: SimpleResolver(func(v interface{}) (interface{}, error) {
			switch t := v.(type) {
			case *ScalarType:
				return t.name, nil
			case *ObjectType:
				return t.name, nil
			case *InterfaceType:
				return t.name, nil
			case *UnionType:
				return t.name, nil
			case *EnumType:
				return t.name, nil
			case *InputObjectType:
				return t.name, nil
			case *NotNilType:
				return nil, nil
			case *ListType:
				return nil, nil
			}
			panic("Unhandled type")
		}),
		typ: introspectionStringType,
	},
	"description": {
		named: named{"description"},
		r: SimpleResolver(func(v interface{}) (interface{}, error) {
			switch t := v.(type) {
			case *ScalarType:
				return t.description, nil
			case *ObjectType:
				return t.description, nil
			case *InterfaceType:
				return t.description, nil
			case *UnionType:
				return t.description, nil
			case *EnumType:
				return t.description, nil
			case *InputObjectType:
				return t.description, nil
			case *NotNilType:
				return nil, nil
			case *ListType:
				return nil, nil
			}
			panic("Unhandled type")
		}),
		typ: introspectionStringType,
	},
	"fields": {
		named: named{"fields"},
		arguments: []*ArgumentDescriptor{
			{
				named:        named{"includeDeprecated"},
				typ:          introspectionBoolType,
				defaultValue: ast.BooleanValue{V: false},
			},
		},
		r: FullResolver(func(rc ResolverContext, v interface{}) (interface{}, error) {
			includeDeprecated, err := rc.GetArgumentValue("includeDeprecated")
			if err != nil {
				return nil, err
			}

			switch t := v.(type) {
			case *ObjectType:
				return listOfFieldDescriptors(makeIntrospectionFields(t.fieldsByName, includeDeprecated.(bool))), nil
			case *InterfaceType:
				return listOfFieldDescriptors(makeIntrospectionFields(t.fields, includeDeprecated.(bool))), nil
			}
			return nil, nil
		}),
		typ: &ListType{&NotNilType{introspectionFieldType}},
	},
	"interfaces": {
		named: named{"interfaces"},
		r: SimpleResolver(func(v interface{}) (interface{}, error) {
			switch t := v.(type) {
			case *ObjectType:
				return listOfInterfaceTypes(t.interfaces), nil
			}
			return nil, nil
		}),
		typ: &ListType{&NotNilType{introspectionTypeType}},
	},
	"possibleTypes": {
		named: named{"possibleTypes"},
		r: SimpleResolver(func(v interface{}) (interface{}, error) {
			switch t := v.(type) {
			case *UnionType:
				return listOfObjectTypes(t.members), nil
			case *InterfaceType:
				return listOfObjectTypes(t.implementations), nil
			}
			return nil, nil
		}),
		typ: &ListType{&NotNilType{introspectionTypeType}},
	},
	"enumValues": {
		named: named{"enumValues"},
		arguments: []*ArgumentDescriptor{
			{
				named:        named{"includeDeprecated"},
				typ:          introspectionBoolType,
				defaultValue: ast.BooleanValue{V: false},
			},
		},
		r: FullResolver(func(rc ResolverContext, v interface{}) (interface{}, error) {
			includeDeprecated, err := rc.GetArgumentValue("includeDeprecated")
			if err != nil {
				return nil, err
			}
			switch t := v.(type) {
			case *EnumType:
				return listOfEnumValues(makeIntrospectionEnumValues(t, includeDeprecated.(bool))), nil
			}
			return nil, nil

		}),
		typ: &ListType{&NotNilType{introspectionEnumValueType}},
	},
	"inputFields": {
		named: named{"inputFields"},
		r: SimpleResolver(func(v interface{}) (interface{}, error) {
			switch t := v.(type) {
			case *InputObjectType:
				fields := make([]*InputObjectFieldDescriptor, 0, len(t.fields))
				for _, f := range t.fields {
					fields = append(fields, f)
				}
				return listOfInputObjectFieldDescriptors(fields), nil
			}
			return nil, nil
		}),
		typ: &ListType{&NotNilType{introspectionInputValueType}},
	},
	"ofType": {
		named: named{"inputFields"},
		r: SimpleResolver(func(v interface{}) (interface{}, error) {
			switch t := v.(type) {
			case *NotNilType:
				return t.of, nil
			case *ListType:
				return t.of, nil
			}
			return nil, nil
		}),
		typ: introspectionTypeType,
	},
}

func init() {
	introspectionTypeType.fieldsByName = introspectionTypeTypeFields
}

var introspectionFieldType = &ObjectType{
	named: named{"__Field"},
	fieldsByName: map[string]*FieldDescriptor{
		"name": {
			named: named{"name"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return v.(*FieldDescriptor).name, nil
			}),
			typ: &NotNilType{introspectionStringType},
		},
		"description": {
			named: named{"description"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return v.(*FieldDescriptor).description, nil
			}),
			typ: introspectionStringType,
		},
		"args": {
			named: named{"args"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return listOfArgumentDescriptors(v.(*FieldDescriptor).arguments), nil
			}),
			typ: &NotNilType{&ListType{&NotNilType{introspectionInputValueType}}},
		},
		"type": {
			named: named{"type"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return v.(*FieldDescriptor).typ, nil
			}),
			typ: &NotNilType{introspectionTypeType},
		},
		"isDeprecated": {
			named: named{"isDeprecated"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				deprecated, _ := checkDeprecated(v.(*FieldDescriptor).schemaElement)
				return deprecated, nil
			}),
			typ: &NotNilType{introspectionBoolType},
		},
		"deprecationReason": {
			named: named{"deprecationReason"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				deprecated, reason := checkDeprecated(v.(*FieldDescriptor).schemaElement)
				if !deprecated {
					return nil, nil
				}
				return reason, nil
			}),
			typ: introspectionStringType,
		},
	},
}

var introspectionInputValueType = &ObjectType{
	named: named{"__InputValue"},
	fieldsByName: map[string]*FieldDescriptor{
		"name": {
			named: named{"name"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				switch t := v.(type) {
				case *ArgumentDescriptor:
					return t.name, nil
				case *InputObjectFieldDescriptor:
					return t.name, nil
				}
				panic("Unknown input value type")
			}),
			typ: &NotNilType{introspectionStringType},
		},
		"description": {
			named: named{"description"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				switch t := v.(type) {
				case *ArgumentDescriptor:
					return t.description, nil
				case *InputObjectFieldDescriptor:
					return t.description, nil
				}
				panic("Unknown input value type")
			}),
			typ: introspectionStringType,
		},
		"type": {
			named: named{"type"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				switch t := v.(type) {
				case *ArgumentDescriptor:
					return t.typ, nil
				case *InputObjectFieldDescriptor:
					return t.typ, nil
				}
				panic("Unknown input value type")
			}),
			typ: &NotNilType{introspectionTypeType},
		},
		"defaultValue": {
			named: named{"defaultValue"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				switch t := v.(type) {
				case *ArgumentDescriptor:
					return makeIntrospectionDefaultValue(t.defaultValue, t.typ), nil
				case *InputObjectFieldDescriptor:
					return makeIntrospectionDefaultValue(t.defaultValueAst, t.typ), nil
				}
				panic("Unknown input value type")
			}),
			typ: introspectionStringType,
		},
	},
}

var introspectionEnumValueType = &ObjectType{
	named: named{"__EnumValue"},
	fieldsByName: map[string]*FieldDescriptor{
		"name": {
			named: named{"name"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return v.(*enumValueDescriptor).name, nil
			}),
			typ: &NotNilType{introspectionStringType},
		},
		"description": {
			named: named{"description"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return v.(*enumValueDescriptor).description, nil
			}),
			typ: introspectionStringType,
		},
		"isDeprecated": {
			named: named{"isDeprecated"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				deprecated, _ := checkDeprecated(v.(*enumValueDescriptor).schemaElement)
				return deprecated, nil
			}),
			typ: &NotNilType{introspectionBoolType},
		},
		"deprecationReason": {
			named: named{"deprecationReason"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				deprecated, reason := checkDeprecated(v.(*enumValueDescriptor).schemaElement)
				if !deprecated {
					return nil, nil
				}
				return reason, nil
			}),
			typ: introspectionStringType,
		},
	},
}

var introspectionDirectiveType = &ObjectType{
	named: named{"__Directive"},
	fieldsByName: map[string]*FieldDescriptor{
		"name": {
			named: named{"name"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return v.(*DirectiveDefinition).name, nil
			}),
			typ: &NotNilType{introspectionStringType},
		},
		"description": {
			named: named{"description"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return v.(*DirectiveDefinition).description, nil
			}),
			typ: introspectionStringType,
		},
		"args": {
			named: named{"args"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				dd := v.(*DirectiveDefinition)
				return listOfArgumentDescriptors(dd.arguments), nil
			}),
			typ: &NotNilType{&ListType{&NotNilType{introspectionInputValueType}}},
		},
		"locations": {
			named: named{"type"},
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return listOfDirectiveLocations(v.(*DirectiveDefinition).locations), nil
			}),
			typ: &NotNilType{&ListType{&NotNilType{introspectionDirectiveLocationType}}},
		},
	},
}

func makeIntrospectionFields(fieldsByName map[string]*FieldDescriptor, includeDeprecated bool) []*FieldDescriptor {
	r := make([]*FieldDescriptor, 0, len(fieldsByName))
	for _, v := range fieldsByName {
		if strings.HasPrefix(v.name, "__") {
			continue
		}
		deprecated, _ := checkDeprecated(v.schemaElement)
		if !includeDeprecated && deprecated {
			continue
		}
		r = append(r, v)
	}
	sort.Stable(sortFieldDescriptorsByName(r))
	return r
}

type sortFieldDescriptorsByName []*FieldDescriptor

func (s sortFieldDescriptorsByName) Len() int {
	return len(s)
}

func (s sortFieldDescriptorsByName) Less(i, j int) bool {
	return s[i].Name() < s[j].Name()
}

func (s sortFieldDescriptorsByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type sortEnumValuesByName []*enumValueDescriptor

func (s sortEnumValuesByName) Len() int {
	return len(s)
}

func (s sortEnumValuesByName) Less(i, j int) bool {
	return s[i].name < s[j].name
}

func (s sortEnumValuesByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type sortInputObjectFieldDescriptorsByName []*InputObjectFieldDescriptor

func (s sortInputObjectFieldDescriptorsByName) Len() int {
	return len(s)
}

func (s sortInputObjectFieldDescriptorsByName) Less(i, j int) bool {
	return s[i].name < s[j].name
}

func (s sortInputObjectFieldDescriptorsByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type sortTypesBySignature []Type

func (s sortTypesBySignature) Len() int {
	return len(s)
}

func (s sortTypesBySignature) Less(i, j int) bool {
	return s[i].signature() < s[j].signature()
}

func (s sortTypesBySignature) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type sortDirectiveDefsByName []*DirectiveDefinition

func (s sortDirectiveDefsByName) Len() int {
	return len(s)
}

func (s sortDirectiveDefsByName) Less(i, j int) bool {
	return s[i].name < s[j].name
}

func (s sortDirectiveDefsByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func makeIntrospectionEnumValues(typ *EnumType, includeDeprecated bool) []*enumValueDescriptor {
	r := make([]*enumValueDescriptor, 0, len(typ.values))
	for _, v := range typ.values {
		deprecated, _ := checkDeprecated(v.schemaElement)
		if !includeDeprecated && deprecated {
			continue
		}
		r = append(r, v)
	}
	return r
}

func makeIntrospectionDefaultValue(val ast.Value, forType Type) interface{} {
	if val == nil {
		return nil
	}

	return val.Representation()
}

func checkDeprecated(se schemaElement) (bool, string) {
	for _, d := range se.directives {
		if d.name == "deprecated" {
			reason := "No longer supported"
			for _, a := range d.arguments {
				if a.name == "reason" {
					if s, ok := a.value.(ast.StringValue); ok {
						reason = string(s.Representation())
					}
				}
			}
			return true, reason
		}
	}

	return false, ""
}

type genericList struct {
	len  int
	item func(i int) interface{}
}

func (l genericList) Len() int {
	return l.len
}

func (l genericList) ForEachElement(cb ListValueCallback) {
	for i := 0; i < l.len; i++ {
		cb(l.item(i))
	}
}

func listOfTypes(t []Type) ListValue {
	return genericList{
		len(t),
		func(i int) interface{} {
			return t[i]
		},
	}
}

func listOfDirectiveDefinitions(t []*DirectiveDefinition) ListValue {
	return genericList{
		len(t),
		func(i int) interface{} {
			return t[i]
		},
	}
}

func listOfFieldDescriptors(t []*FieldDescriptor) ListValue {
	return genericList{
		len(t),
		func(i int) interface{} {
			return t[i]
		},
	}
}

func listOfInterfaceTypes(t []*InterfaceType) ListValue {
	return genericList{
		len(t),
		func(i int) interface{} {
			return t[i]
		},
	}
}

func listOfObjectTypes(t []*ObjectType) ListValue {
	return genericList{
		len(t),
		func(i int) interface{} {
			return t[i]
		},
	}
}

func listOfEnumValues(t []*enumValueDescriptor) ListValue {
	return genericList{
		len(t),
		func(i int) interface{} {
			return t[i]
		},
	}
}

func listOfInputObjectFieldDescriptors(t []*InputObjectFieldDescriptor) ListValue {
	return genericList{
		len(t),
		func(i int) interface{} {
			return t[i]
		},
	}
}

func listOfArgumentDescriptors(t []*ArgumentDescriptor) ListValue {
	return genericList{
		len(t),
		func(i int) interface{} {
			return t[i]
		},
	}
}

func listOfDirectiveLocations(t []DirectiveLocation) ListValue {
	return genericList{
		len(t),
		func(i int) interface{} {
			return t[i]
		},
	}
}
