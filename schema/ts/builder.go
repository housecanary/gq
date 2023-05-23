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

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

type builderType interface {
	describe() string
	parse(namePrefix string) (*gqlTypeInfo, reflect.Type, error)
	build(c *buildContext, sb *schema.Builder) error
}

type buildContext struct {
	goTypeToSchemaType map[reflect.Type]*gqlTypeInfo
	goTypeToBuilder    map[reflect.Type]builderType
	implements         map[string]map[reflect.Type]string
	objectTypes        map[reflect.Type]map[string]*fieldRuntimeInfo
	providers          map[reflect.Type]func(QueryInfo) any
}

func newBuildContext(providers map[reflect.Type]func(QueryInfo) any) *buildContext {
	return &buildContext{
		make(map[reflect.Type]*gqlTypeInfo),
		make(map[reflect.Type]builderType),
		make(map[string]map[reflect.Type]string),
		make(map[reflect.Type]map[string]*fieldRuntimeInfo),
		providers,
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
	gqlType, ok := c.goTypeToSchemaType[typ]
	if !ok {
		return nil, fmt.Errorf("type %s not registered", typeDesc(typ))
	}
	return gqlType.astType, nil
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

func (c *buildContext) registerObjectField(oTyp reflect.Type, fName string, fi *fieldRuntimeInfo) {
	oFields, ok := c.objectTypes[oTyp]
	if !ok {
		oFields = make(map[string]*fieldRuntimeInfo)
		c.objectTypes[oTyp] = oFields
	}

	oFields[fName] = fi
}

type gqlTypeKind int

const (
	kindEnum gqlTypeKind = iota + 1
	kindInputObject
	kindInterface
	kindObject
	kindScalar
	kindUnion
)

type gqlTypeInfo struct {
	astType ast.Type
	kind    gqlTypeKind
}
