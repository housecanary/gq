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

// ParsePartialFieldDefinition parses a partial field definition
func ParsePartialFieldDefinition(input string) (def *ast.FieldDefinition, err ParseError) {
	def = nil
	err = safeParse(input, func(p *gen.GraphqlParser) {
		def = p.PartialFieldDefinition().Accept(&schemaVisitor{}).(*ast.FieldDefinition)
	})
	return
}

// ParseTSResolverFieldDefinition parses a partial field definition for a TS resolver
func ParseTSResolverFieldDefinition(input string) (def *ast.FieldDefinition, err ParseError) {
	def = nil
	err = safeParse(input, func(p *gen.GraphqlParser) {
		def = p.TsResolverFieldDefinition().Accept(&schemaVisitor{}).(*ast.FieldDefinition)
	})
	return
}

// ParsePartialObjectTypeDefinition parses a partial object type definition
func ParsePartialObjectTypeDefinition(input string) (def *ast.ObjectTypeDefinition, err ParseError) {
	def = nil
	err = safeParse(input, func(p *gen.GraphqlParser) {
		def = p.PartialObjectTypeDefinition().Accept(&schemaVisitor{}).(*ast.ObjectTypeDefinition)
	})
	return
}

// ParsePartialInputObjectTypeDefinition parses a partial input object type definition
func ParsePartialInputObjectTypeDefinition(input string) (def *ast.InputObjectTypeDefinition, err ParseError) {
	def = nil
	err = safeParse(input, func(p *gen.GraphqlParser) {
		def = p.PartialInputObjectTypeDefinition().Accept(&schemaVisitor{}).(*ast.InputObjectTypeDefinition)
	})
	return
}

// ParsePartialInputValueDefinition parses a partial input object type definition
func ParsePartialInputValueDefinition(input string) (def *ast.InputValueDefinition, err ParseError) {
	def = nil
	err = safeParse(input, func(p *gen.GraphqlParser) {
		def = p.PartialInputValueDefinition().Accept(&schemaVisitor{}).(*ast.InputValueDefinition)
	})
	return
}

// ParsePartialEnumTypeDefinition parses a partial enum type definition
func ParsePartialEnumTypeDefinition(input string) (def *ast.EnumTypeDefinition, err ParseError) {
	def = nil
	err = safeParse(input, func(p *gen.GraphqlParser) {
		def = p.PartialEnumTypeDefinition().Accept(&schemaVisitor{}).(*ast.EnumTypeDefinition)
	})
	return
}

// ParseEnumValueDefinition parses an enum value definition
func ParseEnumValueDefinition(input string) (def *ast.EnumValueDefinition, err ParseError) {
	def = nil
	err = safeParse(input, func(p *gen.GraphqlParser) {
		def = p.EnumValueDefinition().Accept(&schemaVisitor{}).(*ast.EnumValueDefinition)
	})
	return
}

// ParsePartialInterfaceTypeDefinition parses a partial enum type definition
func ParsePartialInterfaceTypeDefinition(input string) (def *ast.InterfaceTypeDefinition, err ParseError) {
	def = nil
	err = safeParse(input, func(p *gen.GraphqlParser) {
		def = p.PartialInterfaceTypeDefinition().Accept(&schemaVisitor{}).(*ast.InterfaceTypeDefinition)
	})
	return
}

// ParsePartialUnionTypeDefinition parses a partial enum type definition
func ParsePartialUnionTypeDefinition(input string) (def *ast.UnionTypeDefinition, err ParseError) {
	def = nil
	err = safeParse(input, func(p *gen.GraphqlParser) {
		def = p.PartialUnionTypeDefinition().Accept(&schemaVisitor{}).(*ast.UnionTypeDefinition)
	})
	return
}

// ParsePartialScalarTypeDefinition parses a partial scalar type definition
func ParsePartialScalarTypeDefinition(input string) (def *ast.ScalarTypeDefinition, err ParseError) {
	def = nil
	err = safeParse(input, func(p *gen.GraphqlParser) {
		def = p.PartialScalarTypeDefinition().Accept(&schemaVisitor{}).(*ast.ScalarTypeDefinition)
	})
	return
}

// ParseTSTypeDefinition parses a partial scalar type definition
func ParseTSTypeDefinition(input string) (def *ast.BasicTypeDefinition, err ParseError) {
	def = nil
	err = safeParse(input, func(p *gen.GraphqlParser) {
		def = p.TsTypeDefinition().Accept(&schemaVisitor{}).(*ast.BasicTypeDefinition)
	})
	return
}
