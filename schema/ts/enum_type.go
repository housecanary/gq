package ts

import (
	"context"
	"fmt"
	"reflect"

	"github.com/codemodus/kace"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser"
	"github.com/housecanary/gq/schema"
)

// An EnumType represents the GQL type of an enum created from Go structs
type EnumType[E ~string] struct {
	def       string
	valueDefs []string
}

// Enum creates an EnumType and registers it with the given module
func Enum[E ~string, P any](mod *ModuleType[P], def string) *EnumType[E] {
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

type enumTypeBuilder[E ~string] struct {
	et  *EnumType[E]
	def *ast.EnumTypeDefinition
}

func (b *enumTypeBuilder[E]) describe() string {
	typ := typeOf[E]()
	return fmt.Sprintf("enum %s", typeDesc(typ))
}

func (b *enumTypeBuilder[E]) parse(namePrefix string) (string, reflect.Type, error) {
	typ := typeOf[E]()

	typeDef, err := parser.ParsePartialEnumTypeDefinition(b.et.def)
	if err != nil {
		return "", nil, err
	}

	registeredValues := make(map[string]bool)
	for _, typeValueDef := range typeDef.EnumValueDefinitions {
		registeredValues[typeValueDef.Value] = true
	}
	for _, v := range b.et.valueDefs {
		valueDef, err := parser.ParseEnumValueDefinition(v)
		if err != nil {
			return "", nil, fmt.Errorf("invalid enum value definition %s: %w", v, err)
		}
		if registeredValues[valueDef.Value] {
			return "", nil, fmt.Errorf("duplicate enum value definition for value %s", valueDef.Value)
		}
		typeDef.EnumValueDefinitions = append(typeDef.EnumValueDefinitions, valueDef)
	}

	name := typeDef.Name
	if name == "" {
		name = kace.Pascal(typ.Name())
	}
	name = namePrefix + name
	typeDef.Name = name

	b.def = typeDef

	return name, typ, nil
}

func (b *enumTypeBuilder[E]) build(c *buildContext, sb *schema.Builder) error {
	etb := sb.AddEnumType(
		b.def.Name,
		func(ctx context.Context, v interface{}) (schema.LiteralValue, error) {
			ev := v.(E)
			return schema.LiteralString(ev), nil
		},
		func(ctx context.Context, v schema.LiteralValue) (interface{}, error) {
			if sv, ok := v.(schema.LiteralString); ok {
				return E(sv), nil
			}
			return nil, fmt.Errorf("invalid enum value: %v", v)
		},
		reflectionInputListCreator{typeOf[E]()},
	)

	setSchemaElementProps(etb, b.def.Description, b.def.Directives)

	for _, v := range b.def.EnumValueDefinitions {
		vb := etb.AddValue(v.Value)
		setSchemaElementProps(vb, v.Description, v.Directives)
	}

	return nil
}
