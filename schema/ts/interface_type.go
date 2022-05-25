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

type InterfaceBox struct {
	interfaceElement interface{}
	objectType       reflect.Type
}

type InterfaceBoxT interface {
	~struct {
		interfaceElement interface{}
		objectType       reflect.Type
	}
}

type InterfaceType[I InterfaceBoxT] struct {
	def string
}

func Interface[I InterfaceBoxT, P any](mod *ModuleType[P], def string) *InterfaceType[I] {
	it := &InterfaceType[I]{
		def: def,
	}
	mod.addType(&interfaceTypeBuilder[I]{it: it})
	return it
}

func (it *InterfaceType[I]) Nil() I {
	return I{nil, nil}
}

func Implements[O any, I InterfaceBoxT](ot *ObjectType[O], it *InterfaceType[I]) func(*O) I {
	oTyp := typeOf[*O]()
	iTyp := typeOf[I]()
	ot.implements = append(ot.implements, iTyp)
	return func(o *O) I {
		return I{o, oTyp}
	}
}

type interfaceTypeBuilder[I InterfaceBoxT] struct {
	it  *InterfaceType[I]
	def *ast.InterfaceTypeDefinition
}

func (b *interfaceTypeBuilder[I]) describe() string {
	typ := typeOf[I]()
	return fmt.Sprintf("interface %s", typeDesc(typ))
}

func (b *interfaceTypeBuilder[I]) parse(namePrefix string) (string, reflect.Type, error) {
	typ := typeOf[I]()

	typeDef, err := parser.ParsePartialInterfaceTypeDefinition(b.it.def)
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

func (b *interfaceTypeBuilder[I]) build(c *buildContext, sb *schema.Builder) error {
	typeNameMap := c.getInterfaceImplementationMap(b.def.Name)
	tb := sb.AddInterfaceType(b.def.Name, func(ctx context.Context, value interface{}) (interface{}, string) {
		ib := (InterfaceBox)(value.(I))
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
