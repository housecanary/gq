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
	"go/types"
)

func isGoInterface(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Interface:
		return true
	case *types.Named:
		return isGoInterface(t.Underlying())
	default:
		return false
	}
}

func isGoStruct(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Struct:
		return true
	case *types.Named:
		return isGoStruct(t.Underlying())
	default:
		return false
	}
}

func (c *genCtx) isInterfaceType(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Struct:
		if t.NumFields() == 1 {
			f := t.Field(0)
			if f.Embedded() {
				return false
			}

			if f.Name() == "Interface" && isGoInterface(f.Type()) {
				return true
			}
		}
	case *types.Named:
		return c.isInterfaceType(t.Underlying())
	}

	return false
}

func (c *genCtx) isUnionType(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Struct:
		if t.NumFields() == 1 {
			f := t.Field(0)
			if f.Embedded() {
				return false
			}

			if f.Name() == "Union" && isGoInterface(f.Type()) {
				return true
			}
		}
	case *types.Named:
		return c.isUnionType(t.Underlying())
	}

	return false
}

func (c *genCtx) isEnumType(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Struct:
		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)

			if f.Name() == "Enum" {
				if named, ok := f.Type().(*types.Named); ok {
					tn := named.Obj()
					if tn.Name() == "Enum" && tn.Pkg().Path() == "github.com/housecanary/gq/schema/structschema" {
						return true
					}
				}
			}
		}
	case *types.Named:
		return c.isEnumType(t.Underlying())
	}

	return false
}

func (c *genCtx) isInputObjectType(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Struct:
		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)

			if f.Name() == "InputObject" {
				if named, ok := f.Type().(*types.Named); ok {
					tn := named.Obj()
					if tn.Name() == "InputObject" && tn.Pkg().Path() == "github.com/housecanary/gq/schema/structschema" {
						return true
					}
				}
			}
		}
	case *types.Named:
		return c.isInputObjectType(t.Underlying())
	}

	return false
}

func (c *genCtx) isScalarType(typ types.Type) bool {
	sm := c.schemaPkg.Scope().Lookup("ScalarMarshaler")
	sum := c.schemaPkg.Scope().Lookup("ScalarUnmarshaler")
	return types.AssignableTo(typ, sm.Type()) && types.AssignableTo(types.NewPointer(typ), sum.Type())
}

func (c *genCtx) isObjectType(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Named:
		if _, ok := t.Underlying().(*types.Struct); ok {
			return true
		}
	}
	return false
}

func isSSMetaType(typ types.Type) bool {
	if named, ok := typ.(*types.Named); ok {
		tn := named.Obj()
		if tn.Name() == "Meta" && tn.Pkg().Path() == "github.com/housecanary/gq/schema/structschema" {
			return true
		}
	}
	return false
}

func (c *genCtx) isResolverContextType(typ types.Type) bool {
	if named, ok := typ.(*types.Named); ok {
		tn := named.Obj()
		if tn.Name() == "ResolverContext" && tn.Pkg().Path() == "github.com/housecanary/gq/schema" {
			return true
		}
	}
	return false
}

func (c *genCtx) isContextType(typ types.Type) bool {
	if named, ok := typ.(*types.Named); ok {
		tn := named.Obj()
		if tn.Name() == "Context" && tn.Pkg().Path() == "context" {
			return true
		}
	}
	return false
}

func (c *genCtx) isInjectedArg(v *types.Var) bool {
	typ := v.Type()

	if c.isResolverContextType(typ) || c.isContextType(typ) {
		return true
	}

	if pt, ok := typ.(*types.Pointer); ok {
		typ = pt.Elem()
	}

	if c.isInputObjectType(typ) || c.isEnumType(typ) || c.isScalarType(typ) {
		return false
	}

	return true
}

func (c *genCtx) isError(typ types.Type) bool {
	if named, ok := typ.(*types.Named); ok {
		tn := named.Obj()
		return tn.Name() == "error" && tn.Pkg() == nil
	}
	return false
}
