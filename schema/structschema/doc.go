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

// Package structschema is used to create a GraphQL schema via reflection on go types.
//
// A schema is built in structschema by types that follow a particular set of naming conventions:
//
// Object types
//
// A Go struct optionally includes a field of type Meta.  The Meta type is a zero length
// struct that is intended to serve as a place to attach a struct tag containing additional
// GraphQL schema definition.  Meta fields will be ignored for schema processing.  If the Meta
// field exists, its struct tag is expected to contain a partial GraphQL object definition.  This can
// be used to declare the signatures for fields accessed via resolver methods, change the name of the
// type, or add directives to the type.  Note that struct tags on Meta fields do NOT conform to the
// standard Go convention of `namespace:"..."`.  Instead, the entire struct tag is interpreted as
// GraphQL schema definition.
//
// GraphQL fields for the object type are constructed by merging any fields declared on a Meta with
// the public fields of the struct.  Each struct field may optionally have a "gq" field in its struct
// tag.  The gq part of the struct tag consists of two parts separated by a ;.  If the second part is
// unused, the ; may be omitted.  The first part of the field consists of a GraphQL field definition, or
// a subset thereof.  Any missing portions of the GraphQL field signature will be filled in with what
// information can be obtained via reflection.  The second part of the field consists of a doc string
// that is set as the field's description.  As a special case, to omit a field from GraphQL use the string
// "-" (example: `gq:"-"`)
//
// GraphQL fields declared solely via Meta declaration are expected to have a resolver method.  Resolver
// methods are named Resolve<FieldName> (case insensitively), an should conform to the following conventions:
//
// A resolver method optionally takes a ResolverContext as its first argument.  If a ResolverContext is the
// first argument, parameters to accept GraphQL declared arguments become optional.
//
// After the context argument, a resolver method may accept any number of arguments that can be provided by
// ArgProviders registered with the builder.
//
// After the context and injected arguments, the method should declare, in order, parameters matching the GraphQL
// arguments for the field.
//
// The method's return value should conform to one of the following signatures:
// (TypeFromGraphQL) - Synchronously returns a value
// (TypeFromGraphQL, error) - Synchronously returns either a value or an error
// (chan TypeFromGraphQL) - Asynchronously returns a value.  One value will be read from the chan, further values will be ignored
// (chan TypeFromGraphQL, chan error) - Asynchronously returns either a value or an error.  The first value to appear on either the value chan or error chan will be the value used.
// (func () <supported return signature>) - Asynchronously returns something specified by the return value of the returned function
//
// Interface types
//
// A Go struct containing a single field named "Interface" of type interface{...}.  This field
// may optionally have a struct tag which is processed according to the same rule as Meta on an
// object type.
//
// Union types
//
// A Go struct containing a single field named "Union" of type interface{...}.  This field
// may optionally have a struct tag which is processed according to the same rule as Meta on an
// object type.
//
// Enum types
//
// A Go struct with a single anonymous embedding of type ss.Enum.  This field
// may optionally have a struct tag which is processed according to the same rule as Meta on an
// object type.  ss.Enum is an alias for string, so at runtime the value of the enum will be stored
// in this field.
//
// Scalar types
//
// Any Go type which implements ScalarMarshaler and ScalarUnmarshaler.
package structschema
