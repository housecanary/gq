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

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

var schemaMetaType = reflect.TypeOf((*Meta)(nil)).Elem()
var schemaEnumType = reflect.TypeOf((*Enum)(nil)).Elem()
var schemaInputObjectType = reflect.TypeOf((*InputObject)(nil)).Elem()
var scalarMarshalerType = reflect.TypeOf((*schema.ScalarMarshaler)(nil)).Elem()
var scalarUnmarshalerType = reflect.TypeOf((*schema.ScalarUnmarshaler)(nil)).Elem()
var resolverContextType = reflect.TypeOf((*schema.ResolverContext)(nil)).Elem()
var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
var errorType = reflect.TypeOf((*error)(nil)).Elem()

type typeKind int

const (
	typeKindObject typeKind = iota
	typeKindInterface
	typeKindUnion
	typeKindEnum
	typeKindScalar
	typeKindInputObject
)

type typeMeta struct {
	Name        string
	Kind        typeKind
	ReflectType reflect.Type
	GqlType     interface{}
	Fields      []*fieldMeta
	InputFields []*inputFieldMeta
}

type fieldMeta struct {
	Name     string
	GqlField *ast.FieldDefinition
	Resolver schema.Resolver
}

type inputFieldMeta struct {
	Name               string
	StructName         string
	GqlValueDefinition *ast.InputValueDefinition
}

type validatable interface {
	Validate() error
}

// Merges two field definitions
func mergeFieldDef(target, source *ast.FieldDefinition) {
	if target.Description == "" {
		target.Description = source.Description
	}
	if target.Name == "" {
		target.Name = source.Name
	}
	if target.ArgumentsDefinition == nil {
		target.ArgumentsDefinition = source.ArgumentsDefinition
	}
	if target.Type == nil {
		target.Type = source.Type
	}
	if target.Directives == nil {
		target.Directives = source.Directives
	}
}

func isInterfaceType(typ reflect.Type) bool {
	if typ.Kind() == reflect.Struct {
		if f, ok := typ.FieldByName("Interface"); ok && f.Type.Kind() == reflect.Interface && typ.NumField() == 1 {
			return true
		}
	}

	return false
}

func isUnionType(typ reflect.Type) bool {
	if typ.Kind() == reflect.Struct {
		if f, ok := typ.FieldByName("Union"); ok && f.Type.Kind() == reflect.Interface && typ.NumField() == 1 {
			return true
		}
	}

	return false
}

func isEnumType(typ reflect.Type) bool {
	if typ.Kind() == reflect.Struct {
		if f, ok := typ.FieldByName("Enum"); ok && f.Type.AssignableTo(schemaEnumType) {
			return true
		}
	}

	return false
}

func isInputObjectType(typ reflect.Type) bool {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() == reflect.Struct {
		if f, ok := typ.FieldByName("InputObject"); ok && f.Type.AssignableTo(schemaInputObjectType) {
			return true
		}
	}

	return false
}

func isScalarType(typ reflect.Type) bool {
	if typ.AssignableTo(scalarMarshalerType) || reflect.PtrTo(typ).AssignableTo(scalarUnmarshalerType) {
		return true
	}

	return false
}

func implementsInterface(objMeta, intfMeta *typeMeta) bool {
	f, _ := intfMeta.ReflectType.FieldByName("Interface")
	intfType := f.Type
	return objMeta.ReflectType.AssignableTo(intfType)
}

func implementsUnion(objMeta, unionMeta *typeMeta) bool {
	f, _ := unionMeta.ReflectType.FieldByName("Union")
	intfType := f.Type
	return objMeta.ReflectType.AssignableTo(intfType)
}

func makeEncodeScalar(typ reflect.Type) schema.EncodeScalar {
	if typ.AssignableTo(scalarMarshalerType) {
		return schema.EncodeScalarMarshaler
	}

	// Should not be reachable
	panic(fmt.Errorf("Cannot make scalar encode function for type %v", typ))
}

func makeDecodeScalar(typ reflect.Type) schema.DecodeScalar {
	if reflect.PtrTo(typ).AssignableTo(scalarUnmarshalerType) {
		return func(ctx context.Context, in schema.LiteralValue) (interface{}, error) {
			val := reflect.New(typ)
			intf := val.Interface().(schema.ScalarUnmarshaler)
			err := intf.FromLiteralValue(in)
			if err != nil {
				return nil, err
			}
			return val.Elem().Interface(), nil
		}
	}

	// Should not be reachable
	panic(fmt.Errorf("Cannot make scalar decode function for type %v", typ))
}

func makeDecodeInputObject(meta *typeMeta) schema.DecodeInputObject {
	type fieldInfo struct {
		gqlName string
		index   []int
	}
	typ := meta.ReflectType
	fields := make([]fieldInfo, len(meta.InputFields))
	for i, e := range meta.InputFields {
		f, ok := typ.FieldByName(e.StructName)
		if !ok {
			// Should be unreachable
			panic(fmt.Errorf("Previously resolved field %s not found", e.Name))
		}
		fields[i] = fieldInfo{
			e.Name,
			f.Index,
		}
	}

	return func(ctx schema.InputObjectDecodeContext) (interface{}, error) {
		if ctx.IsNil() {
			return reflect.Zero(reflect.PtrTo(typ)).Interface(), nil
		}
		val := reflect.New(typ)
		elem := val.Elem()
		for _, f := range fields {
			fieldVal, err := ctx.GetFieldValue(f.gqlName)
			if err != nil {
				return nil, err
			}

			field := elem.FieldByIndex(f.index)
			fv := reflect.ValueOf(fieldVal)
			if fieldVal == nil {
				fv = reflect.Zero(field.Type())
			}
			field.Set(fv)
		}

		obj := val.Interface()
		if v, ok := obj.(validatable); ok {
			if err := v.Validate(); err != nil {
				return nil, err
			}
		}
		return obj, nil
	}
}

func makeEncodeEnum(typ reflect.Type) schema.EncodeEnum {
	return func(ctx context.Context, v interface{}) (schema.LiteralValue, error) {
		enm := v.(EnumValue)
		if !enm.Nil() {
			return schema.LiteralString(enm.String()), nil
		}
		return nil, nil
	}
}

func makeDecodeEnum(typ reflect.Type) schema.DecodeEnum {
	f, _ := typ.FieldByName("Enum")
	return func(ctx context.Context, v schema.LiteralValue) (interface{}, error) {
		if v == nil {
			rv := reflect.New(typ).Elem()
			rv.FieldByIndex(f.Index).SetString("")
			return rv.Interface(), nil
		}
		s, ok := v.(schema.LiteralString)
		if !ok {
			return nil, fmt.Errorf("Invalid enum input: %v is not an string", v)
		}
		rv := reflect.New(typ).Elem()
		rv.FieldByIndex(f.Index).SetString(string(s))
		return rv.Interface(), nil
	}
}

type reflectionInputListCreator struct {
	typ reflect.Type
}

func (r reflectionInputListCreator) NewList(size int, get func(i int) (interface{}, error)) (interface{}, error) {
	lst := reflect.MakeSlice(reflect.SliceOf(r.typ), size, size)
	for i := 0; i < size; i++ {
		v, err := get(i)
		if err != nil {
			return nil, err
		}
		lst.Index(i).Set(reflect.ValueOf(v))
	}
	return lst.Interface(), nil
}

func (r reflectionInputListCreator) Creator() schema.InputListCreator {
	return reflectionInputListCreator{reflect.SliceOf(r.typ)}
}
