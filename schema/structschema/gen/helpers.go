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
	"reflect"

	"github.com/housecanary/gq/ast"
)

// Merges two field definitions
func mergeFieldDef(target, source *ast.FieldDefinition) {
	if target.Description == "" {
		target.Description = source.Description
	}
	if target.Name == "" {
		target.Name = source.Name
	}
	if target.ArgumentsDefinition == nil {
		target.ArgumentsDefinition = source.ArgumentsDefinition
	}
	if target.Type == nil {
		target.Type = source.Type
	}
	if target.Directives == nil {
		target.Directives = source.Directives
	}
}

func stripNotNil(typ ast.Type) ast.Type {
	switch t := typ.(type) {
	case *ast.NotNilType:
		return stripNotNil(t.ContainedType())
	case *ast.ListType:
		return &ast.ListType{Of: stripNotNil(t.ContainedType())}
	default:
		return typ
	}
}

func areTypesEqual(a, b ast.Type) bool {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	switch t := a.(type) {
	case *ast.SimpleType:
		return t.Name == (b.(*ast.SimpleType)).Name
	case *ast.ListType:
		return areTypesEqual(t.Of, (b.(*ast.ListType)).Of)
	case *ast.NotNilType:
		return areTypesEqual(t.Of, (b.(*ast.ListType)).Of)
	}

	panic("Unknown type")
}
