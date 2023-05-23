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
	"fmt"
	"reflect"
)

// An Interface is the base type to use for defining a GQL interface
type Interface[T any] struct {
	Value      T
	objectType reflect.Type
}

// InterfaceT is a type constraint that matches types derived from Interface
type InterfaceT[T any] interface {
	~struct {
		Value      T
		objectType reflect.Type
	}
}

// An InterfaceType represents a GQL interface type
type InterfaceType[T any] struct {
	def string
}

// NewInterfaceType creates a new InterfaceType
func NewInterfaceType[I InterfaceT[T], T any](mod *Module, def string) *InterfaceType[I] {
	it := &InterfaceType[I]{
		def: def,
	}
	mod.addType(&interfaceTypeBuilder[I, T]{it: it})
	return it
}

// Nil returns a new value of this interface with a nil value
func (it *InterfaceType[I]) Nil() I {
	var i I
	return i
}

// Implements registers the given ObjectType as implementing the given InterfaceType and returns
// a function to create an instance of the interface from the object
func Implements[O any, I InterfaceT[T], T any](ot *ObjectType[O], it *InterfaceType[I]) func(*O) I {
	oTyp := typeOf[*O]()
	iTyp := typeOf[I]()
	tType := typeOf[T]()
	if !oTyp.AssignableTo(tType) {
		panic(fmt.Sprintf("invalid Implements call: %s is not assignable to %s", oTyp.String(), tType.String()))
	}
	ot.implements = append(ot.implements, iTyp)
	return func(o *O) I {
		return I{any(o).(T), oTyp}
	}
}
