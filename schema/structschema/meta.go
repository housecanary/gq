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

package structschema

// Meta is a marker field for object types.  It can be used
// to attach a GraphQL object type definition to an object.
// Note that such a definition is entirely optional, and
// many portions of the definition are relaxed where they
// can be derived from metadata.
type Meta struct{}

// InputObject is a marker field for InputObject types.  It
// serves much the same purpose as Meta for object types.
type InputObject struct{}

// Enum is a marker field for enum types.  It also serves as the container
// for an enum value.
//
// Example usage:
//
// 	type Episode struct {
// 		Enum `{
// 			# Released in 1977.
// 			NEWHOPE
//
// 			# Released in 1980.
// 			EMPIRE
//
// 			# Released in 1983.
// 			JEDI
// 		}`
// 	}
type Enum string

func (e Enum) String() string {
	return string(e)
}

// Nil whether this enum value represents nil
func (e Enum) Nil() bool {
	return e == ""
}

func (e Enum) isEnum() {}

// EnumValue is the interface implemented by enums
type EnumValue interface {
	Nil() bool
	String() string
	isEnum()
}
