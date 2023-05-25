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
	"fmt"

	"github.com/housecanary/gq/ast"
)

// DirectiveLocation specifies where a directive may be used
type DirectiveLocation string

// Directive location values
const (
	DirectiveLocationQuery                DirectiveLocation = "QUERY"
	DirectiveLocationMutation                               = "MUTATION"
	DirectiveLocationSubscription                           = "SUBSCRIPTION"
	DirectiveLocationField                                  = "FIELD"
	DirectiveLocationFragmentDefinition                     = "FRAGMENT_DEFINITION"
	DirectiveLocationFragmentSpread                         = "FRAGMENT_SPREAD"
	DirectiveLocationInlineFragment                         = "INLINE_FRAGMENT"
	DirectiveLocationSchema                                 = "SCHEMA"
	DirectiveLocationScalar                                 = "SCALAR"
	DirectiveLocationObject                                 = "OBJECT"
	DirectiveLocationFieldDefinition                        = "FIELD_DEFINITION"
	DirectiveLocationArgumentDefinition                     = "ARGUMENT_DEFINITION"
	DirectiveLocationInterface                              = "INTERFACE"
	DirectiveLocationUnion                                  = "UNION"
	DirectiveLocationEnum                                   = "ENUM"
	DirectiveLocationEnumValue                              = "ENUM_VALUE"
	DirectiveLocationInputObject                            = "INPUT_OBJECT"
	DirectiveLocationInputFieldDefinition                   = "INPUT_FIELD_DEFINITION"
)

// DirectiveDefinition is a GraphQL directive definition
type DirectiveDefinition struct {
	named
	description string
	arguments   []*ArgumentDescriptor
	locations   []DirectiveLocation
	// FUTURE:  Should add hooks for directives to supply a callback
	// that can be used when a query element references the directive
}

func (d *DirectiveDefinition) writeSchemaDefinition(w *schemaWriter) {
	w.writeDescription(d.description)
	w.writeIndent()
	fmt.Fprintf(w, "directive @%s", d.name)
	if len(d.arguments) > 0 {
		w.write("(")
		argWriter := w.indented()
		for _, e := range d.arguments {
			argWriter.writeNL()
			e.writeSchemaDefinition(argWriter)
			argWriter.writeNL()
		}
		w.writeIndent()
		w.write(")")
	}
	w.write(" on")
	for _, e := range d.locations {
		w.write(" ")
		w.write(string(e))
	}
}

// DirectiveArgument represents an argument to a directive applied to a schema element
type DirectiveArgument struct {
	named
	value ast.Value
}

func (d *DirectiveArgument) writeSchemaDefinition(w *schemaWriter) {
	fmt.Fprintf(w, "%s: ", d.name)
	w.write(d.value.Representation())
}

// Value returns the value of the directive argument
func (d *DirectiveArgument) Value() LiteralValue {
	return LiteralValueFromAstValue(d.value)
}

// Directive represents a directive applied to a schema element
type Directive struct {
	named
	arguments []*DirectiveArgument
}

// Argument returns the argument with the specified name or nil
func (d *Directive) Argument(name string) *DirectiveArgument {
	for _, arg := range d.arguments {
		if arg.name == name {
			return arg
		}
	}

	return nil
}

func (d *Directive) writeSchemaDefinition(w *schemaWriter) {
	fmt.Fprintf(w, "@%s", d.name)
	if len(d.arguments) > 0 {
		w.write("(")
		for i, e := range d.arguments {
			if i != 0 {
				w.write(", ")
			}
			e.writeSchemaDefinition(w)
		}
		w.write(")")
	}
}
