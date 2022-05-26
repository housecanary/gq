package ts

import (
	"context"
	"fmt"
	"reflect"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

type scalarTypeBuilder[S schema.ScalarMarshaler, PS ScalarUnmarshaller[S]] struct {
	st  *ScalarType[S, PS]
	def *ast.BasicTypeDefinition
}

func (b *scalarTypeBuilder[S, PS]) describe() string {
	typ := typeOf[S]()
	return fmt.Sprintf("scalar %s", typeDesc(typ))
}

func (b *scalarTypeBuilder[S, PS]) parse(namePrefix string) (*gqlTypeInfo, reflect.Type, error) {
	return parseTypeDef[S, S](kindScalar, b.st.def, namePrefix, &b.def)
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
