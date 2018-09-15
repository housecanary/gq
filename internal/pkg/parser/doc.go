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

// Package parser implements parsing/ast building of GraphQL documents
//
// Implementation notes:  everything in the "gen" folder is generated from the
// ANTLR grammar in the root level "grammar folder".  Query parsing is pretty slow
// could likely improve by using ANTLR semantic actions or ParseListener instead
// of the visitor APIs.
package parser
