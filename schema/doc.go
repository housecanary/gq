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

// Package schema is used to describe a GraphQL schema
//
// A schema is essentially a registry of types.  A type provides
// metadata for GraphQL clients via introspections, as well as providing
// type specific methods to allow the query package to operate over instances
// of the type.
//
// Design notes:
//   - No reflection allowed in this package.
//   - Schemas are immutable.  One creates a schema by using a Builder.
//   - Schemas are safe for use across threads
package schema
