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

type ScalarUnmarshaller[S schema.ScalarMarshaler] interface {
	schema.ScalarUnmarshaler
	*S
}

type ScalarType[S schema.ScalarMarshaler, PS ScalarUnmarshaller[S]] struct {
	def string
}

func Scalar[S schema.ScalarMarshaler, PS ScalarUnmarshaller[S], P any](mod *ModuleType[P], def string) *ScalarType[S, PS] {
	st := &ScalarType[S, PS]{
		def: def,
	}

	mod.addType(&scalarTypeBuilder[S, PS]{st: st})

	return st
}

type scalarTypeBuilder[S schema.ScalarMarshaler, PS ScalarUnmarshaller[S]] struct {
	st  *ScalarType[S, PS]
	def *ast.ScalarTypeDefinition
}

func (b *scalarTypeBuilder[S, PS]) describe() string {
	typ := typeOf[S]()
	return fmt.Sprintf("scalar %s", typeDesc(typ))
}

func (b *scalarTypeBuilder[S, PS]) parse(namePrefix string) (string, reflect.Type, error) {
	typ := typeOf[S]()

	typeDef, err := parser.ParsePartialScalarTypeDefinition(b.st.def)
	if err != nil {
		return "", nil, err
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

func (b *scalarTypeBuilder[S, PS]) build(c *buildContext, sb *schema.Builder) error {
	tb := sb.AddScalarType(
		b.def.Name,
		func(ctx context.Context, v interface{}) (schema.LiteralValue, error) {
			sv := v.(S)
			return sv.ToLiteralValue()
		},
		func(ctx context.Context, v schema.LiteralValue) (interface{}, error) {
			var sv S
			err := PS(&sv).FromLiteralValue(v)
			return sv, err
		},
		reflectionInputListCreator{typeOf[S]()},
	)

	setSchemaElementProps(tb, b.def.Description, b.def.Directives)
	return nil
}
