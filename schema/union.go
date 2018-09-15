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

package schema

import (
	"context"
	"fmt"
	"sort"
)

// UnwrapUnion takes in a value, maps it to a member of the union,
// and returns the raw object
type UnwrapUnion func(context.Context, interface{}) (interface{}, string)

var _ Type = (*UnionType)(nil)

// An UnionType represents a GraphQL Union
type UnionType struct {
	named
	schemaElement
	unwrap  UnwrapUnion
	members []*ObjectType
}

func (t *UnionType) isType() {}

// Members returns the list of object types that compose this union
func (t *UnionType) Members() []*ObjectType {
	return t.members
}

// Unwrap converts a union value to an (object value, type) pair
func (t *UnionType) Unwrap(ctx context.Context, v interface{}) (interface{}, string) {
	return t.unwrap(ctx, v)
}

func (t *UnionType) signature() string {
	return t.name
}

func (t *UnionType) writeSchemaDefinition(w *schemaWriter) {
	w.writeDescription(t.description)
	w.writeIndent()
	fmt.Fprintf(w, "union %s", t.name)

	for _, e := range t.directives {
		w.write(" ")
		e.writeSchemaDefinition(w)
	}

	if len(t.members) > 0 {
		w.write(" =")
		typs := make([]Type, len(t.members))
		for i, e := range t.members {
			typs[i] = e
		}
		sort.Stable(sortTypesBySignature(typs))
		for _, e := range typs {
			w.write(" | ")
			w.write(e.(*ObjectType).name)
		}
	}
}
