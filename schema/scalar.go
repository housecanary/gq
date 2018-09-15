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
)

// EncodeScalar is a function that is capable of turning a Go value into
// a literal value
type EncodeScalar func(ctx context.Context, v interface{}) (LiteralValue, error)

// DecodeScalar is a function that is capable of turning a literal value
// into a Go value
type DecodeScalar func(ctx context.Context, v LiteralValue) (interface{}, error)

// ScalarMarshaler defines a value that can be converted to a GraphQL scalar type
type ScalarMarshaler interface {
	ToLiteralValue() (LiteralValue, error)
}

// EncodeScalarMarshaler is a EncodeScalar for values that implements ScalarConvertible
func EncodeScalarMarshaler(ctx context.Context, v interface{}) (LiteralValue, error) {
	cv, ok := v.(ScalarMarshaler)
	if !ok {
		return nil, fmt.Errorf("%v is not convertible to a scalar", v)
	}

	return cv.ToLiteralValue()
}

// ScalarUnmarshaler defines a value that can be converted from a GraphQL scalar type
type ScalarUnmarshaler interface {
	FromLiteralValue(LiteralValue) error
}

// A ScalarCollector is an object that can receive a scalar value
type ScalarCollector interface {
	Int(v int64)
	Float(v float64)
	Bool(v bool)
	String(v string)
}

// A CollectableScalar is a scalar that can be collected into a ScalarCollector
type CollectableScalar interface {
	CollectInto(col ScalarCollector)
}

var _ Type = (*ScalarType)(nil)

// A ScalarType represents a GraphQL Scalar
type ScalarType struct {
	named
	schemaElement
	encode EncodeScalar
	decode DecodeScalar
}

func (t *ScalarType) isType() {}

// Encode translates a Go value into a literal value
func (t *ScalarType) Encode(ctx context.Context, v interface{}) (LiteralValue, error) {
	return t.encode(ctx, v)
}

// Decode translates a literal value into a Go value
func (t *ScalarType) Decode(ctx context.Context, v LiteralValue) (interface{}, error) {
	return t.decode(ctx, v)
}

func (t *ScalarType) signature() string {
	return t.name
}

func (t *ScalarType) writeSchemaDefinition(w *schemaWriter) {
	w.writeDescription(t.description)
	w.writeIndent()
	fmt.Fprintf(w, "scalar %s", t.name)

	for _, e := range t.directives {
		w.write(" ")
		e.writeSchemaDefinition(w)
	}
}
