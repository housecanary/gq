package ts

import (
	"context"
	"fmt"
	"reflect"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser"
	"github.com/housecanary/gq/schema"
)

type enumTypeBuilder[E ~string] struct {
	et        *EnumType[E]
	def       *ast.BasicTypeDefinition
	valueDefs []*ast.EnumValueDefinition
}

func (b *enumTypeBuilder[E]) describe() string {
	typ := typeOf[E]()
	return fmt.Sprintf("enum %s", typeDesc(typ))
}

func (b *enumTypeBuilder[E]) parse(namePrefix string) (*gqlTypeInfo, reflect.Type, error) {
	var valueDefs []*ast.EnumValueDefinition
	for _, v := range b.et.valueDefs {
		valueDef, err := parser.ParseEnumValueDefinition(v)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid enum value definition %s: %w", v, err)
		}
		valueDefs = append(valueDefs, valueDef)
	}

	b.valueDefs = valueDefs
	return parseTypeDef[E, E](kindEnum, b.et.def, namePrefix, &b.def)
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

	for _, v := range b.valueDefs {
		vb := etb.AddValue(v.Value)
		setSchemaElementProps(vb, v.Description, v.Directives)
	}

	return nil
}
