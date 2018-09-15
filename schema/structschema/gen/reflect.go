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
)

type fieldInfo struct {
	field *types.Var
	tag   string
}

type typeQueue []*types.Named

func (q *typeQueue) pop() *types.Named {
	l := len(*q)
	if l == 0 {
		return nil
	}
	v := (*q)[l-1]
	*q = (*q)[0 : l-1]
	return v
}

func (q *typeQueue) push(v *types.Named) {
	*q = append(*q, v)
}

// flatFields returns all public fields defined on the supplied type or embedded types
// that are not schema meta fields
func flatFields(typ *types.Named) []*fieldInfo {
	names := make(map[string]bool)
	queue := typeQueue{typ}
	for ptyp := queue.pop(); ptyp != nil; ptyp = queue.pop() {
		typ := *ptyp
		structTyp := typ.Underlying().(*types.Struct)
		for i := 0; i < structTyp.NumFields(); i++ {
			f := structTyp.Field(i)
			if isSSMetaType(f.Type()) {
				continue
			}
			if !f.Exported() {
				continue
			}
			if f.Embedded() {
				if named, ok := f.Type().(*types.Named); ok {
					if _, ok := named.Underlying().(*types.Struct); ok {
						queue.push(named)
						continue
					}
				}

			} else {
				names[f.Name()] = true
			}
		}
	}

	result := make([]*fieldInfo, 0, len(names))
	structTyp := typ.Underlying().(*types.Struct)
	for k := range names {
		obj, index, _ := types.LookupFieldOrMethod(structTyp, true, typ.Obj().Pkg(), k)
		if v, ok := obj.(*types.Var); ok {
			result = append(result, makeFieldInfo(structTyp, v, index))
		}
	}

	return result
}

// flatEmbeddedFieldsWithMeta finds all embedded fields of typ that have
// a schema meta field attached.
func flatEmbeddedFieldsWithMeta(typ *types.Named) []*fieldInfo {
	names := make(map[string]bool)
	queue := typeQueue{typ}
	for ptyp := queue.pop(); ptyp != nil; ptyp = queue.pop() {
		typ := *ptyp
		structTyp := typ.Underlying().(*types.Struct)
		for i := 0; i < structTyp.NumFields(); i++ {
			f := structTyp.Field(i)
			if isSSMetaType(f.Type()) {
				continue
			}
			if f.Anonymous() && isGoStruct(f.Type()) {
				names[f.Name()] = true
				if named, ok := f.Type().(*types.Named); ok {
					if _, ok := named.Underlying().(*types.Struct); ok {
						queue.push(named)
					}
				}
				continue
			}
		}
	}

	result := make([]*fieldInfo, 0, len(names))
	structTyp := typ.Underlying().(*types.Struct)
	for k := range names {
		obj, _, _ := types.LookupFieldOrMethod(structTyp, true, typ.Obj().Pkg(), k)
		if v, ok := obj.(*types.Var); ok {
			if !v.Anonymous() {
				continue
			}

			if _, ok := v.Type().(*types.Named); !ok {
				continue
			}

			fieldTyp := v.Type().(*types.Named)

			obj, index, _ := types.LookupFieldOrMethod(fieldTyp, true, fieldTyp.Obj().Pkg(), "Meta")
			if obj != nil && isSSMetaType(obj.Type()) {
				result = append(result, makeFieldInfo(structTyp, v, index))
			}
		}
	}

	return result
}

func fieldByName(typ *types.Named, name string) *fieldInfo {
	obj, index, _ := types.LookupFieldOrMethod(typ, true, typ.Obj().Pkg(), name)
	if v, ok := obj.(*types.Var); ok {
		return makeFieldInfo(typ.Underlying().(*types.Struct), v, index)
	}
	return nil
}

func makeFieldInfo(typ *types.Struct, field *types.Var, index []int) *fieldInfo {
	containingType := typ
	for len(index) > 1 {
		i := index[0]
		index = index[1:]
		containingType = toStructType(containingType.Field(i).Type())
	}

	i := index[0]
	return &fieldInfo{field: field, tag: containingType.Tag(i)}
}

func toStructType(typ types.Type) *types.Struct {
	switch t := typ.(type) {
	case *types.Struct:
		return t
	case *types.Named:
		return toStructType(t.Underlying())
	case *types.Pointer:
		return toStructType(t.Elem())
	}
	panic(fmt.Errorf("Cannot translate %v to a struct type", typ))
}
