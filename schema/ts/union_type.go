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

type UnionBox struct {
	unionElement interface{}
	objectType   reflect.Type
}

type UnionBoxT interface {
	~struct {
		unionElement interface{}
		objectType   reflect.Type
	}
}

type UnionType[U UnionBoxT] struct {
	def     string
	members []reflect.Type
}

func Union[U UnionBoxT, P any](mod *ModuleType[P], def string) *UnionType[U] {
	ut := &UnionType[U]{
		def: def,
	}

	mod.addType(&unionTypeBuilder[U]{ut: ut})
	return ut
}

func UnionMember[O any, U UnionBoxT](ut *UnionType[U], ot *ObjectType[O]) func(*O) U {
	oTyp := typeOf[*O]()
	ut.members = append(ut.members, oTyp)
	return func(o *O) U {
		return U{o, oTyp}
	}
}

type unionTypeBuilder[U UnionBoxT] struct {
	ut  *UnionType[U]
	def *ast.UnionTypeDefinition
}

func (b *unionTypeBuilder[U]) describe() string {
	typ := typeOf[U]()
	return fmt.Sprintf("union %s", typeDesc(typ))
}

func (b *unionTypeBuilder[U]) parse(namePrefix string) (string, reflect.Type, error) {
	typ := typeOf[U]()

	typeDef, err := parser.ParsePartialUnionTypeDefinition(b.ut.def)
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

func (b *unionTypeBuilder[U]) build(c *buildContext, sb *schema.Builder) error {
	typeNameMap := make(map[reflect.Type]string)
	members := b.def.UnionMembership
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
