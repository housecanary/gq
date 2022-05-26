package ts

import (
	"context"
	"fmt"
	"reflect"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

type unionTypeBuilder[U UnionBoxT] struct {
	ut  *UnionType[U]
	def *ast.BasicTypeDefinition
}

func (b *unionTypeBuilder[U]) describe() string {
	typ := typeOf[U]()
	return fmt.Sprintf("union %s", typeDesc(typ))
}

func (b *unionTypeBuilder[U]) parse(namePrefix string) (*gqlTypeInfo, reflect.Type, error) {
	return parseTypeDef[U, U](kindUnion, b.ut.def, namePrefix, &b.def)
}

func (b *unionTypeBuilder[U]) build(c *buildContext, sb *schema.Builder) error {
	typeNameMap := make(map[reflect.Type]string)
	var members []string
	for _, t := range b.ut.members {
		st, err := c.astTypeForGoType(t)
		if err != nil {
			return err
		}
		typeNameMap[t] = st.Signature()
		members = append(members, st.Signature())
	}

	tb := sb.AddUnionType(b.def.Name, members, func(ctx context.Context, value interface{}) (interface{}, string) {
		ub := (UnionBox)(value.(U))
		if ub.objectType == nil {
			return nil, ""
		}
		return ub.unionElement, typeNameMap[ub.objectType]
	})
	setSchemaElementProps(tb, b.def.Description, b.def.Directives)
	return nil
}
