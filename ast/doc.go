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

// Package ast provides an AST of a GraphQL query or schema.
//
// This package's API is UNSTABLE.  New members may be added at
// any time, or renamed/moved.
//
// Design notes: this package should consist of simple structs, and perhaps
// some marshalling code.  No business logic, should just be storage of a parsed
// document
package ast
