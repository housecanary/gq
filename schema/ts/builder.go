package ts

import (
	"fmt"
	"reflect"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

type BuilderModule interface {
	namePrefix() string
	types() []builderType
}

type builderType interface {
	describe() string
	parse(namePrefix string) (string, reflect.Type, error)
	build(c *buildContext, sb *schema.Builder) error
}

// NewSchemaBuilder creates a schema.Builder from a set of ts modules.
func NewSchemaBuilder(mods ...BuilderModule) (*schema.Builder, error) {
	bc := newBuildContext()
	sb := schema.NewBuilder()
	var allTypes []builderType
	for _, mod := range append(mods, BuiltinTypes) {
		types := mod.types()
		for _, bt := range types {
			name, goType, err := bt.parse(mod.namePrefix())
			if err != nil {
				return nil, fmt.Errorf("cannot parse type %s: %w", bt.describe(), err)
			}
			st := &ast.SimpleType{Name: name}
			bc.goTypeToSchemaType[goType] = st
		}
		allTypes = append(allTypes, types...)
	}

	// TODO: should we allow the same type to be registered in multiple modules

	for _, bt := range allTypes {
		err := bt.build(bc, sb)
		if err != nil {
			return nil, fmt.Errorf("cannot build type %s: %w", bt.describe(), err)
		}
	}

	return sb, nil
}

type buildContext struct {
	goTypeToSchemaType map[reflect.Type]ast.Type
	implements         map[string]map[reflect.Type]string
}

func newBuildContext() *buildContext {
	return &buildContext{
		make(map[reflect.Type]ast.Type),
		make(map[string]map[reflect.Type]string),
	}
}

func (c *buildContext) astTypeForGoType(typ reflect.Type) (ast.Type, error) {
	if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		refType, err := c.astTypeForGoType(typ.Elem())
		if err != nil {
			return nil, err
		}
		return &ast.ListType{Of: refType}, nil
	}
	st, ok := c.goTypeToSchemaType[typ]
	if !ok {
		return nil, fmt.Errorf("type %s not registered", typeDesc(typ))
	}
	return st, nil
}

func (c *buildContext) checkTypeCompatible(typ reflect.Type, astType ast.Type) error {
	for astType.Kind() == ast.KindNotNil {
		astType = astType.ContainedType()
	}

	if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		if astType.Kind() != ast.KindList {
			return fmt.Errorf("incompatible types: go type %s is a slice or array, but ast type %s is not a list", typ.String(), astType.Signature())
		}
		return c.checkTypeCompatible(typ.Elem(), astType.ContainedType())
	}

	shouldType, err := c.astTypeForGoType(typ)
	if err != nil {
		return err
	}

	if shouldType.Signature() != astType.Signature() {
		return fmt.Errorf("incompatible types: expected %s, but got %s", shouldType.Signature(), astType.Signature())
	}

	return nil
}

func (c *buildContext) registerImplements(oTyp reflect.Type, oname, iname string) {
	m, ok := c.implements[iname]
	if !ok {
		m = make(map[reflect.Type]string)
		c.implements[iname] = m
	}
	m[oTyp] = oname
}

func (c *buildContext) getInterfaceImplementationMap(iname string) map[reflect.Type]string {
	m, ok := c.implements[iname]
	if !ok {
		m = make(map[reflect.Type]string)
		c.implements[iname] = m
	}
	return m
}
