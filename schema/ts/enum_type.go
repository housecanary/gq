package ts

import (
	"github.com/housecanary/gq/internal/pkg/parser"
)

// An EnumType represents the GQL type of an enum created from Go structs
type EnumType[E ~string] struct {
	def       string
	valueDefs []string
}

// Enum creates an EnumType and registers it with the given module
func Enum[E ~string](mod *ModuleType, def string) *EnumType[E] {
	et := &EnumType[E]{
		def: def,
	}
	mod.addType(&enumTypeBuilder[E]{et: et})
	return et
}

// Value adds a value to the enum type, and returns the string used to
// represent that value.
func (et *EnumType[E]) Value(def string) E {
	enumValueDef, err := parser.ParseEnumValueDefinition(def)
	if err != nil {
		// Note: we just return a dummy value in case of error here - errors
		// should get reported properly when we build the schema later on, and
		// there's no place to report the error here
		return "INVALID_VALUE"
	}
	et.valueDefs = append(et.valueDefs, def)
	return E(enumValueDef.Value)
}
