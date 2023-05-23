// Copyright 2023 HouseCanary, Inc.
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

// Package ts is used to create a GraphQL schema from Go structs and types.
//
// "ts" is short for "typed schema". This package uses generics to ensure as much
// as possible compile time safety for the graphql types generated from the types.
// If it compiles, and the schema builds it should work without errors at runtime.
//
// Each GraphQL type has an example showing how to create that type from structs.
// Refer to these examples for details on how the mapping process works for that type.
package ts
