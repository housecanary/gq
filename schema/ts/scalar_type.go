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
