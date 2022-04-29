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

// Package types provides implementations of all GraphQL built in types.
//
// As a convenience, the types in this package are Scannable from SQL and Unmarshalable
// from JSON (as well as providing fast unmarhshalling via jsoniter) so that they can be
// used to directly map data from a database or JSON speaking service.
package types
