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

package parser

import (
	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser/gen"
)

// ParseQuery parses a textual representation of a GraphQL query into an AST
func ParseQuery(input string) (doc *ast.Document, err ParseError) {
	doc = nil
	err = safeParse(input, func(p *gen.GraphqlParser) {
		doc = p.Document().Accept(&queryDocumentVisitor{}).(*ast.Document)
	})
	return
}
