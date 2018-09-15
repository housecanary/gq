// Copyright 2018 HouseCanary, Inc.
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

package structschema

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/codemodus/kace"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser"
	"github.com/housecanary/gq/schema"
)

// A Builder creates a schema.Schema from a set of annotated struct types.  See
// readme for additional details and examples.
type Builder struct {
	// Types should be a slice of instances or pointer to instances of types
	// to build into the schema. Only roots need to be specified here, all types
	// directly reachable through fields of any type here will be added to the
	// schema. (e.g. if the query type refers to a union, you may need to
	// supply the union members here)
	Types []interface{}

	meta         map[string]*typeMeta
	argProviders map[string]ArgProvider
}

// An ArgProvider can provide a value to be passed to a resolver argument.
type ArgProvider func(context.Context) interface{}

// RegisterArgProvider registers an ArgProvider for the given argument type
func (b *Builder) RegisterArgProvider(argSig string, provider ArgProvider) {
	if b.argProviders == nil {
		b.argProviders = make(map[string]ArgProvider)
	}

	b.argProviders[argSig] = provider
}

func (b *Builder) registerMeta(meta *typeMeta, typ reflect.Type) (*typeMeta, bool, error) {
	// Prevent registration of duplicate named type
	if existing, ok := b.meta[meta.Name]; ok {
		if typ == existing.ReflectType {
			return existing, true, nil
		}
		return nil, true, fmt.Errorf("Trying to add object type with name %s, but already added with a different type", meta.Name)
	}
	b.meta[meta.Name] = meta
	return meta, false, nil
}

func (b *Builder) addInterface(typ reflect.Type) (*typeMeta, error) {
	// Find and parse the meta field that contains partial GraphQL definition
	// of this type
	var err error
	f, _ := typ.FieldByName("Interface")
	gqlTypeDef, err := parser.ParsePartialInterfaceTypeDefinition(string(f.Tag))
	if err != nil {
		return nil, fmt.Errorf("Cannot parse GQL metadata for interface %s: %v", typ.Name(), err)
	}
	td := gqlTypeDef

	// Assign name if not defined in GraphQL
	if td.Name == "" {
		td.Name = kace.Pascal(typ.Name())
	}

	meta, _, err := b.registerMeta(&typeMeta{
		Name:        td.Name,
		Kind:        typeKindInterface,
		ReflectType: typ,
		GqlType:     td,
	}, typ)

	return meta, err
}

func (b *Builder) addUnion(typ reflect.Type) (*typeMeta, error) {
	// Find and parse the meta field that contains partial GraphQL definition
	// of this type
	td := &ast.UnionTypeDefinition{}
	f, _ := typ.FieldByName("Union")
	if f.Tag != "" {
		gqlTypeDef, err := parser.ParsePartialUnionTypeDefinition(string(f.Tag))
		if err != nil {
			return nil, fmt.Errorf("Cannot parse GQL metadata for union %s: %v", typ.Name(), err)
		}
		td = gqlTypeDef
	}

	// Assign name if not defined in GraphQL
	if td.Name == "" {
		td.Name = kace.Pascal(typ.Name())
	}

	meta, _, err := b.registerMeta(&typeMeta{
		Name:        td.Name,
		Kind:        typeKindUnion,
		ReflectType: typ,
		GqlType:     td,
	}, typ)
	return meta, err
}

func (b *Builder) addEnum(typ reflect.Type) (*typeMeta, error) {
	// Find and parse the meta field that contains partial GraphQL definition
	// of this type
	td := &ast.EnumTypeDefinition{}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Type == schemaEnumType {
			gqlTypeDef, err := parser.ParsePartialEnumTypeDefinition(string(f.Tag))
			if err != nil {
				return nil, fmt.Errorf("Cannot parse GQL metadata for enum %s: %v", typ.Name(), err)
			}
			td = gqlTypeDef
			break
		}
	}

	// Assign name if not defined in GraphQL
	if td.Name == "" {
		td.Name = kace.Pascal(typ.Name())
	}

	meta, _, err := b.registerMeta(&typeMeta{
		Name:        td.Name,
		Kind:        typeKindEnum,
		ReflectType: typ,
		GqlType:     td,
	}, typ)

	return meta, err
}

func (b *Builder) addScalar(typ reflect.Type) (*typeMeta, error) {
	td := &ast.ScalarTypeDefinition{}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Type == schemaMetaType {
			gqlTypeDef, err := parser.ParsePartialScalarTypeDefinition(string(f.Tag))
			if err != nil {
				return nil, fmt.Errorf("Cannot parse GQL metadata for scalar %s: %v", typ.Name(), err)
			}
			td = gqlTypeDef
			break
		}
	}

	// Assign name if not defined in GraphQL
	if td.Name == "" {
		td.Name = kace.Pascal(typ.Name())
	}

	meta, _, err := b.registerMeta(&typeMeta{
		Name:        td.Name,
		Kind:        typeKindScalar,
		ReflectType: typ,
		GqlType:     td,
	}, typ)
	return meta, err
}

func (b *Builder) addInputObject(typ reflect.Type) (*typeMeta, error) {
	// Find and parse the meta field that contains partial GraphQL definition
	// of this type
	td := &ast.InputObjectTypeDefinition{}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Type == schemaInputObjectType {
			if f.Tag != "" {
				gqlTypeDef, err := parser.ParsePartialInputObjectTypeDefinition(string(f.Tag))
				if err != nil {
					return nil, fmt.Errorf("Cannot parse GQL metadata for input object %s: %v", typ.Name(), err)
				}
				td = gqlTypeDef
			}
			break
		}
	}

	// Assign name if not defined in GraphQL
	if td.Name == "" {
		td.Name = kace.Pascal(typ.Name())
	}

	objectMeta, exists, err := b.registerMeta(&typeMeta{
		Name:        td.Name,
		Kind:        typeKindInputObject,
		ReflectType: typ,
		GqlType:     td,
	}, typ)
	if err != nil {
		return nil, err
	}

	if exists {
		return objectMeta, err
	}

	fieldsByName := make(map[string]*inputFieldMeta)

	for _, f := range flatFields(typ) {
		valueDef := &ast.InputValueDefinition{}

		if tag, ok := f.Tag.Lookup("gq"); ok {
			tag := strings.TrimSpace(tag)
			parts := strings.SplitN(tag, ";", 2)
			gql := strings.TrimSpace(parts[0])
			doc := ""
			if len(parts) > 1 {
				doc = parts[1]
			}

			if len(gql) > 0 {
				if strings.HasPrefix(gql, "-") {
					continue
				} else {
					gqlValueDef, err := parser.ParsePartialInputValueDefinition(gql)
					if err != nil {
						return nil, fmt.Errorf("Cannot parse GQL metadata for input object %s, field %s: %v", typ.Name(), f.Name, err)
					}
					valueDef = gqlValueDef
				}
			}

			if doc != "" {
				valueDef.Description = doc
			}
		}

		if valueDef.Name == "" {
			valueDef.Name = kace.Camel(f.Name)
		}

		refType, err := b.goTypeToSchemaType(f.Type)
		if err != nil {
			return nil, err
		}
		if valueDef.Type == nil {
			valueDef.Type = refType
		}

		meta := &inputFieldMeta{
			Name:               valueDef.Name,
			StructName:         f.Name,
			GqlValueDefinition: valueDef,
		}
		fieldsByName[valueDef.Name] = meta
	}

	allFields := make([]*inputFieldMeta, 0, len(fieldsByName))
	for _, v := range fieldsByName {
		allFields = append(allFields, v)
	}

	objectMeta.InputFields = allFields
	return objectMeta, nil
}

func (b *Builder) addObject(typ reflect.Type) (*typeMeta, error) {
	// Find and parse the meta field that contains partial GraphQL definition
	// of this type
	td := &ast.ObjectTypeDefinition{}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Type == schemaMetaType {
			gqlTypeDef, err := parser.ParsePartialObjectTypeDefinition(string(f.Tag))
			if err != nil {
				return nil, fmt.Errorf("Cannot parse GQL metadata for object %s: %v", typ.Name(), err)
			}
			td = gqlTypeDef
			break
		}
	}

	// Assign name if not defined in GraphQL
	if td.Name == "" {
		td.Name = kace.Pascal(typ.Name())
	}

	objectMeta, exists, err := b.registerMeta(&typeMeta{
		Name:        td.Name,
		Kind:        typeKindObject,
		ReflectType: typ,
		GqlType:     td,
	}, typ)
	if err != nil {
		return nil, err
	}

	if exists {
		return objectMeta, err
	}

	// Merge in field definition data defined in GQL with data
	// discovered by reflecting over the fields of the struct
	fieldsByName := make(map[string]*fieldMeta)
	for _, fd := range td.FieldsDefinition {
		fieldsByName[fd.Name] = &fieldMeta{
			Name:     fd.Name,
			GqlField: fd,
		}
	}

	for _, f := range flatFields(typ) {
		fieldDef := &ast.FieldDefinition{}

		if tag, ok := f.Tag.Lookup("gq"); ok {
			tag := strings.TrimSpace(tag)
			parts := strings.SplitN(tag, ";", 2)
			gql := strings.TrimSpace(parts[0])
			doc := ""
			if len(parts) > 1 {
				doc = parts[1]
			}

			if len(gql) > 0 {
				if strings.HasPrefix(gql, "-") {
					continue
				} else {
					gqlFieldDef, err := parser.ParsePartialFieldDefinition(gql)
					if err != nil {
						return nil, fmt.Errorf("Cannot parse GQL metadata for object %s, field %s: %v", typ.Name(), f.Name, err)
					}
					fieldDef = gqlFieldDef
				}
			}

			if doc != "" {
				fieldDef.Description = doc
			}
		}

		if fieldDef.Name == "" {
			fieldDef.Name = kace.Camel(f.Name)
		}

		if existingMeta, ok := fieldsByName[fieldDef.Name]; ok {
			mergeFieldDef(fieldDef, existingMeta.GqlField)
		}

		if fieldDef.Type == nil {
			refType, err := b.goTypeToSchemaType(f.Type)
			if err != nil {
				return nil, err
			}
			fieldDef.Type = refType
		}

		meta := &fieldMeta{
			Name:     fieldDef.Name,
			GqlField: fieldDef,
		}
		if err := meta.buildFieldResolver(b, typ, f.Name); err != nil {
			return nil, err
		}
		fieldsByName[fieldDef.Name] = meta
	}

	allFields := make([]*fieldMeta, 0, len(fieldsByName))
	for k, v := range fieldsByName {
		if v.Resolver == nil {
			if err := v.buildMethodResolver(b, typ, k); err != nil {
				return nil, err
			}
		}
		allFields = append(allFields, v)
	}

	// Make sure we add in all the resolver method fields from embedded structs
	for _, f := range flatEmbeddedFieldsWithMeta(typ) {
		metaF, _ := f.Type.FieldByName("Meta")
		gqlTypeDef, err := parser.ParsePartialObjectTypeDefinition(string(metaF.Tag))
		if err != nil {
			return nil, fmt.Errorf("Cannot parse GQL metadata for object %s: %v", metaF.Type.Name(), err)
		}

		for _, fd := range gqlTypeDef.FieldsDefinition {
			if _, ok := fieldsByName[fd.Name]; !ok {
				fieldMeta := &fieldMeta{
					Name:     fd.Name,
					GqlField: fd,
				}
				fieldsByName[fd.Name] = fieldMeta

				if err := fieldMeta.buildMethodResolver(b, typ, fd.Name); err != nil {
					return nil, err
				}
				allFields = append(allFields, fieldMeta)
			}
		}
	}

	objectMeta.Fields = allFields
	return objectMeta, nil
}

func (b *Builder) goTypeToSchemaType(typ reflect.Type) (ast.Type, error) {
	if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		refType, err := b.goTypeToSchemaType(typ.Elem())
		if err != nil {
			return nil, err
		}
		return &ast.ListType{Of: refType}, nil
	} else if typ.Kind() == reflect.Chan {
		refType, err := b.goTypeToSchemaType(typ.Elem())
		if err != nil {
			return nil, err
		}
		return refType, nil
	} else if typ.Kind() == reflect.Ptr {
		typeMeta, err := b.addTypeInfo(typ)
		if err != nil {
			return nil, err
		}
		return &ast.SimpleType{Name: typeMeta.Name}, nil
	} else {
		typeMeta, err := b.addTypeInfo(typ)
		if err != nil {
			return nil, err
		}
		return &ast.SimpleType{Name: typeMeta.Name}, nil
	}
}

func (b *Builder) addTypeInfo(typStruct interface{}) (*typeMeta, error) {
	typ, ok := typStruct.(reflect.Type)
	if !ok {
		typ = reflect.TypeOf(typStruct)
	}
	if isEnumType(typ) {
		return b.addEnum(typ)
	} else if isInputObjectType(typ) {
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		return b.addInputObject(typ)
	} else if isScalarType(typ) {
		return b.addScalar(typ)
	} else {
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		if typ.Kind() != reflect.Struct {
			return nil, fmt.Errorf("Cannot add type %s as a schema element:  expected a struct or pointer to struct, got a %v", typ.Name(), typ.Kind())
		}

		if isInterfaceType(typ) {
			return b.addInterface(typ)
		} else if isUnionType(typ) {
			return b.addUnion(typ)
		} else {
			return b.addObject(typ)
		}
	}
}

func (b *Builder) schemaBuilder() (*schema.Builder, error) {
	b.meta = make(map[string]*typeMeta)
	for _, typ := range b.Types {
		if _, err := b.addTypeInfo(typ); err != nil {
			return nil, err
		}
	}

	schemaBuilder := schema.NewBuilder()
	objectTypes := make([]*typeMeta, 0)
	objectTypeBuilders := make(map[string]*schema.ObjectTypeBuilder, 0)
	interfaceTypes := make([]*typeMeta, 0)
	unionTypes := make([]*typeMeta, 0)
	for _, v := range b.meta {
		switch v.Kind {
		case typeKindEnum:
			etb := schemaBuilder.AddEnumType(v.Name, makeEncodeEnum(v.ReflectType), makeDecodeEnum(v.ReflectType))
			etd := v.GqlType.(*ast.EnumTypeDefinition)
			setSchemaElementProps(etb, etd.Description, etd.Directives)
			for _, valueDef := range etd.EnumValueDefinitions {
				evb := etb.AddValue(valueDef.Value)
				setSchemaElementProps(evb, valueDef.Description, valueDef.Directives)
			}
		case typeKindInterface:
			impls := make(map[reflect.Type]string)
			for _, candidate := range b.meta {
				if candidate.Kind == typeKindObject {
					impls[reflect.PtrTo(candidate.ReflectType)] = candidate.Name
					impls[candidate.ReflectType] = candidate.Name
				}
			}
			f, _ := v.ReflectType.FieldByName("Interface")
			index := f.Index
			itb := schemaBuilder.AddInterfaceType(v.Name, func(ctx context.Context, val interface{}) (interface{}, string) {
				rv := reflect.ValueOf(val).FieldByIndex(index)
				if rv.IsNil() {
					return nil, ""
				}
				rvTyp := reflect.TypeOf(rv.Interface())
				typeName := impls[rvTyp]
				return rv.Interface(), typeName
			})
			itd := v.GqlType.(*ast.InterfaceTypeDefinition)
			setSchemaElementProps(itb, itd.Description, itd.Directives)
			for _, f := range itd.FieldsDefinition {
				fb := itb.AddField(f.Name, f.Type)
				setSchemaElementProps(fb, f.Description, f.Directives)
				for _, a := range f.ArgumentsDefinition {
					ab := fb.AddArgument(a.Name, a.Type, a.DefaultValue)
					setSchemaElementProps(ab, a.Description, a.Directives)
				}
			}
			interfaceTypes = append(interfaceTypes, v)
		case typeKindObject:
			otb := schemaBuilder.AddObjectType(v.Name)
			otd := v.GqlType.(*ast.ObjectTypeDefinition)
			setSchemaElementProps(otb, otd.Description, otd.Directives)
			for _, f := range v.Fields {
				fb := otb.AddField(f.GqlField.Name, f.GqlField.Type, f.Resolver)
				setSchemaElementProps(fb, f.GqlField.Description, f.GqlField.Directives)
				for _, a := range f.GqlField.ArgumentsDefinition {
					ab := fb.AddArgument(a.Name, a.Type, a.DefaultValue)
					setSchemaElementProps(ab, a.Description, a.Directives)
				}
			}

			objectTypeBuilders[v.Name] = otb
			objectTypes = append(objectTypes, v)

		case typeKindScalar:
			std := v.GqlType.(*ast.ScalarTypeDefinition)
			stb := schemaBuilder.AddScalarType(v.Name, makeEncodeScalar(v.ReflectType), makeDecodeScalar(v.ReflectType))
			setSchemaElementProps(stb, std.Description, std.Directives)
		case typeKindUnion:
			unionTypes = append(unionTypes, v)
		case typeKindInputObject:
			iob := schemaBuilder.AddInputObjectType(v.Name, makeDecodeInputObject(v))
			iot := v.GqlType.(*ast.InputObjectTypeDefinition)
			setSchemaElementProps(iob, iot.Description, iot.Directives)
			for _, f := range v.InputFields {
				fb := iob.AddField(f.Name, f.GqlValueDefinition.Type, f.GqlValueDefinition.DefaultValue)
				setSchemaElementProps(fb, f.GqlValueDefinition.Description, f.GqlValueDefinition.Directives)
			}

		default:
			panic("Unknown type kind")
		}
	}
	for _, intfMeta := range interfaceTypes {
		for _, objMeta := range objectTypes {
			if implementsInterface(objMeta, intfMeta) {
				objectTypeBuilders[objMeta.Name].Implements(intfMeta.Name)
			}
		}
	}
	for _, unionMeta := range unionTypes {
		var unionMembers []string
		impls := make(map[reflect.Type]string)
		for _, objMeta := range objectTypes {
			if implementsUnion(objMeta, unionMeta) {
				unionMembers = append(unionMembers, objMeta.Name)
				impls[reflect.PtrTo(objMeta.ReflectType)] = objMeta.Name
				impls[objMeta.ReflectType] = objMeta.Name
			}
		}
		f, _ := unionMeta.ReflectType.FieldByName("Union")
		index := f.Index
		utb := schemaBuilder.AddUnionType(unionMeta.Name, unionMembers, func(ctx context.Context, val interface{}) (interface{}, string) {
			rv := reflect.ValueOf(val).FieldByIndex(index)
			if rv.IsNil() {
				return nil, ""
			}
			rvTyp := reflect.TypeOf(rv.Interface())
			typeName := impls[rvTyp]
			return rv.Interface(), typeName
		})
		utd := unionMeta.GqlType.(*ast.UnionTypeDefinition)
		setSchemaElementProps(utb, utd.Description, utd.Directives)
	}

	return schemaBuilder, nil
}

// SchemaBuilder creates a schema builder from the supplied types
func (b *Builder) SchemaBuilder() (*schema.Builder, error) {
	schemaBuilder, err := b.schemaBuilder()
	return schemaBuilder, err
}

// Build creates a schema from this builder
func (b *Builder) Build(queryTypeName string) (*schema.Schema, error) {
	schemaBuilder, err := b.schemaBuilder()
	if err != nil {
		return nil, err
	}
	return schemaBuilder.Build(queryTypeName)
}

// MustBuild is the same as Build but panics on error
func (b *Builder) MustBuild(queryTypeName string) *schema.Schema {
	s, err := b.Build(queryTypeName)
	if err != nil {
		panic(err)
	}
	return s
}

func setSchemaElementProps(e schema.BuilderSchemaElement, desc string, directives ast.Directives) {
	e.SetDescription(desc)
	for _, d := range directives {
		db := e.AddDirective(d.Name)
		for _, a := range d.Arguments {
			db.AddArgument(a.Name, a.Value)
		}

	}
}
