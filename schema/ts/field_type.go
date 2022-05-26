package ts

import (
	"context"
	"fmt"
	"reflect"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser"
	"github.com/housecanary/gq/schema"
	"github.com/housecanary/gq/schema/ts/result"
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
type FieldType[O any] struct {
	def          string
	rType        reflect.Type
	aType        reflect.Type
	makeResolver func(c *buildContext) (schema.Resolver, fieldInvoker, error)
}

func (ft *FieldType[O]) buildFieldDef(c *buildContext) (*ast.FieldDefinition, bool, error) {
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

// Field creates a new field resolved by a method taking a source object
func Field[R, O any](def string, rf func(*O) Result[R]) *FieldType[O] {
	return &FieldType[O]{
		def:   def,
		rType: typeOf[R](),
		makeResolver: func(c *buildContext) (schema.Resolver, fieldInvoker, error) {
			resolver := schema.SimpleResolver(func(v interface{}) (interface{}, error) {
				return returnResult(rf(v.(*O)))
			})

			invoker := func(q QueryInfo, o interface{}) interface{} {
				return rf(o.(*O))
			}

			return resolver, invoker, nil
		},
	}
}

// FieldA creates a new field resolved by a method taking a source object and arguments
func FieldA[R, A, O any](def string, rf func(*O, *A) Result[R]) *FieldType[O] {
	return &FieldType[O]{
		def:   def,
		rType: typeOf[R](),
		aType: typeOf[A](),
		makeResolver: func(c *buildContext) (schema.Resolver, fieldInvoker, error) {
			bindArgs, err := makeArgBinder[A](c)
			if err != nil {
				return nil, nil, err
			}
			resolver := schema.FullResolver(func(ctx schema.ResolverContext, v interface{}) (interface{}, error) {
				var args A
				if err := bindArgs(queryInfo{ctx}, &args); err != nil {
					var empty R
					return empty, err
				}
				return returnResult(rf(v.(*O), &args))
			})

			invoker := func(q QueryInfo, o interface{}) interface{} {
				var args A
				if err := bindArgs(q, &args); err != nil {
					return result.Error[R](err)
				}
				return rf(o.(*O), &args)
			}

			return resolver, invoker, nil
		},
	}
}

// FieldQ creates a new field resolved by a method taking a source object and query info
func FieldQ[R, O any](def string, rf func(QueryInfo, *O) Result[R]) *FieldType[O] {
	return &FieldType[O]{
		def:   def,
		rType: typeOf[R](),
		makeResolver: func(c *buildContext) (schema.Resolver, fieldInvoker, error) {
			resolver := schema.FullResolver(func(ctx schema.ResolverContext, v interface{}) (interface{}, error) {
				return returnResult(rf(queryInfo{ctx}, v.(*O)))
			})

			invoker := func(q QueryInfo, o interface{}) interface{} {
				return rf(q, o.(*O))
			}

			return resolver, invoker, nil
		},
	}
}

// FieldQA creates a new field resolved by a method taking a source object, arguments and query info
func FieldQA[R, A, O any](def string, rf func(QueryInfo, *O, *A) Result[R]) *FieldType[O] {
	return &FieldType[O]{
		def:   def,
		rType: typeOf[R](),
		aType: typeOf[A](),
		makeResolver: func(c *buildContext) (schema.Resolver, fieldInvoker, error) {
			bindArgs, err := makeArgBinder[A](c)
			if err != nil {
				return nil, nil, err
			}
			resolver := schema.FullResolver(func(ctx schema.ResolverContext, v interface{}) (interface{}, error) {
				var args A
				if err := bindArgs(queryInfo{ctx}, &args); err != nil {
					var empty R
					return empty, err
				}
				return returnResult(rf(queryInfo{ctx}, v.(*O), &args))
			})

			invoker := func(q QueryInfo, o interface{}) interface{} {
				var args A
				if err := bindArgs(q, &args); err != nil {
					return result.Error[R](err)
				}
				return rf(q, o.(*O), &args)
			}

			return resolver, invoker, nil
		},
	}
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
