// Copyright 2023 HouseCanary, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ts

// A Module represents a group of related types from which a GQL schema can be constructed.
type Module struct {
	typePrefix string
	elements   []builderType
}

// NewModule creates a new Module
func NewModule() *Module {
	return &Module{}
}

// WithPrefix prefixes all types registered in this module with the given prefix.
// This is useful when using multiple modules as a form of namespacing
func (mt *Module) WithPrefix(prefix string) *Module {
	return &Module{
		typePrefix: prefix,
	}
}

func (mt *Module) addType(bt builderType) {
	mt.elements = append(mt.elements, bt)
}
