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

package query

import (
	"context"
	"fmt"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

// A compileContext holds the schema and ast of a query being compiled, and provides
// helpers to validate and build the query.  It is the global state of the query compilation.
type compileContext struct {
	*schema.Schema
	*ast.Document
	row int
	col int
}

func (c *compileContext) withLocation(row, col int) *compileContext {
	return &compileContext{
		c.Schema,
		c.Document,
		row,
		col,
	}
}

func (c *compileContext) newDefaultSelector() defaultSelector {
	return defaultSelector{c.row, c.col}
}

// expandFragment recursively expands a fragment into a flattened set of fields for the given object type.
func (c *compileContext) expandFragment(typeCondition string, selections ast.SelectionSet, typ *schema.ObjectType) ([]ast.Field, error) {
	if typ.Name() != typeCondition && !typ.HasInterface(typeCondition) {
		return nil, nil
	}

	fields := make([]ast.Field, 0)
	for _, sel := range selections {
		switch v := sel.(type) {
		case *ast.FieldSelection:
			fields = append(fields, v.Field)
		case *ast.FragmentSpreadSelection:
			fragDef := c.LookupFragmentDefinition(v.FragmentName)
			if fragDef == nil {
				// FUTURE: location info
				return nil, fmt.Errorf("Invalid query - unknown fragment %s", v.FragmentName)
			}
			fragFields, err := c.expandFragment(fragDef.OnType, fragDef.SelectionSet, typ)
			if err != nil {
				return nil, err
			}
			fields = append(fields, fragFields...)

		case *ast.InlineFragmentSelection:
			fragFields, err := c.expandFragment(v.OnType, v.SelectionSet, typ)
			if err != nil {
				return nil, err
			}
			fields = append(fields, fragFields...)
		}
	}
	return fields, nil
}

// makeArgumentResolver creates a function that can translate a literal value into
// the corresponding runtime object using the schema
func (c *compileContext) makeArgumentResolver(typ schema.InputableType) (argumentResolver, error) {
	switch t := typ.(type) {
	case *schema.InputObjectType:
		return func(ctx context.Context, v schema.LiteralValue) (interface{}, error) {
			return t.Decode(ctx, v)
		}, nil
	case *schema.ListType:
		elementResolver, err := c.makeArgumentResolver(t.Unwrap().(schema.InputableType))
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context, v schema.LiteralValue) (interface{}, error) {
			if v == nil {
				return nil, nil
			}

			listCreator := t.Unwrap().(schema.InputableType).InputListCreator()

			if av, ok := v.(schema.LiteralArray); ok {
				return listCreator.NewList(len(av), func(i int) (interface{}, error) {
					return elementResolver(ctx, av[i])
				})
			}

			// if we get a non-list value we have to wrap into a single element
			// list.
			// See https://facebook.github.io/graphql/June2018/#sec-Type-System.List
			resultElement, err := elementResolver(ctx, v)
			if err != nil {
				return nil, err
			}
			return listCreator.NewList(1, func(i int) (interface{}, error) {
				return resultElement, nil
			})
		}, nil

	case *schema.NotNilType:
		elementResolver, err := c.makeArgumentResolver(t.Unwrap().(schema.InputableType))
		if err != nil {
			return nil, err
		}
		return func(ctx context.Context, v schema.LiteralValue) (interface{}, error) {
			if v == nil {
				return nil, fmt.Errorf("Required value was not supplied")
			}
			return elementResolver(ctx, v)
		}, nil
	case *schema.ScalarType:
		return func(ctx context.Context, v schema.LiteralValue) (interface{}, error) {
			return t.Decode(ctx, v)
		}, nil
	case *schema.EnumType:
		return func(ctx context.Context, v schema.LiteralValue) (interface{}, error) {
			if v == nil {
				return t.Decode(ctx, v)
			}
			val, ok := v.(schema.LiteralString)
			if !ok {
				return nil, fmt.Errorf("Expected string, got %v", v)
			}
			return t.Decode(ctx, val)
		}, nil
	default:
		return nil, fmt.Errorf("Invalid type for input argument: %v", typ)
	}
}
