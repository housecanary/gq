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

// EncodeEnum is a function that is capable of turning a Go value into
// a literal value
type EncodeEnum func(ctx context.Context, v interface{}) (LiteralValue, error)

// DecodeEnum is a function that is capable of turning a literal value
// into a Go value
type DecodeEnum func(ctx context.Context, v LiteralValue) (interface{}, error)

var _ Type = (*EnumType)(nil)

// A EnumType represents a GraphQL Enum
type EnumType struct {
	named
	schemaElement
	values      map[LiteralString]*enumValueDescriptor
	encode      EncodeEnum
	decode      DecodeEnum
	listCreator InputListCreator
}

func (t *EnumType) isType() {}

type enumValueDescriptor struct {
	named
	schemaElement
}

// Encode translates a Go value into a literal value
func (t *EnumType) Encode(ctx context.Context, v interface{}) (LiteralValue, error) {
	lv, err := t.encode(ctx, v)
	if err != nil {
		return nil, err
	}

	if lv == nil {
		return nil, nil
	}

	if sv, ok := lv.(LiteralString); ok {
		if _, ok := t.values[sv]; !ok {
			return nil, fmt.Errorf("Enum encoded to invalid value %s", sv)
		}
		return lv, nil
	}
	panic("unreachable")
}

// Decode translates a literal value into a Go value
func (t *EnumType) Decode(ctx context.Context, v LiteralValue) (interface{}, error) {
	if v != nil {
		if sv, ok := v.(LiteralString); ok {
			if _, ok := t.values[sv]; !ok {
				return nil, fmt.Errorf("Received invalid enum value %s", sv)
			}
		} else {
			return nil, fmt.Errorf("Enum input value was not a string: %v", v)
		}
	}
	return t.decode(ctx, v)
}

// InputListCreator returns a creator for lists of this enum type
func (t *EnumType) InputListCreator() InputListCreator {
	return t.listCreator
}

func (t *EnumType) signature() string {
	return t.name
}

func (t *EnumType) writeSchemaDefinition(w *schemaWriter) {
	w.writeDescription(t.description)
	w.writeIndent()
	fmt.Fprintf(w, "enum %s", t.name)
	for _, e := range t.directives {
		w.write(" ")
		e.writeSchemaDefinition(w)
	}
	w.write(" {")
	vals := make([]*enumValueDescriptor, 0, len(t.values))
	for _, v := range t.values {
		vals = append(vals, v)
	}
	sort.Stable(sortEnumValuesByName(vals))
	iw := w.indented()
	for _, e := range vals {
		w.writeNL()
		e.writeSchemaDefinition(iw)
		w.writeNL()
	}
	w.writeIndent()
	w.write("}")
}

func (t *enumValueDescriptor) writeSchemaDefinition(w *schemaWriter) {
	w.writeDescription(t.description)
	w.writeIndent()
	w.write(t.name)
	for _, e := range t.directives {
		w.write(" ")
		e.writeSchemaDefinition(w)
	}
}
