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
	"io"
	"sort"
)

// A Schema represents a GraphQL schema against which queries
// may be executed
type Schema struct {
	QueryType  *ObjectType
	allTypes   map[string]Type
	directives []*DirectiveDefinition
}

// WriteDefinition writes this schema as a GraphQL schema definition
func (s *Schema) WriteDefinition(w io.Writer) error {
	ec := &errorCollector{}
	sw := &schemaWriter{w, ec, nil}

	sw.write("schema {")

	iw := sw.indented()
	iw.writeNL()
	iw.writeIndent()
	iw.write("query: ")
	iw.write(s.QueryType.signature())
	iw.writeNL()

	dds := make([]*DirectiveDefinition, len(s.directives))
	copy(dds, s.directives)

	sort.Stable(sortDirectiveDefsByName(dds))
	for _, e := range dds {
		iw.writeNL()
		e.writeSchemaDefinition(iw)
		iw.writeNL()
	}

	typs := make([]Type, 0, len(s.allTypes))
	for _, v := range s.allTypes {
		typs = append(typs, v)
	}

	sort.Stable(sortTypesBySignature(typs))

	for _, e := range typs {
		if isBuiltin(e) {
			continue
		}
		if ss, ok := e.(schemaSerializable); ok {
			iw.writeNL()
			ss.writeSchemaDefinition(iw)
			iw.writeNL()
		}
	}

	sw.write("}")
	return ec.err
}

// Type is a marker interface for all types.
type Type interface {
	isType()
	signature() string
}

// A WrappedType is a wrapper around another existing type.
// NotNullType and ListType are Wrapped Types
type WrappedType interface {
	Type
	Unwrap() Type
}

// An InputListCreator is used to create lists of input elements
type InputListCreator interface {
	NewList(size int, get func(i int) (interface{}, error)) (interface{}, error)
	Creator() InputListCreator
}

// An InputableType is an interface for types that can be used as input elements
// InputObjectType, ScalarType, ListType, NotNilType and EnumType implement this interface.
type InputableType interface {
	// InputListCreator returns an adapter that can be used to create lists
	// of this type
	InputListCreator() InputListCreator
}

func isBuiltin(v interface{}) bool {
	if st, ok := v.(*ScalarType); ok {
		switch st.name {
		case "ID":
			fallthrough
		case "String":
			fallthrough
		case "Int":
			fallthrough
		case "Float":
			fallthrough
		case "Boolean":
			return true
		}
	}

	return false
}
