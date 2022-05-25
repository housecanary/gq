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

	"github.com/housecanary/gq/ast"
)

// DecodeInputObject is a function that is capable of turning a raw value
// into an input object
type DecodeInputObject func(ctx InputObjectDecodeContext) (interface{}, error)

// InputObjectDecodeContext is passed to DecodeInputObject to allow that
// function to read field values as go interface{}s
type InputObjectDecodeContext interface {
	IsNil() bool
	GetFieldValue(name string) (interface{}, error)
}

var _ Type = (*InputObjectType)(nil)

// An InputObjectType represents a GraphQL input object
type InputObjectType struct {
	named
	schemaElement
	fields      map[string]*InputObjectFieldDescriptor
	decode      DecodeInputObject
	listCreator InputListCreator
}

func (t *InputObjectType) isType() {}

// Decode translates a literal value into the corresponding InputObject
// representation
func (t *InputObjectType) Decode(ctx context.Context, v LiteralValue) (interface{}, error) {
	return t.decode(&inputObjectDecodeContext{t, v, ctx})
}

// InputListCreator returns a creator for lists of this input object type
func (t *InputObjectType) InputListCreator() InputListCreator {
	return t.listCreator
}

// A InputObjectFieldDescriptor represents a field in a GraphQL input object.
// It has a name and a type.  It must be a scalar, a reference to an input object
// or a not-nil or list of either of these types
type InputObjectFieldDescriptor struct {
	named
	schemaElement
	typ             Type
	defaultValue    LiteralValue
	defaultValueAst ast.Value
	decoder         DecodeScalar
}

// Type returns the type of this field
func (d *InputObjectFieldDescriptor) Type() Type {
	return d.typ
}

type inputObjectDecodeContext struct {
	t    *InputObjectType
	root LiteralValue
	ctx  context.Context
}

func (i *inputObjectDecodeContext) IsNil() bool {
	return i.root == nil
}

func (i *inputObjectDecodeContext) GetFieldValue(name string) (interface{}, error) {
	fd, ok := i.t.fields[name]
	if !ok {
		return nil, fmt.Errorf("Input object requested value of invalid field %s", name)
	}

	if lo, ok := i.root.(LiteralObject); ok {
		val, ok := lo[name]
		if ok {
			return fd.decoder(i.ctx, val)
		}
		return fd.decoder(i.ctx, fd.defaultValue)
	}

	return nil, fmt.Errorf("Input value is not an object (%v)", i.root)

}

func inputObjectElementDecoder(typ InputableType) (DecodeScalar, error) {
	switch t := typ.(type) {
	case *ListType:
		elementDecoder, err := inputObjectElementDecoder(t.Unwrap().(InputableType))
		if err != nil {
			return nil, err
		}

		return func(ctx context.Context, in LiteralValue) (interface{}, error) {
			if in == nil {
				return nil, nil
			}

			if la, ok := in.(LiteralArray); ok {
				listCreator := t.Unwrap().(InputableType).InputListCreator()
				return listCreator.NewList(len(la), func(i int) (interface{}, error) {
					v, err := elementDecoder(ctx, la[i])
					if err != nil {
						return nil, err
					}
					return v, nil
				})
			}

			return nil, fmt.Errorf("Input value is not an array")
		}, nil
	case *NotNilType:
		elementDecoder, err := inputObjectElementDecoder(t.Unwrap().(InputableType))
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context, in LiteralValue) (interface{}, error) {
			if in == nil {
				return nil, fmt.Errorf("Value is unexpectedly nil")
			}
			return elementDecoder(ctx, in)
		}, nil
	case *EnumType:
		return func(ctx context.Context, in LiteralValue) (interface{}, error) {
			if in == nil {
				return t.Decode(ctx, nil)
			}
			strVal, ok := in.(LiteralString)
			if !ok {
				return nil, fmt.Errorf("Enum value was not a string")
			}
			return t.Decode(ctx, strVal)
		}, nil
	case *ScalarType:
		return t.Decode, nil
	case *InputObjectType:
		return t.Decode, nil
	}

	return nil, fmt.Errorf("Invalid input object type %v", typ)
}

func (t *InputObjectType) signature() string {
	return t.name
}

func (t *InputObjectType) writeSchemaDefinition(w *schemaWriter) {
	w.writeDescription(t.description)
	w.writeIndent()
	fmt.Fprintf(w, "input %s", t.name)
	for _, e := range t.directives {
		w.write(" ")
		e.writeSchemaDefinition(w)
	}

	vals := make([]*InputObjectFieldDescriptor, 0, len(t.fields))
	for _, v := range t.fields {
		vals = append(vals, v)
	}

	if len(vals) > 0 {
		sort.Stable(sortInputObjectFieldDescriptorsByName(vals))
		iw := w.indented()
		w.write(" {")
		for _, e := range vals {
			w.writeNL()
			e.writeSchemaDefinition(iw)
			w.writeNL()
		}
		w.writeIndent()
		w.write("}")
	}
}

func (d *InputObjectFieldDescriptor) writeSchemaDefinition(w *schemaWriter) {
	w.writeDescription(d.description)
	w.writeIndent()
	fmt.Fprintf(w, "%s: %s", d.name, d.typ.signature())
	if d.defaultValue != nil {
		w.write(" = ")
		w.write(d.defaultValueAst.Representation())
	}
	for _, e := range d.directives {
		w.write(" ")
		e.writeSchemaDefinition(w)
	}
}
