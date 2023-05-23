// Copyright 2023 HouseCanary, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ts

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser"
	"github.com/housecanary/gq/schema"
)

type objectTypeBuilder[O any] struct {
	ot  *ObjectType[O]
	def *ast.BasicTypeDefinition
}

func (b *objectTypeBuilder[O]) describe() string {
	typ := typeOf[O]()
	return fmt.Sprintf("object %s", typeDesc(typ))
}

func (b *objectTypeBuilder[O]) parse(namePrefix string) (*gqlTypeInfo, reflect.Type, error) {
	return parseTypeDef[O, *O](kindObject, b.ot.def, namePrefix, &b.def)
}

func (b *objectTypeBuilder[O]) build(c *buildContext, sb *schema.Builder) error {
	typ := typeOf[O]()
	tb := sb.AddObjectType(b.def.Name)
	setSchemaElementProps(tb, b.def.Description, b.def.Directives)

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
		if !field.IsExported() {
			continue
		}
		def, typeInferred, err := parseStructField(c, field, parser.ParsePartialFieldDefinition)
		if err != nil {
			spew.Dump(typ, field, field.Index)
			return fmt.Errorf("error processing field %s: %w", field.Name, err)
		}
		if def == nil {
			continue
		}
		fields[def.Name] = append(fields[def.Name], &fieldInfo{def, typeInferred, len(field.Index) * 2, field})
	}

	// Next: fields registered on embedded structs, priority comes from depth in tree
	for _, field := range reflect.VisibleFields(typ) {
		if field.Anonymous {
			isPtr := false
			var fieldBuilder builderType
			if field.Type.Kind() == reflect.Struct {
				if bt, ok := c.goTypeToBuilder[reflect.PointerTo(field.Type)]; ok {
					fieldBuilder = bt
				}
			} else if bt, ok := c.goTypeToBuilder[field.Type]; ok {
				fieldBuilder = bt
				isPtr = true
			}

			if fieldBuilder != nil {
				if oft, ok := fieldBuilder.(hasObjectFieldTypes); ok {
					for _, fieldType := range oft.objectFieldTypesForEmbeddedField(getTypedPointerToOffset[O](field.Offset, isPtr)) {
						def, typeInferred, err := fieldType.buildFieldDef(c)
						if err != nil {
							return fmt.Errorf("error processing field %s of embedded field %s: %w", fieldType.originalDefinition(), field.Name, err)
						}

						priority := len(field.Index)*2 + 1

						// Ensure that we respect any fields with a higher precedence
						if existing, hasExisting := fields[def.Name]; hasExisting {
							if existing[len(existing)-1].priority < priority {
								continue
							}
						}
						fields[def.Name] = append(fields[def.Name], &fieldInfo{def, typeInferred, priority, fieldType})
					}
				}
			}
		}
	}

	// Next priority: fields explicitly registered
	for _, fieldType := range b.ot.fields {
		def, typeInferred, err := fieldType.buildFieldDef(c)
		if err != nil {
			return fmt.Errorf("error processing field %s: %w", fieldType.originalDefinition(), err)
		}
		fields[def.Name] = append(fields[def.Name], &fieldInfo{def, typeInferred, 1, fieldType})
	}

	for _, implT := range b.ot.implements {
		if hitd, ok := c.goTypeToBuilder[implT].(hasInterfaceTypeDefinition); ok {
			for _, interfaceField := range hitd.getInterfaceDefinition().FieldsDefinition {
				objectFieldInfos := fields[interfaceField.Name]
				if objectFieldInfos == nil {
					return fmt.Errorf("field %s from interface %s does not exist", interfaceField.Name, hitd.getInterfaceDefinition().Name)
				}

				objectField, _ := collapseFieldInfos(interfaceField.Name, objectFieldInfos)
				if objectField.Type.Signature() != interfaceField.Type.Signature() {
					return fmt.Errorf("field %s from interface %s has incompatible type: got %s, wanted %s", interfaceField.Name, hitd.getInterfaceDefinition().Name, objectField.Type.Signature(), interfaceField.Type.Signature())
				}

				if len(interfaceField.ArgumentsDefinition) != len(objectField.ArgumentsDefinition) {
					return fmt.Errorf("field %s from interface %s has incompatible type: different number of arguments", interfaceField.Name, hitd.getInterfaceDefinition().Name)
				}

				for i := range interfaceField.ArgumentsDefinition {
					intfArg, objectArg := interfaceField.ArgumentsDefinition[i], objectField.ArgumentsDefinition[i]
					if intfArg.Name != objectArg.Name {
						return fmt.Errorf("field %s from interface %s has incompatible type: argument %d has a different name", interfaceField.Name, hitd.getInterfaceDefinition().Name, i)
					}

					if intfArg.Type != objectArg.Type {
						return fmt.Errorf("field %s from interface %s has incompatible type: argument %d has a different type", interfaceField.Name, hitd.getInterfaceDefinition().Name, i)
					}
				}
			}
		}
	}

	for name, defs := range fields {
		if err := b.buildField(c, tb, name, defs); err != nil {
			return fmt.Errorf("error processing field %s: %w", name, err)
		}
	}

	return nil
}

func (otb *objectTypeBuilder[O]) objectFieldTypesForEmbeddedField(getPointer func(v any) unsafe.Pointer) []objectFieldType {
	var result []objectFieldType
	for _, field := range otb.ot.fields {
		if rf, ok := field.(resolverFactory); ok {
			result = append(result, &embeddedResolverFactory[O]{
				objectFieldType: field,
				delegate:        rf,
				getPointer:      getPointer,
			})
		}
	}
	return result
}

func getTypedPointerToOffset[T any](offset uintptr, fieldIsPtr bool) func(any) unsafe.Pointer {
	return func(v any) unsafe.Pointer {
		if fieldIsPtr {
			pp := (*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(v.(*T)), offset))
			return *pp
		}
		return unsafe.Add(unsafe.Pointer(v.(*T)), offset)
	}
}

type hasObjectFieldTypes interface {
	objectFieldTypesForEmbeddedField(getPointer func(v any) unsafe.Pointer) []objectFieldType
}

type hasInterfaceTypeDefinition interface {
	getInterfaceDefinition() *ast.InterfaceTypeDefinition
}

type fieldInfo struct {
	def          *ast.FieldDefinition
	typeInferred bool
	priority     int
	source       interface{}
}

type resolverFactory interface {
	makeResolver(c *buildContext) (schema.Resolver, fieldInvoker, error)
}

type embeddedResolverFactory[T any] struct {
	objectFieldType
	delegate   resolverFactory
	getPointer func(v any) unsafe.Pointer
}

func (e *embeddedResolverFactory[T]) makeResolver(c *buildContext) (schema.Resolver, fieldInvoker, error) {
	dr, di, err := e.delegate.makeResolver(c)
	if err != nil {
		return nil, nil, err
	}

	resolver := schema.FullResolver(func(ctx schema.ResolverContext, v interface{}) (interface{}, error) {
		return dr.Resolve(ctx, (*T)(e.getPointer(v)))
	})

	invoker := func(q QueryInfo, o interface{}) interface{} {
		return di(q, (*T)(e.getPointer(o)))
	}

	return resolver, invoker, nil
}

func (b *objectTypeBuilder[O]) buildField(c *buildContext, tb *schema.ObjectTypeBuilder, name string, infos []*fieldInfo) error {
	fd, source := collapseFieldInfos(name, infos)
	var resolver schema.Resolver
	switch source := source.(type) {
	case reflect.StructField:
		var getValue func(v interface{}) interface{}

		if source.Type.Kind() == reflect.Struct {
			if gqlType, ok := c.goTypeToSchemaType[reflect.PointerTo(source.Type)]; ok && gqlType.kind == kindObject {
				getValue = func(v interface{}) interface{} {
					tv := v.(*O)
					if tv == nil {
						return nil
					}

					rv := reflect.ValueOf(tv)
					return rv.Elem().FieldByIndex(source.Index).Addr().Interface()
				}
			}
		}

		if getValue == nil {
			getValue = func(v interface{}) interface{} {
				tv := v.(*O)
				if tv == nil {
					return nil
				}

				rv := reflect.ValueOf(tv)
				return rv.Elem().FieldByIndex(source.Index).Interface()
			}
		}

		resolver = schema.SimpleResolver(func(v interface{}) (interface{}, error) {
			return getValue(v), nil
		})
		c.registerObjectField(typeOf[*O](), name, &fieldRuntimeInfo{
			sourceField: source,
			invoker: func(q QueryInfo, o interface{}) interface{} {
				return getValue(o)
			},
		})
	case resolverFactory:
		r, i, err := source.makeResolver(c)
		if err != nil {
			return err
		}
		resolver = r
		c.registerObjectField(typeOf[*O](), name, &fieldRuntimeInfo{
			invoker: i,
		})
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

func collapseFieldInfos(name string, infos []*fieldInfo) (*ast.FieldDefinition, any) {
	var fd ast.FieldDefinition
	fd.Name = name
	var source any

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

	return &fd, source
}
