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

var _ WrappedType = (*ListType)(nil)

// ListType represents a GraphQL list type i.e. [Type]
type ListType struct {
	of Type
}

func (t *ListType) isType() {}

// Unwrap returns the wrapped type (i.e. Type in [Type])
func (t *ListType) Unwrap() Type {
	return t.of
}

// InputListCreator returns a factory of lists of lists of the contained type
func (t *ListType) InputListCreator() InputListCreator {
	return t.of.(InputableType).InputListCreator().Creator()
}

// A ListValueCallback is a function invoked for each member of a list
type ListValueCallback func(interface{})

// A ListValue is an interface implemented by list values
type ListValue interface {
	Len() int
	ForEachElement(ListValueCallback)
}

// ListOf creates a ListValue from a collection of objects
func ListOf(v ...interface{}) ListValue {
	return simpleList(v)
}

type simpleList []interface{}

func (s simpleList) Len() int {
	return len(s)
}

func (s simpleList) ForEachElement(cb ListValueCallback) {
	for _, e := range s {
		cb(e)
	}
}

func (t *ListType) signature() string {
	return "[" + t.of.signature() + "]"
}
