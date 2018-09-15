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
	"encoding/json"
	"io"
)

type named struct {
	name string
}

// Name returns the name of this object
func (n *named) Name() string {
	return n.name
}

type schemaElement struct {
	description string
	directives  []*Directive
}

// Description returns the description of this object
func (e *schemaElement) Description() string {
	return e.description
}

// Directives returns the directives applied to this object
func (e *schemaElement) Directives() []*Directive {
	return e.directives
}

// Gets a directive by name.  Returns nil if the directive is not found.
func (e *schemaElement) GetDirective(name string) *Directive {
	for _, d := range e.directives {
		if d.name == name {
			return d
		}
	}

	return nil
}

type schemaSerializable interface {
	// convention: implementers write themselves with no trailing newline.
	writeSchemaDefinition(w *schemaWriter)
}

type errorCollector struct {
	err error
}

func (c *errorCollector) reportError(err error) {
	c.err = err
}

type schemaWriter struct {
	io.Writer
	*errorCollector
	prefix []byte
}

func (w *schemaWriter) write(s string) {
	io.WriteString(w, s)
}

func (w *schemaWriter) writeNL() {
	io.WriteString(w, "\n")
}

func (w *schemaWriter) writeIndent() {
	w.Write(w.prefix)
}

func (w *schemaWriter) writeEscapedString(s string) {
	b, err := json.Marshal(s)
	if err != nil {
		w.reportError(err)
		return
	}
	w.Write(b)
}

func (w *schemaWriter) writeDescription(desc string) {
	if desc != "" {
		w.writeIndent()
		w.writeEscapedString(desc)
		io.WriteString(w, "\n")
	}
}

func (w *schemaWriter) indented() *schemaWriter {
	amount := 2
	start := len(w.prefix)
	prefix := make([]byte, start+amount)
	copy(prefix, w.prefix)
	for i := 0; i < amount; i++ {
		prefix[i+start] = ' '
	}
	return &schemaWriter{
		w.Writer,
		w.errorCollector,
		prefix,
	}
}

func (w *schemaWriter) Write(b []byte) (int, error) {
	n, err := w.Writer.Write(b)
	if err != nil {
		w.reportError(err)
	}
	return n, err
}
