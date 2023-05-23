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
	"github.com/housecanary/gq/schema"
)

// ScalarUnmarshaller is an interface used as a type constraint for a
// pointer to an object that implements schema.ScalarMarshaler
type ScalarUnmarshaller[S schema.ScalarMarshaler] interface {
	schema.ScalarUnmarshaler
	*S
}

// A ScalarType represents a new GQL scalar type
type ScalarType[S schema.ScalarMarshaler, PS ScalarUnmarshaller[S]] struct {
	def string
}

// NewScalarType constructs a scalar type
func NewScalarType[S schema.ScalarMarshaler, PS ScalarUnmarshaller[S]](mod *Module, def string) *ScalarType[S, PS] {
	st := &ScalarType[S, PS]{
		def: def,
	}

	mod.addType(&scalarTypeBuilder[S, PS]{st: st})

	return st
}
