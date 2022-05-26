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

package ast

type BasicTypeDefinition struct {
	Description string
	Name        string
	Directives  Directives
}

type ScalarTypeDefinition struct {
	Description string
	Name        string
	Directives  Directives
}

type ObjectTypeDefinition struct {
	Description          string
	Name                 string
	ImplementsInterfaces ImplementsInterfaces
	Directives           Directives
	FieldsDefinition     FieldsDefinition
}

type ImplementsInterfaces []string

type FieldsDefinition []*FieldDefinition

type FieldDefinition struct {
	Description         string
	Name                string
	ArgumentsDefinition ArgumentsDefinition
	Type                Type
	Directives          Directives
}

type InterfaceTypeDefinition struct {
	Description      string
	Name             string
	Directives       Directives
	FieldsDefinition FieldsDefinition
}

type UnionTypeDefinition struct {
	Description     string
	Name            string
	Directives      Directives
	UnionMembership UnionMembership
}

type UnionMembership []string

type EnumTypeDefinition struct {
	Description          string
	Name                 string
	Directives           Directives
	EnumValueDefinitions EnumValueDefinitions
}

type EnumValueDefinitions []*EnumValueDefinition

type EnumValueDefinition struct {
	Description string
	Value       string
	Directives  Directives
}

type InputObjectTypeDefinition struct {
	Description                 string
	Name                        string
	Directives                  Directives
	InputObjectValueDefinitions InputObjectValueDefinitions
}

type InputObjectValueDefinitions []*InputValueDefinition
