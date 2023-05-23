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

package ts

import (
	"context"
	"fmt"
	"reflect"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser"
	"github.com/housecanary/gq/schema"
)

// A Result is a value returned from a resolver function that encapsulates the
// value the function produces. The Result interface allows for more complicated
// return values that might require asynchronous resolution.
//
// See the result subpackage for implementation of many helper result types
type Result[T any] interface {
	UnpackResult() (T, func(context.Context) (T, error), error)
}

// A FieldType represents the GQL type of a virtual field on an object fulfilled
// by invoking a method
type FieldType[R any] struct {
	def              string
	rType            reflect.Type
	aType            reflect.Type
	ResolverFunction R
	makeResolverFn   func(c *buildContext) (schema.Resolver, fieldInvoker, error)
}

func (ft *FieldType[R]) originalDefinition() string {
	return ft.def
}

func (ft *FieldType[R]) makeResolver(c *buildContext) (schema.Resolver, fieldInvoker, error) {
	return ft.makeResolverFn(c)
}

func (ft *FieldType[R]) buildFieldDef(c *buildContext) (*ast.FieldDefinition, bool, error) {
	fd, err := parser.ParseTSResolverFieldDefinition(ft.def)
	if err != nil {
		return nil, false, err
	}

	if fd.Name == "" {
		return nil, false, fmt.Errorf("name is required in field definition %s", ft.def)
	}

	wasTypeInferred := true
	if fd.Type == nil {
		typ, err := c.astTypeForGoType(ft.rType)
		if err != nil {
			return nil, false, err
		}
		fd.Type = typ
		wasTypeInferred = true
	} else {
		err := c.checkTypeCompatible(ft.rType, fd.Type)
		if err != nil {
			return nil, false, err
		}
	}

	if ft.aType != nil {
		fields := reflect.VisibleFields(ft.aType)
		for _, field := range fields {
			if field.Tag.Get("ts") == "inject" {
				continue
			}

			ad, _, err := parseStructField(c, field, parser.ParsePartialInputValueDefinition)
			if err != nil {
				return nil, false, fmt.Errorf("cannot parse input argument %s: %w", field.Name, err)
			}

			if ad == nil {
				continue
			}

			fd.ArgumentsDefinition = append(fd.ArgumentsDefinition, ad)
		}
	}

	return fd, wasTypeInferred, nil
}

type argBinder[A any] func(QueryInfo, *A) error

func makeArgBinder[A any](c *buildContext) (argBinder[A], error) {
	typ := typeOf[A]()
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("invalid arguments type, expected a struct got %v", typ.Kind())
	}

	fields := reflect.VisibleFields(typ)
	binds := make([]func(QueryInfo, reflect.Value) error, 0, len(fields))
	for _, field := range fields {
		field := field

		if field.Tag.Get("ts") == "inject" {
			provider, ok := c.providers[field.Type]
			if !ok {
				return nil, fmt.Errorf("No provider registered for type %v", field.Type)
			}

			binds = append(binds, func(qi QueryInfo, v reflect.Value) error {
				av := provider(qi)
				v.FieldByIndex(field.Index).Set(reflect.ValueOf(av))
				return nil
			})
			continue
		}

		ad, _, err := parseStructField(c, field, parser.ParsePartialInputValueDefinition)
		if ad == nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("cannot parse input argument %s: %w", field.Name, err)
		}

		binds = append(binds, func(qi QueryInfo, v reflect.Value) error {
			av, err := qi.ArgumentValue(ad.Name)
			if err != nil {
				return err
			}
			v.FieldByIndex(field.Index).Set(reflect.ValueOf(av))
			return nil
		})
	}

	return func(qi QueryInfo, a *A) error {
		rv := reflect.ValueOf(a).Elem()
		for _, bind := range binds {
			if err := bind(qi, rv); err != nil {
				return err
			}
		}
		return nil
	}, nil
}

type queryInfo struct {
	schema.ResolverContext
}

func (qi queryInfo) QueryContext() context.Context {
	return qi.ResolverContext
}

func (qi queryInfo) ArgumentValue(name string) (interface{}, error) {
	return qi.ResolverContext.GetArgumentValue(name)
}

func returnResult[T any](r Result[T]) (interface{}, error) {
	t, f, e := r.UnpackResult()
	if e != nil {
		return nil, e
	}
	if f != nil {
		return schema.AsyncValueFunc(func(ctx context.Context) (interface{}, error) {
			return f(ctx)
		}), nil
	}

	return t, nil
}
