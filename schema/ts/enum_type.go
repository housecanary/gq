// Copyright 2023 HouseCanary, Inc.
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

package ts

import (
	"github.com/housecanary/gq/internal/pkg/parser"
)

type Enum[T any] struct {
	value T
}

// An EnumType represents the GQL type of an enum created from Go structs
type EnumType[E ~string] struct {
	def       string
	valueDefs []string
}

// NewEnumType creates an EnumType and registers it with the given module
func NewEnumType[E ~string](mod *Module, def string) *EnumType[E] {
	et := &EnumType[E]{
		def: def,
	}
	mod.addType(&enumTypeBuilder[E]{et: et})
	return et
}

// Value adds a value to the enum type, and returns the string used to
// represent that value.
func (et *EnumType[E]) Value(def string) E {
	enumValueDef, err := parser.ParseEnumValueDefinition(def)
	if err != nil {
		// Note: we just return a dummy value in case of error here - errors
		// should get reported properly when we build the schema later on, and
		// there's no place to report the error here
		return "INVALID_VALUE"
	}
	et.valueDefs = append(et.valueDefs, def)
	return E(enumValueDef.Value)
}
