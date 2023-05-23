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
	"reflect"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
	"github.com/housecanary/gq/schema/ts/result"
)

// An ObjectType represents a GQL object type
type ObjectType[O any] struct {
	def        string
	fields     []objectFieldType
	implements []reflect.Type
}

// NewObjectType creates a new ObjectType. See example for full details.
func NewObjectType[O any](mod *Module, def string) *ObjectType[O] {
	ot := &ObjectType[O]{
		def: def,
	}
	mod.addType(&objectTypeBuilder[O]{ot: ot})
	return ot
}

// NewInstance makes a new instance of the struct backing this ObjectType
func (ot *ObjectType[O]) NewInstance() *O {
	var o O
	return &o
}

// AddField adds a new field to the given object type
func AddField[R, O any, F func(o *O) Result[R]](objectType *ObjectType[O], fieldDefinition string, resolverFn F) *FieldType[F] {
	ft := &FieldType[F]{
		ResolverFunction: resolverFn,
		def:              fieldDefinition,
		rType:            typeOf[R](),
		makeResolverFn: func(c *buildContext) (schema.Resolver, fieldInvoker, error) {
			resolver := schema.SimpleResolver(func(v interface{}) (interface{}, error) {
				return returnResult(resolverFn(v.(*O)))
			})

			invoker := func(q QueryInfo, o interface{}) interface{} {
				return resolverFn(o.(*O))
			}

			return resolver, invoker, nil
		},
	}

	objectType.fields = append(objectType.fields, ft)
	return ft
}

// AddFieldWithArgs adds a new field with input args to the given object type
func AddFieldWithArgs[R, A, O any, F func(o *O, a *A) Result[R]](objectType *ObjectType[O], fieldDefinition string, resolverFn F) *FieldType[F] {
	ft := &FieldType[F]{
		ResolverFunction: resolverFn,
		def:              fieldDefinition,
		rType:            typeOf[R](),
		aType:            typeOf[A](),
		makeResolverFn: func(c *buildContext) (schema.Resolver, fieldInvoker, error) {
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
				return returnResult(resolverFn(v.(*O), &args))
			})

			invoker := func(q QueryInfo, o interface{}) interface{} {
				var args A
				if err := bindArgs(q, &args); err != nil {
					return result.Error[R](err)
				}
				return resolverFn(o.(*O), &args)
			}

			return resolver, invoker, nil
		},
	}
	objectType.fields = append(objectType.fields, ft)
	return ft
}

type objectFieldType interface {
	buildFieldDef(c *buildContext) (*ast.FieldDefinition, bool, error)
	originalDefinition() string
}
