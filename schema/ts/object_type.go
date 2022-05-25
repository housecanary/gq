package ts

import (
	"fmt"
	"reflect"

	"github.com/codemodus/kace"
	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser"
	"github.com/housecanary/gq/schema"
)

type ObjectType[O any] struct {
	def        string
	fields     []*FieldType[O]
	implements []reflect.Type
}

func Object[O, P any](mod *ModuleType[P], def string, fields ...*FieldType[O]) *ObjectType[O] {
	ot := &ObjectType[O]{
		def:    def,
		fields: fields,
	}
	mod.addType(&objectTypeBuilder[O]{ot: ot})
	return ot
}

type objectTypeBuilder[O any] struct {
	ot  *ObjectType[O]
	def *ast.ObjectTypeDefinition
}

func (b *objectTypeBuilder[O]) describe() string {
	typ := typeOf[O]()
	return fmt.Sprintf("object %s", typeDesc(typ))
}

func (b *objectTypeBuilder[O]) parse(namePrefix string) (string, reflect.Type, error) {
	typ := typeOf[O]()

	if typ.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("Objects must be represented by a struct, not a %v", typ.Kind())
	}

	typeDef, err := parser.ParsePartialObjectTypeDefinition(b.ot.def)
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
	return name, reflect.PointerTo(typ), nil
}

func (b *objectTypeBuilder[O]) build(c *buildContext, sb *schema.Builder) error {
	typ := typeOf[O]()
	dedup := make(map[string]bool)
	tb := sb.AddObjectType(b.def.Name)
	setSchemaElementProps(tb, b.def.Description, b.def.Directives)

	for _, iname := range b.def.ImplementsInterfaces {
		dedup[iname] = true
		tb.Implements(iname)
		c.registerImplements(reflect.PointerTo(typ), b.def.Name, iname)
	}
	for _, implT := range b.ot.implements {
		iname, err := c.astTypeForGoType(implT)
		if err != nil {
			return fmt.Errorf("error in interface declaration %s: %w", implT.Name(), err)
		}
		tb.Implements(iname.Signature())
		c.registerImplements(reflect.PointerTo(typ), b.def.Name, iname.Signature())
	}

	// Collect fields from all sources into a map - we'll merge later
	fields := make(map[string][]*fieldInfo)

	// Lowest priority: fields from struct tags
	for _, field := range reflect.VisibleFields(typ) {
		def, typeInferred, err := parseStructField(c, field, parser.ParsePartialFieldDefinition)
		if err != nil {
			return fmt.Errorf("error processing field %s: %w", field.Name, err)
		}
		if def == nil {
			continue
		}
		fields[def.Name] = append(fields[def.Name], &fieldInfo{def, typeInferred, field})
	}

	// Next priority: fields from the gql definition of the object
	for _, def := range b.def.FieldsDefinition {
		fields[def.Name] = append(fields[def.Name], &fieldInfo{def, false, nil})
	}

	// Next priority: fields explicitly registered
	for _, fieldType := range b.ot.fields {
		def, typeInferred, err := fieldType.buildFieldDef(c)
		if err != nil {
			return fmt.Errorf("error processing field %s: %w", fieldType.def, err)
		}
		fields[def.Name] = append(fields[def.Name], &fieldInfo{def, typeInferred, fieldType})
	}

	for name, defs := range fields {
		if err := b.buildField(c, tb, name, defs); err != nil {
			return fmt.Errorf("error processing field %s: %w", name, err)
		}
	}

	return nil
}

type fieldInfo struct {
	def          *ast.FieldDefinition
	typeInferred bool
	source       interface{}
}

func (b *objectTypeBuilder[O]) buildField(c *buildContext, tb *schema.ObjectTypeBuilder, name string, infos []*fieldInfo) error {
	var fd ast.FieldDefinition
	fd.Name = name
	var source interface{}

	for _, info := range infos {
		if info.def.Description != "" {
			fd.Description = info.def.Description
		}
		if info.def.ArgumentsDefinition != nil {
			fd.ArgumentsDefinition = info.def.ArgumentsDefinition
		}
		if info.def.Directives != nil {
			fd.Directives = info.def.Directives
		}
		if info.def.Type != nil && !info.typeInferred {
			fd.Type = info.def.Type
		}
		if info.source != nil {
			source = info.source
		}
	}

	if fd.Type == nil {
		for _, info := range infos {
			if info.def.Type != nil {
				fd.Type = info.def.Type
			}
		}
	}

	var resolver schema.Resolver
	switch source := source.(type) {
	case reflect.StructField:
		resolver = schema.SimpleResolver(func(v interface{}) (interface{}, error) {
			tv := v.(*O)
			if tv == nil {
				return nil, nil
			}

			rv := reflect.ValueOf(tv)
			return rv.Elem().FieldByIndex(source.Index).Interface(), nil
		})
	case *FieldType[O]:
		r, err := source.makeResolver(c)
		if err != nil {
			return err
		}
		resolver = r
	case nil:
		resolver = schema.SimpleResolver(func(v interface{}) (interface{}, error) {
			return nil, nil
		})
	default:
		panic(fmt.Sprintf("unexpected field source: %v", source))
	}

	fb := tb.AddField(name, fd.Type, resolver)
	setSchemaElementProps(fb, fd.Description, fd.Directives)
	for _, arg := range fd.ArgumentsDefinition {
		ab := fb.AddArgument(arg.Name, arg.Type, arg.DefaultValue)
		setSchemaElementProps(ab, arg.Description, arg.Directives)
	}
	return nil
}
