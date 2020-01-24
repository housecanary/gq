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

package gen

import (
	"fmt"
	"go/types"
	"reflect"
	"strings"

	"github.com/codemodus/kace"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser"
)

func (c *genCtx) processObjectType(typ *types.Named) (*objMeta, error) {
	// Find and parse the meta field that contains partial GraphQL definition
	// of this type
	td := &ast.ObjectTypeDefinition{}

	structTyp := typ.Underlying().(*types.Struct)
	methodSet := types.NewMethodSet(types.NewPointer(typ))
	for i := 0; i < structTyp.NumFields(); i++ {
		f := structTyp.Field(i)
		if isSSMetaType(f.Type()) {
			gqlTypeDef, err := parser.ParsePartialObjectTypeDefinition(structTyp.Tag(i))
			if err != nil {
				return nil, fmt.Errorf("Cannot parse GQL metadata for object %s: %v", typ.Obj().Name(), err)
			}
			td = gqlTypeDef
			break
		}
	}

	// Assign name if not defined in GraphQL
	if td.Name == "" {
		td.Name = kace.Pascal(typ.Obj().Name())
	}

	if err := c.checkRegistration(td.Name, typ); err != nil {
		return nil, err
	}

	if existing, ok := c.meta[td.Name]; ok {
		return existing.(*objMeta), nil
	}

	meta := &objMeta{
		baseMeta: baseMeta{
			name:      td.Name,
			namedType: typ,
		},
		GQL: td,
	}
	c.meta[td.Name] = meta

	// Merge in field definition data defined in GQL with data
	// discovered by reflecting over the fields of the struct
	fieldsByName := make(map[string]*fieldMeta)
	for _, fd := range td.FieldsDefinition {
		fieldsByName[fd.Name] = &fieldMeta{
			Obj:  meta,
			Name: fd.Name,
			GQL:  fd,
		}
	}

	for _, f := range flatFields(typ) {
		fieldDef := &ast.FieldDefinition{}
		fieldTag := reflect.StructTag(f.tag)
		if tag, ok := fieldTag.Lookup("gq"); ok {
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
						return nil, fmt.Errorf("Cannot parse GQL metadata for object %s, field %s: %v", typ.Obj().String(), f.field.Name(), err)
					}
					fieldDef = gqlFieldDef
				}
			}

			if doc != "" {
				fieldDef.Description = doc
			}
		}

		if fieldDef.Name == "" {
			fieldDef.Name = kace.Camel(f.field.Name())
		}

		if existingMeta, ok := fieldsByName[fieldDef.Name]; ok {
			mergeFieldDef(fieldDef, existingMeta.GQL)
		}

		refType, err := c.goTypeToSchemaType(f.field.Type())
		if err != nil {
			return nil, err
		}
		if fieldDef.Type == nil {
			fieldDef.Type = refType
		}
		c.validateFieldType(refType, fieldDef)

		meta := &fieldMeta{
			Obj:   meta,
			Name:  fieldDef.Name,
			GQL:   fieldDef,
			Field: f.field,
		}
		fieldsByName[fieldDef.Name] = meta
	}

	allFields := make([]*fieldMeta, 0, len(fieldsByName))
	for k, v := range fieldsByName {
		if v.Field == nil {
			for i := 0; i < methodSet.Len(); i++ {
				fun := methodSet.At(i)
				if !fun.Obj().Exported() {
					continue
				}
				if strings.EqualFold(fun.Obj().Name(), "resolve"+k) {
					if err := c.validateMethod(fun, v.GQL); err != nil {
						return nil, err
					}
					v.Method = fun
				}
			}
		}
		allFields = append(allFields, v)
	}

	// Make sure we add in all the resolver method fields from embedded structs
	for _, f := range flatEmbeddedFieldsWithMeta(typ) {
		fieldTyp := f.field.Type()
		structTyp := toStructType(fieldTyp)
		for i := 0; i < structTyp.NumFields(); i++ {
			f := structTyp.Field(i)
			if isSSMetaType(f.Type()) {
				gqlTypeDef, err := parser.ParsePartialObjectTypeDefinition(structTyp.Tag(i))
				if err != nil {
					return nil, fmt.Errorf("Cannot parse GQL metadata for object %s: %v", fieldTyp.String(), err)
				}
				for _, fd := range gqlTypeDef.FieldsDefinition {
					if _, ok := fieldsByName[fd.Name]; !ok {
						fieldMeta := &fieldMeta{
							Obj:  meta,
							Name: fd.Name,
							GQL:  fd,
						}
						fieldsByName[fd.Name] = fieldMeta

						for i := 0; i < methodSet.Len(); i++ {
							fun := methodSet.At(i)
							if !fun.Obj().Exported() {
								continue
							}
							if strings.EqualFold(fun.Obj().Name(), "resolve"+fd.Name) {
								if err := c.validateMethod(fun, fd); err != nil {
									return nil, err
								}
								fieldMeta.Method = fun
							}
						}
						allFields = append(allFields, fieldMeta)
					}
				}
				break
			}
		}
	}

	meta.Fields = allFields
	return meta, nil
}

func (c *genCtx) validateMethod(fun *types.Selection, fd *ast.FieldDefinition) error {
	sig := fun.Type().(*types.Signature)
	parms := sig.Params()
	defIndex := 0
	for i := 0; i < parms.Len(); i++ {
		parm := parms.At(i)
		if c.isInjectedArg(parm) {
			continue
		}

		parmType, err := c.goTypeToSchemaType(parm.Type())
		if err != nil {
			return err
		}

		if defIndex >= len(fd.ArgumentsDefinition) {
			return fmt.Errorf("Too many arguments to %s", fun.String())
		}

		if err := c.validateParameter(parmType, fd.ArgumentsDefinition[defIndex]); err != nil {
			return err
		}
		defIndex++
	}

	resultType, err := c.unpackResultType(sig.Results())
	if err != nil {
		return err
	}

	if err := c.validateResult(resultType, fd); err != nil {
		return err
	}

	return nil
}

func (c *genCtx) validateFieldType(astType ast.Type, def *ast.FieldDefinition) error {
	if !areTypesEqual(stripNotNil(astType), stripNotNil(def.Type)) {
		return fmt.Errorf("GraphQL field type %s does not match field type %s", def.Type.Signature(), astType.Signature())
	}
	return nil
}

func (c *genCtx) validateParameter(astType ast.Type, def *ast.InputValueDefinition) error {
	if !areTypesEqual(stripNotNil(astType), stripNotNil(def.Type)) {
		return fmt.Errorf("GraphQL param type %s does not match param type %s", def.Type.Signature(), astType.Signature())
	}
	return nil
}

func (c *genCtx) validateResult(astType ast.Type, def *ast.FieldDefinition) error {
	if !areTypesEqual(stripNotNil(astType), stripNotNil(def.Type)) {
		return fmt.Errorf("GraphQL field type %s does not match method return type %s", def.Type.Signature(), astType.Signature())
	}
	return nil
}

func (c *genCtx) unpackResultType(results *types.Tuple) (resultType ast.Type, err error) {
	if results.Len() != 1 && results.Len() != 2 {
		return nil, fmt.Errorf("Invalid return type, expected 1 or 2 values")
	}
	if results.Len() >= 1 {
		r := results.At(0).Type()
		switch t := r.(type) {
		case *types.Signature:
			resultType, err = c.unpackResultType(t.Results())
		case *types.Chan:
			resultType, err = c.goTypeToSchemaType(t.Elem())
		default:
			resultType, err = c.goTypeToSchemaType(t)
		}
	}
	if results.Len() == 2 {
		r := results.At(1).Type()
		switch t := r.(type) {
		case *types.Chan:
			if !c.isError(t.Elem()) {
				return nil, fmt.Errorf("Second return value must be error or chan error")
			}
		default:
			if !c.isError(r) {
				return nil, fmt.Errorf("Second return value must be error or chan error")
			}
		}
	}
	return
}
