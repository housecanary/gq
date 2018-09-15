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

	"golang.org/x/tools/go/packages"

	"github.com/housecanary/gq/ast"
)

// Generate runs code generation
func Generate(
	outputPath string,
	outputFileName string,
	outputPackageName string,
	types []string,
	packageNames []string,
) error {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.LoadAllSyntax,
	}, append([]string{
		"github.com/housecanary/gq/schema",
		"github.com/housecanary/gq/schema/structschema",
	}, packageNames...)...)

	if err != nil {
		return err
	}
	schemaPkg, ssPkg := pkgs[0].Types, pkgs[1].Types

	pkgs = pkgs[1:]
	c := &genCtx{
		schemaPkg,
		ssPkg,
		pkgs,
		outputPackageName,
		outputPath,
		outputFileName,
		make(map[string]gqlMeta),
	}

	for _, name := range types {
		fmt.Println("Adding type to schema", name)
		if err := c.addTypeByName(name); err != nil {
			return err
		}
	}

	return c.createOutput()
}

type genCtx struct {
	schemaPkg         *types.Package
	ssPkg             *types.Package
	pkgs              []*packages.Package
	outputPackageName string
	outputPath        string
	outputFileName    string
	meta              map[string]gqlMeta
}

func (c *genCtx) addTypeByName(name string) error {
	for _, pkg := range c.pkgs {
		scope := pkg.Types.Scope()
		obj := scope.Lookup(name)
		if obj == nil {
			continue
		}

		_, err := c.addTypeInfo(obj.Type())
		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("Cannot find type %s in any package", name)
}

func (c *genCtx) checkRegistration(name string, typ *types.Named) error {
	if item, ok := c.meta[name]; ok {
		if item.NamedType().Obj().Id() == typ.Obj().Id() {
			return nil
		}
		return fmt.Errorf("Name %s already registered to type %v", name, item.NamedType())
	}
	return nil
}

func (c *genCtx) goTypeToSchemaType(typ types.Type) (ast.Type, error) {
	switch t := typ.(type) {
	case *types.Array:
		refType, err := c.goTypeToSchemaType(t.Elem())
		if err != nil {
			return nil, err
		}
		return &ast.ListType{Of: refType}, nil
	case *types.Slice:
		refType, err := c.goTypeToSchemaType(t.Elem())
		if err != nil {
			return nil, err
		}
		return &ast.ListType{Of: refType}, nil
	case *types.Chan:
		refType, err := c.goTypeToSchemaType(t.Elem())
		if err != nil {
			return nil, err
		}
		return &ast.ListType{Of: refType}, nil
	case *types.Pointer:
		refType, err := c.goTypeToSchemaType(t.Elem())
		if err != nil {
			return nil, err
		}
		return refType, nil
	default:
		return c.addTypeInfo(typ)
	}
}

func (c *genCtx) addTypeInfo(typ types.Type) (ast.Type, error) {
	if c.isInterfaceType(typ) {
		t, err := c.processInterfaceType(typ.(*types.Named))
		if err != nil {
			return nil, err
		}
		return &ast.SimpleType{Name: t.Name()}, nil
	}
	if c.isUnionType(typ) {
		t, err := c.processUnionType(typ.(*types.Named))
		if err != nil {
			return nil, err
		}
		return &ast.SimpleType{Name: t.Name()}, nil
	}
	if c.isEnumType(typ) {
		et, err := c.processEnumType(typ.(*types.Named))
		if err != nil {
			return nil, err
		}
		return &ast.SimpleType{Name: et.Name()}, nil
	}
	if c.isInputObjectType(typ) {
		iot, err := c.processInputObjectType(typ.(*types.Named))
		if err != nil {
			return nil, err
		}
		return &ast.SimpleType{Name: iot.Name()}, nil
	}
	if c.isScalarType(typ) {
		st, err := c.processScalarType(typ.(*types.Named))
		if err != nil {
			return nil, err
		}
		return &ast.SimpleType{Name: st.Name()}, nil
	}
	if c.isObjectType(typ) {
		ot, err := c.processObjectType(typ.(*types.Named))
		if err != nil {
			return nil, err
		}
		return &ast.SimpleType{Name: ot.Name()}, nil
	}

	return nil, fmt.Errorf("Cannot process type %v", typ)
}
