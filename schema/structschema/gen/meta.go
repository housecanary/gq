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

package gen

import (
	"go/types"

	"github.com/housecanary/gq/ast"
)

type typeKind int

const (
	typeKindObject typeKind = iota
	typeKindInterface
	typeKindUnion
	typeKindEnum
	typeKindScalar
	typeKindInputObject
)

type gqlMeta interface {
	Name() string
	Kind() typeKind
	NamedType() *types.Named
}

type sortMetaByName []gqlMeta

func (s sortMetaByName) Len() int {
	return len(s)
}

func (s sortMetaByName) Less(i, j int) bool {
	return s[i].Name() < s[j].Name()
}

func (s sortMetaByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type baseMeta struct {
	name      string
	namedType *types.Named
}

func (m *baseMeta) Name() string {
	return m.name
}

func (m *baseMeta) NamedType() *types.Named {
	return m.namedType
}

type objMeta struct {
	baseMeta
	GQL    *ast.ObjectTypeDefinition
	Fields []*fieldMeta
}

func (m *objMeta) Kind() typeKind {
	return typeKindObject
}

type fieldMeta struct {
	Obj    *objMeta
	Name   string
	GQL    *ast.FieldDefinition
	Method *types.Selection
	Field  *types.Var
}

type sortFieldMetasByName []*fieldMeta

func (s sortFieldMetasByName) Len() int {
	return len(s)
}

func (s sortFieldMetasByName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s sortFieldMetasByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type inputObjMeta struct {
	baseMeta
	GQL    *ast.InputObjectTypeDefinition
	Fields []*inputFieldMeta
}

func (m *inputObjMeta) Kind() typeKind {
	return typeKindInputObject
}

type inputFieldMeta struct {
	Name       string
	StructName string
	GQL        *ast.InputValueDefinition
	Type       types.Type
}

type sortInputFieldMetasByName []*inputFieldMeta

func (s sortInputFieldMetasByName) Len() int {
	return len(s)
}

func (s sortInputFieldMetasByName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s sortInputFieldMetasByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type scalarMeta struct {
	baseMeta
	GQL *ast.ScalarTypeDefinition
}

func (m *scalarMeta) Kind() typeKind {
	return typeKindScalar
}

type enumMeta struct {
	baseMeta
	GQL *ast.EnumTypeDefinition
}

func (m *enumMeta) Kind() typeKind {
	return typeKindEnum
}

type interfaceMeta struct {
	baseMeta
	GQL           *ast.InterfaceTypeDefinition
	InterfaceType types.Type
	OriginalTag   string
}

func (m *interfaceMeta) Kind() typeKind {
	return typeKindInterface
}

type unionMeta struct {
	baseMeta
	GQL         *ast.UnionTypeDefinition
	UnionType   types.Type
	OriginalTag string
}

func (m *unionMeta) Kind() typeKind {
	return typeKindUnion
}
