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

	"github.com/housecanary/gq/ast"
)

// AsyncValue is returned by Resolvers.  It indicates that the value of the field
// is being computed asynchronously, and the caller should invoke Await() to obtain
// the final value.
type AsyncValue interface {
	Await(context.Context) (interface{}, error)
}

// AsyncValueFunc adapts a function to the AsyncValue interface
type AsyncValueFunc func(context.Context) (interface{}, error)

// Await implements AsyncValue.Await
func (f AsyncValueFunc) Await(ctx context.Context) (interface{}, error) {
	return f(ctx)
}

// ResolverContext represents the contract between a Resolver and the
// runtime environment
type ResolverContext interface {
	context.Context
	ChildWalker
	GetArgumentValue(name string) (interface{}, error)
	GetRawArgumentValue(name string) (LiteralValue, error)
	ChildFieldsIterator() FieldSelectionIterator
}

// ChildWalker is an object that can walk the child selections of a selection
type ChildWalker interface {
	// FieldWalkCB is called for each (ast selection, field) pair of the flattened
	// children of the value being resolved.  If the callback returns true, the walk
	// is aborted.
	//
	// NOTE: the in the case of interfaces or unions, the walker will be called for
	// each possible selection.  For example, given the query
	//
	// foo {
	//   a
	//   b
	//   ... on Typ1 {
	//     c
	//   }
	//   ... on Typ2 {
	//     d
	//   }
	// }
	//
	// where foo is an interface type implemented by Typ1, Typ2, Typ3
	// the sequence of calls would be
	// (a, Typ1.a)
	// (b, Typ1.b)
	// (c, Typ1.c)
	// (a, Typ2.a)
	// (b, Typ2.b)
	// (d, Typ2.d)
	// (a, Typ3.a)
	// (b, Typ3.b)
	WalkChildSelections(FieldWalkCB) bool
}

// FieldWalkCB is a callback for ChildWalker.WalkChildSelections
type FieldWalkCB func(selection *ast.Field, field *FieldDescriptor, walker ChildWalker) bool

type FieldSelectionIterator interface {
	Next() bool
	Selection() *ast.Field
	SchemaField() *FieldDescriptor
	ChildFieldsIterator() FieldSelectionIterator
}

// A Resolver is responsible for extracting the value of a field from
// a containing object.
type Resolver interface {
	// Resolve extracts a field value from the containing object
	Resolve(ctx context.Context, v interface{}) (interface{}, error)
}

// A SafeResolver is a resolver that can resolve values without panicing.
type SafeResolver interface {
	Resolver
	ResolveSafe(ctx context.Context, v interface{}) (interface{}, error)
}

// MarkSafe marks a resolver as Safe
func MarkSafe(resolver Resolver) SafeResolver {
	return safeResolver{resolver}
}

type safeResolver struct {
	Resolver
}

func (s safeResolver) ResolveSafe(ctx context.Context, v interface{}) (interface{}, error) {
	return s.Resolve(ctx, v)
}

// A SimpleResolver is a function that can be used
// to implement the Resolver contract
type SimpleResolver func(v interface{}) (interface{}, error)

// Resolve implements the Resolver interface
func (r SimpleResolver) Resolve(ctx context.Context, v interface{}) (interface{}, error) {
	return r(v)
}

// ContextResolver is a function that can be used to resolve values given
// a context.Context
type ContextResolver func(ctx context.Context, v interface{}) (interface{}, error)

// Resolve implements the Resolver interface
func (r ContextResolver) Resolve(ctx context.Context, v interface{}) (interface{}, error) {
	return r(ctx, v)
}

// FullResolver is a function that can be used to resolve values given
// a ResolverContext
type FullResolver func(ctx ResolverContext, v interface{}) (interface{}, error)

// Resolve implements the Resolver interface
func (r FullResolver) Resolve(ctx context.Context, v interface{}) (interface{}, error) {
	return r(ctx.(ResolverContext), v)
}

var _ Type = (*ObjectType)(nil)

// ObjectType defines the type of a GraphQL Object
type ObjectType struct {
	named
	schemaElement
	fieldsByName map[string]*FieldDescriptor
	interfaces   []*InterfaceType
}

func (t *ObjectType) isType() {}

// Field looks up a field by name.  If not found, nil is returned.
func (t *ObjectType) Field(name string) *FieldDescriptor {
	return t.fieldsByName[name]
}

// HasInterface checks if the object implements a particular interface
func (t *ObjectType) HasInterface(name string) bool {
	for _, it := range t.interfaces {
		if it.name == name {
			return true
		}
	}
	return false
}

// A FieldDescriptor represents a field in a GraphQL object.  It has a name, a type,
// and a function used to resolve the value of the field from a containing object.
type FieldDescriptor struct {
	named
	schemaElement
	arguments []*ArgumentDescriptor
	typ       Type
	r         Resolver
}

// Arguments returns the defined arguments of this field
func (d *FieldDescriptor) Arguments() []*ArgumentDescriptor {
	return d.arguments
}

// Type returns the type of this field
func (d *FieldDescriptor) Type() Type {
	return d.typ
}

// Resolver returns the function used to extract the value of this field from a
// containing object
func (d *FieldDescriptor) Resolver() Resolver {
	return d.r
}

// An ArgumentDescriptor represents an argument to a field.
type ArgumentDescriptor struct {
	named
	schemaElement
	typ          Type
	defaultValue ast.Value
}

// Type returns the type of this argument
func (d *ArgumentDescriptor) Type() Type {
	return d.typ
}

// DefaultValue returns the default value of this field
func (d *ArgumentDescriptor) DefaultValue() LiteralValue {
	return literalValueFromAstValue(d.defaultValue)
}

func (t *ObjectType) signature() string {
	return t.name
}

func (t *ObjectType) writeSchemaDefinition(w *schemaWriter) {
	w.writeDescription(t.description)
	w.writeIndent()
	fmt.Fprintf(w, "object %s", t.name)

	if len(t.interfaces) > 0 {
		w.write(" implements")
		for _, e := range t.interfaces {
			w.write(" & ")
			w.write(e.name)
		}
	}

	for _, e := range t.directives {
		w.write(" ")
		e.writeSchemaDefinition(w)
	}

	vals := make([]*FieldDescriptor, 0, len(t.fieldsByName))
	for _, v := range t.fieldsByName {
		if strings.HasPrefix(v.name, "__") {
			continue
		}
		vals = append(vals, v)
	}

	if len(vals) > 0 {
		fw := w.indented()
		sort.Stable(sortFieldDescriptorsByName(vals))
		w.write(" {")
		for _, e := range vals {
			w.writeNL()
			e.writeSchemaDefinition(fw)
			w.writeNL()
		}
		w.writeIndent()
		w.write("}")
	}
}

func (d *FieldDescriptor) writeSchemaDefinition(w *schemaWriter) {
	w.writeDescription(d.description)
	w.writeIndent()
	fmt.Fprintf(w, "%s", d.name)

	if len(d.arguments) > 0 {
		w.write(" (")
		argWriter := w.indented()
		for _, e := range d.arguments {
			argWriter.writeNL()
			e.writeSchemaDefinition(argWriter)
			argWriter.writeNL()
		}
		w.writeIndent()
		w.write(")")
	}

	w.write(": ")
	w.write(d.typ.signature())

	for _, e := range d.directives {
		w.write(" ")
		e.writeSchemaDefinition(w)
	}
}

func (d *ArgumentDescriptor) writeSchemaDefinition(w *schemaWriter) {
	w.writeDescription(d.description)
	w.writeIndent()
	fmt.Fprintf(w, "%s: %s", d.name, d.typ.signature())
	if d.defaultValue != nil {
		w.write(" = ")
		w.write(d.defaultValue.Representation())
	}
	for _, e := range d.directives {
		w.write(" ")
		e.writeSchemaDefinition(w)
	}
}
