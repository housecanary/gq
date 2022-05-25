package ts

import (
	"context"
	"fmt"
	"reflect"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser"
	"github.com/housecanary/gq/schema"
)

type QueryInfo interface {
	schema.ChildWalker
	GetArgumentValue(name string) (interface{}, error)
}

type Result[T any] interface {
	UnpackResult() (interface{}, error)
}

type FieldType[O any] struct {
	def          string
	rType        reflect.Type
	aType        reflect.Type
	makeResolver func(c *buildContext) (schema.Resolver, error)
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

func Field[R, O any](def string, rf func(*O) Result[R]) *FieldType[O] {
	return &FieldType[O]{
		def:   def,
		rType: typeOf[R](),
		makeResolver: func(c *buildContext) (schema.Resolver, error) {
			return schema.SimpleResolver(func(v interface{}) (interface{}, error) {
				return rf(v.(*O)).UnpackResult()
			}), nil
		},
	}
}

func FieldA[R, A, O any](def string, rf func(*O, *A) Result[R]) *FieldType[O] {
	return &FieldType[O]{
		def:   def,
		rType: typeOf[R](),
		aType: typeOf[A](),
		makeResolver: func(c *buildContext) (schema.Resolver, error) {
			bindArgs, err := makeArgBinder[A](c)
			if err != nil {
				return nil, err
			}
			return schema.FullResolver(func(ctx schema.ResolverContext, v interface{}) (interface{}, error) {
				var args A
				if err := bindArgs(ctx, &args); err != nil {
					var empty R
					return empty, err
				}
				return rf(v.(*O), &args).UnpackResult()
			}), nil
		},
	}
}

func FieldQ[R, O any](def string, rf func(QueryInfo, *O) Result[R]) *FieldType[O] {
	return &FieldType[O]{
		def:   def,
		rType: typeOf[R](),
		makeResolver: func(c *buildContext) (schema.Resolver, error) {
			return schema.FullResolver(func(ctx schema.ResolverContext, v interface{}) (interface{}, error) {
				return rf(ctx, v.(*O)).UnpackResult()
			}), nil
		},
	}
}

func FieldQA[R, A, O any](def string, rf func(QueryInfo, *O, *A) Result[R]) *FieldType[O] {
	return &FieldType[O]{
		def:   def,
		rType: typeOf[R](),
		aType: typeOf[A](),
		makeResolver: func(c *buildContext) (schema.Resolver, error) {
			bindArgs, err := makeArgBinder[A](c)
			if err != nil {
				return nil, err
			}
			return schema.FullResolver(func(ctx schema.ResolverContext, v interface{}) (interface{}, error) {
				var args A
				if err := bindArgs(ctx, &args); err != nil {
					var empty R
					return empty, err
				}
				return rf(ctx, v.(*O), &args).UnpackResult()
			}), nil
		},
	}
}

func FieldP[R, O, P any](def string, mod *ModuleType[P], rf func(P, *O) Result[R]) *FieldType[O] {
	return &FieldType[O]{
		def:   def,
		rType: typeOf[R](),
		makeResolver: func(c *buildContext) (schema.Resolver, error) {
			return schema.ContextResolver(func(ctx context.Context, v interface{}) (interface{}, error) {
				p := mod.GetProvider(ctx)
				return rf(p, v.(*O)).UnpackResult()
			}), nil
		},
	}
}

func FieldPA[R, A, O, P any](def string, mod *ModuleType[P], rf func(P, *O, *A) Result[R]) *FieldType[O] {
	return &FieldType[O]{
		def:   def,
		rType: typeOf[R](),
		aType: typeOf[A](),
		makeResolver: func(c *buildContext) (schema.Resolver, error) {
			bindArgs, err := makeArgBinder[A](c)
			if err != nil {
				return nil, err
			}
			return schema.FullResolver(func(ctx schema.ResolverContext, v interface{}) (interface{}, error) {
				var args A
				if err := bindArgs(ctx, &args); err != nil {
					var empty R
					return empty, err
				}
				p := mod.GetProvider(ctx)
				return rf(p, v.(*O), &args).UnpackResult()
			}), nil
		},
	}
}

func FieldPQ[R, O, P any](def string, mod *ModuleType[P], rf func(P, QueryInfo, *O) Result[R]) *FieldType[O] {
	return &FieldType[O]{
		def:   def,
		rType: typeOf[R](),
		makeResolver: func(c *buildContext) (schema.Resolver, error) {
			return schema.FullResolver(func(ctx schema.ResolverContext, v interface{}) (interface{}, error) {
				p := mod.GetProvider(ctx)
				return rf(p, ctx, v.(*O)).UnpackResult()
			}), nil
		},
	}
}

func FieldPQA[R, A, O, P any](def string, mod *ModuleType[P], rf func(P, QueryInfo, *O, *A) Result[R]) *FieldType[O] {
	return &FieldType[O]{
		def:   def,
		rType: typeOf[R](),
		aType: typeOf[A](),
		makeResolver: func(c *buildContext) (schema.Resolver, error) {
			bindArgs, err := makeArgBinder[A](c)
			if err != nil {
				return nil, err
			}
			return schema.FullResolver(func(ctx schema.ResolverContext, v interface{}) (interface{}, error) {
				var args A
				if err := bindArgs(ctx, &args); err != nil {
					var empty R
					return empty, err
				}
				p := mod.GetProvider(ctx)
				return rf(p, ctx, v.(*O), &args).UnpackResult()
			}), nil
		},
	}
}

type argBinder[A any] func(schema.ResolverContext, *A) error

func makeArgBinder[A any](c *buildContext) (argBinder[A], error) {
	typ := typeOf[A]()
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("invalid arguments type, expected a struct got %v", typ.Kind())
	}

	fields := reflect.VisibleFields(typ)
	binds := make([]func(schema.ResolverContext, reflect.Value) error, 0, len(fields))
	for _, field := range fields {
		ad, _, err := parseStructField(c, field, parser.ParsePartialInputValueDefinition)
		if ad == nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("cannot parse input argument %s: %w", field.Name, err)
		}

		binds = append(binds, func(rc schema.ResolverContext, v reflect.Value) error {
			av, err := rc.GetArgumentValue(ad.Name)
			if err != nil {
				return err
			}
			v.FieldByIndex(field.Index).Set(reflect.ValueOf(av))
			return nil
		})
	}

	return func(rc schema.ResolverContext, a *A) error {
		rv := reflect.ValueOf(a).Elem()
		for _, bind := range binds {
			if err := bind(rc, rv); err != nil {
				return err
			}
		}
		return nil
	}, nil
}
