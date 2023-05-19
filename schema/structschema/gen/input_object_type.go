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

func (c *genCtx) processInputObjectType(typ *types.Named) (*inputObjMeta, error) {
	// Find and parse the meta field that contains partial GraphQL definition
	// of this type
	td := &ast.InputObjectTypeDefinition{}

	structTyp := typ.Underlying().(*types.Struct)
	for i := 0; i < structTyp.NumFields(); i++ {
		f := structTyp.Field(i)
		if isSSInputObjectType(f.Type()) {
			gqlTypeDef, err := parser.ParsePartialInputObjectTypeDefinition(structTyp.Tag(i))
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
		return existing.(*inputObjMeta), nil
	}

	meta := &inputObjMeta{
		baseMeta: baseMeta{
			name:      td.Name,
			namedType: typ,
		},
		GQL: td,
	}
	c.meta[td.Name] = meta

	fieldsByName := make(map[string]*inputFieldMeta)
	for _, f := range flatFields(typ) {
		valueDef := &ast.InputValueDefinition{}

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
					gqlValueDef, err := parser.ParsePartialInputValueDefinition(gql)
					if err != nil {
						return nil, fmt.Errorf("Cannot parse GQL metadata for object %s, field %s: %v", typ.Obj().String(), f.field.Name(), err)
					}
					valueDef = gqlValueDef
				}
			}

			if doc != "" {
				valueDef.Description = doc
			}
		}

		if valueDef.Name == "" {
			valueDef.Name = kace.Camel(f.field.Name())
		}

		refType, err := c.goTypeToSchemaType(f.field.Type())
		if err != nil {
			return nil, err
		}

		if valueDef.Type == nil {
			valueDef.Type = refType
		}

		meta := &inputFieldMeta{
			Name:       valueDef.Name,
			GQL:        valueDef,
			StructName: f.field.Name(),
			Type:       f.field.Type(),
		}
		fieldsByName[valueDef.Name] = meta
	}

	allFields := make([]*inputFieldMeta, 0, len(fieldsByName))
	for _, v := range fieldsByName {
		allFields = append(allFields, v)
	}

	meta.Fields = allFields
	return meta, nil
}
