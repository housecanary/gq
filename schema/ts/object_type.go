package ts

import (
	"reflect"
)

// An ObjectType represents a GQL object type
type ObjectType[O any] struct {
	def        string
	fields     []*FieldType[O]
	implements []reflect.Type
}

// Object creates a new ObjectType. See example for full details.
func Object[O any](mod *ModuleType, def string, fields ...*FieldType[O]) *ObjectType[O] {
	ot := &ObjectType[O]{
		def:    def,
		fields: fields,
	}
	mod.addType(&objectTypeBuilder[O]{ot: ot})
	return ot
}

// NewInstance makes a new instance of the struct backing this ObjectType
func (ot *ObjectType[O]) NewInstance() *O {
	var o O
	return &o
}
