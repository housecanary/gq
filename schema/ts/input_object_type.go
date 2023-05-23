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

// An InputObjectType represents the GQL type of an input object created from Go structs
type InputObjectType[O any] struct {
	def string
}

// NewInputObjectType creates an InputObjectType and registers it with the given module
func NewInputObjectType[O any](mod *Module, def string) *InputObjectType[O] {
	it := &InputObjectType[O]{
		def: def,
	}
	mod.addType(&inputObjectTypeBuilder[O]{it: it})
	return it
}
