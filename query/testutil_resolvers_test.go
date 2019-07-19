// Copyright 2019 HouseCanary, Inc.
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

	"github.com/housecanary/gq/schema"
	"github.com/housecanary/gq/types"
)

type dummyAsyncValue struct {
	sourceResolver schema.Resolver
	ctx            schema.ResolverContext
	v              interface{}
}

func (v dummyAsyncValue) Await(context.Context) (interface{}, error) {
	return v.sourceResolver.Resolve(v.ctx, v.v)
}

func asyncResolver(sourceResolver schema.Resolver) schema.Resolver {
	return schema.FullResolver(func(ctx schema.ResolverContext, v interface{}) (result interface{}, err error) {
		return dummyAsyncValue{sourceResolver, ctx, v}, nil
	})
}

func stringResolver(value string) schema.Resolver {
	return schema.SimpleResolver(func(v interface{}) (interface{}, error) {
		return types.NewString(value), nil
	})
}

func errorResolver(value error) schema.Resolver {
	return schema.SimpleResolver(func(v interface{}) (interface{}, error) {
		return nil, value
	})
}

func panicResolver(value error) schema.Resolver {
	return schema.SimpleResolver(func(v interface{}) (interface{}, error) {
		panic(value)
	})
}
