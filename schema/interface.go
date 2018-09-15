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
	"strings"
)

// UnwrapInterface takes in a value, maps it to an object implementing the interface,
// and returns the raw object
type UnwrapInterface func(context.Context, interface{}) (interface{}, string)

var _ Type = (*InterfaceType)(nil)

// An InterfaceType represents a GraphQL Interface
type InterfaceType struct {
	named
	schemaElement
	fields          map[string]*FieldDescriptor
	unwrap          UnwrapInterface
	implementations []*ObjectType
}

func (t *InterfaceType) isType() {}

// HasField checks if the specified field is defined on this interface
func (t *InterfaceType) HasField(name string) bool {
	return t.fields[name] != nil
}

// Implementations returns the list of object types that implement this interface
func (t *InterfaceType) Implementations() []*ObjectType {
	return t.implementations
}

// Unwrap converts an interface value to an (object value, type) pair
func (t *InterfaceType) Unwrap(ctx context.Context, v interface{}) (interface{}, string) {
	return t.unwrap(ctx, v)
}

func (t *InterfaceType) signature() string {
	return t.name
}

func (t *InterfaceType) writeSchemaDefinition(w *schemaWriter) {
	w.writeDescription(t.description)
	w.writeIndent()
	fmt.Fprintf(w, "interface %s", t.name)
	for _, e := range t.directives {
		w.write(" ")
		e.writeSchemaDefinition(w)
	}

	vals := make([]*FieldDescriptor, 0, len(t.fields))
	for _, v := range t.fields {
		if strings.HasPrefix(v.name, "__") {
			continue
		}
		vals = append(vals, v)
	}

	if len(vals) > 0 {
		sort.Stable(sortFieldDescriptorsByName(vals))
		w.write(" {")
		iw := w.indented()
		for _, e := range vals {
			w.writeNL()
			e.writeSchemaDefinition(iw)
			w.writeNL()
		}
		w.writeIndent()
		w.write("}")
	}
}
