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
	"reflect"
	"unicode"
	"unicode/utf8"
)

type typeQueue []reflect.Type

func (q *typeQueue) pop() *reflect.Type {
	l := len(*q)
	if l == 0 {
		return nil
	}
	v := (*q)[l-1]
	*q = (*q)[0 : l-1]
	return &v
}

func (q *typeQueue) push(v reflect.Type) {
	*q = append(*q, v)
}

// flatFields returns all public fields defined on the supplied type or embedded types
// that are not schema meta fields
func flatFields(typ reflect.Type) []reflect.StructField {
	names := make(map[string]bool)
	queue := typeQueue{typ}
	for ptyp := queue.pop(); ptyp != nil; ptyp = queue.pop() {
		typ := *ptyp
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if f.Type == schemaMetaType {
				continue
			}
			r, _ := utf8.DecodeRuneInString(f.Name)
			if unicode.IsLower(r) {
				// Private field
				continue
			}
			if f.Anonymous && f.Type.Kind() == reflect.Struct {
				queue.push(f.Type)
				continue
			} else {
				names[f.Name] = true
			}
		}
	}

	result := make([]reflect.StructField, 0, len(names))
	for k := range names {
		f, ok := typ.FieldByName(k)
		if ok {
			result = append(result, f)
		}
	}

	return result
}

// flatEmbeddedFieldsWithMeta finds all embedded fields of typ that have
// a schema meta field attached.
func flatEmbeddedFieldsWithMeta(typ reflect.Type) []reflect.StructField {
	names := make(map[string]bool)
	queue := typeQueue{typ}
	for ptyp := queue.pop(); ptyp != nil; ptyp = queue.pop() {
		typ := *ptyp
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if f.Type == schemaMetaType {
				continue
			}
			if f.Anonymous && f.Type.Kind() == reflect.Struct {
				names[f.Name] = true
				queue.push(f.Type)
				continue
			}
		}
	}

	result := make([]reflect.StructField, 0, len(names))
	for k := range names {
		f, ok := typ.FieldByName(k)
		if !ok || !f.Anonymous {
			continue
		}
		metaF, ok := f.Type.FieldByName("Meta")
		if ok && metaF.Type == schemaMetaType {
			result = append(result, f)
		}
	}

	return result
}
