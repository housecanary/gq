package ts

import (
	"reflect"
)

// An UnionBox is the base type to use for defining a GQL union
type UnionBox struct {
	unionElement interface{}
	objectType   reflect.Type
}

// UnionBoxT is a type constraint that matches types derived from UnionBox
type UnionBoxT interface {
	~struct {
		unionElement interface{}
		objectType   reflect.Type
	}
}

// A UnionType represents a GQL union type
type UnionType[U UnionBoxT] struct {
	def     string
	members []reflect.Type
}

// Union constructs a UnionType
func Union[U UnionBoxT](mod *ModuleType, def string) *UnionType[U] {
	ut := &UnionType[U]{
		def: def,
	}

	mod.addType(&unionTypeBuilder[U]{ut: ut})
	return ut
}

// Nil returns a new value of this interface with a nil value
func (ut *UnionType[U]) Nil() U {
	return U{nil, nil}
}

// UnionMember adds a member to the specified union, and returns a function used to construct a union value from that type.
func UnionMember[O any, U UnionBoxT](ut *UnionType[U], ot *ObjectType[O]) func(*O) U {
	oTyp := typeOf[*O]()
	ut.members = append(ut.members, oTyp)
	return func(o *O) U {
		return U{o, oTyp}
	}
}
