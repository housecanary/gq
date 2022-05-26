package ts

import (
	"reflect"
)

// An InterfaceBox is the base type to use for defining a GQL interface
type InterfaceBox struct {
	interfaceElement interface{}
	objectType       reflect.Type
}

// InterfaceBoxT is a type constraint that matches types derived from InterfaceBox
type InterfaceBoxT interface {
	~struct {
		interfaceElement interface{}
		objectType       reflect.Type
	}
}

// An InterfaceType represents a GQL interface type
type InterfaceType[I InterfaceBoxT] struct {
	def string
}

// Interface creates a new InterfaceType
func Interface[I InterfaceBoxT](mod *ModuleType, def string) *InterfaceType[I] {
	it := &InterfaceType[I]{
		def: def,
	}
	mod.addType(&interfaceTypeBuilder[I]{it: it})
	return it
}

// Nil returns a new value of this interface with a nil value
func (it *InterfaceType[I]) Nil() I {
	return I{nil, nil}
}

// Implements registers the given ObjectType as implementing the given InterfaceType and returns
// a function to create an instance of the interface from the object
func Implements[O any, I InterfaceBoxT](ot *ObjectType[O], it *InterfaceType[I]) func(*O) I {
	oTyp := typeOf[*O]()
	iTyp := typeOf[I]()
	ot.implements = append(ot.implements, iTyp)
	return func(o *O) I {
		return I{o, oTyp}
	}
}
