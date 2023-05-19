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

type interfaceTypeBuilder[I InterfaceT] struct {
	it  *InterfaceType[I]
	def *ast.InterfaceTypeDefinition
}

func (b *interfaceTypeBuilder[I]) describe() string {
	typ := typeOf[I]()
	return fmt.Sprintf("interface %s", typeDesc(typ))
}

func (b *interfaceTypeBuilder[I]) parse(namePrefix string) (*gqlTypeInfo, reflect.Type, error) {
	typ := typeOf[I]()

	typeDef, err := parser.ParsePartialInterfaceTypeDefinition(b.it.def)
	if err != nil {
		return nil, nil, err
	}

	name := typeDef.Name
	if name == "" {
		name = kace.Pascal(typ.Name())
	}
	name = namePrefix + name
	typeDef.Name = name

	b.def = typeDef
	return &gqlTypeInfo{&ast.SimpleType{Name: name}, kindInterface}, typ, nil
}

func (b *interfaceTypeBuilder[I]) build(c *buildContext, sb *schema.Builder) error {
	typeNameMap := c.getInterfaceImplementationMap(b.def.Name)
	tb := sb.AddInterfaceType(b.def.Name, func(ctx context.Context, value interface{}) (interface{}, string) {
		ib := (Interface)(value.(I))
		if ib.objectType == nil {
			return nil, ""
		}
		return ib.interfaceElement, typeNameMap[ib.objectType]
	})
	setSchemaElementProps(tb, b.def.Description, b.def.Directives)
	for _, fd := range b.def.FieldsDefinition {
		fb := tb.AddField(fd.Name, fd.Type)
		setSchemaElementProps(fb, fd.Description, fd.Directives)
		for _, ad := range fd.ArgumentsDefinition {
			ab := fb.AddArgument(ad.Name, ad.Type, ad.DefaultValue)
			setSchemaElementProps(ab, ad.Description, ad.Directives)
		}
	}
	return nil
}
