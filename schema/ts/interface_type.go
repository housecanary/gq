package ts

import (
	"reflect"
)

// An Interface is the base type to use for defining a GQL interface
type Interface struct {
	interfaceElement interface{}
	objectType       reflect.Type
}

// InterfaceT is a type constraint that matches types derived from Interface
type InterfaceT interface {
	~struct {
		interfaceElement interface{}
		objectType       reflect.Type
	}
}

// An InterfaceType represents a GQL interface type
type InterfaceType[I InterfaceT] struct {
	def string
}

// NewInterfaceType creates a new InterfaceType
func NewInterfaceType[I InterfaceT](mod *Module, def string) *InterfaceType[I] {
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
func Implements[O any, I InterfaceT](ot *ObjectType[O], it *InterfaceType[I]) func(*O) I {
	oTyp := typeOf[*O]()
	iTyp := typeOf[I]()
	ot.implements = append(ot.implements, iTyp)
	return func(o *O) I {
		return I{o, oTyp}
	}
}
