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

type argBinder[A any, Q internalQueryInfo] func(Q, *A) error

func makeArgBinder[A any, Q internalQueryInfo](c *buildContext) (argBinder[A, Q], error) {
	typ := typeOf[A]()
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("invalid arguments type, expected a struct got %v", typ.Kind())
	}

	fields := reflect.VisibleFields(typ)
	binds := make([]func(Q, reflect.Value) error, 0, len(fields))
	for _, field := range fields {
		field := field

		if field.Tag.Get("ts") == "inject" {
			provider, ok := c.providers[field.Type]
			if !ok {
				return nil, fmt.Errorf("No provider registered for type %v", field.Type)
			}

			binds = append(binds, func(qi Q, v reflect.Value) error {
				av := provider(qi)
				target, err := v.FieldByIndexErr(field.Index)
				if err != nil {
					return err
				}
				target.Set(reflect.ValueOf(av))
				return nil
			})
			continue
		}

		ad, _, err := parseStructField(c, field, parser.ParsePartialInputValueDefinition)
		if ad == nil && err == nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("cannot parse input argument %s: %w", field.Name, err)
		}

		converter := makeInputConverterForType(c, ad.Type, field.Type)
		if converter == nil {
			return nil, fmt.Errorf("input argument %s cannot be mapped to an input type", field.Name)
		}
		binds = append(binds, func(qi Q, v reflect.Value) error {
			target, err := v.FieldByIndexErr(field.Index)
			if err != nil {
				return err
			}
			err = qi.setArgumentValue(ad.Name, target, converter)
			if err != nil {
				return fmt.Errorf("invalid input for argument %s: %w", ad.Name, err)
			}
			return nil
		})
	}

	return func(qi Q, a *A) error {
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

func (qi queryInfo) ArgumentValue(name string) (any, error) {
	return qi.ResolverContext.GetArgumentValue(name)
}

func (qi queryInfo) setArgumentValue(name string, dest reflect.Value, converter inputConverter) error {
	raw, err := qi.ResolverContext.GetRawArgumentValue(name)

	if err != nil {
		return err
	}
	return converter(raw, dest)
}

func returnResult[T any](r Result[T], transform func(T) any) (any, error) {
	t, f, e := r.UnpackResult()
	if e != nil {
		return nil, e
	}
	if f != nil {
		return schema.AsyncValueFunc(func(ctx context.Context) (any, error) {
			t, e := f(ctx)
			if e != nil {
				return nil, e
			}
			return transform(t), e
		}), nil
	}

	return transform(t), nil
}
