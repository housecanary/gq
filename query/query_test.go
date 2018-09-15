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
	"testing"

	"github.com/housecanary/gq/schema"
	"github.com/housecanary/gq/types"

	"github.com/housecanary/gq/ast"
)

type Query struct{}

func FooResolver(v interface{}) (interface{}, error) {
	return types.NewString("bar"), nil
}

func WithArgResolver(ctx schema.ResolverContext, v interface{}) (interface{}, error) {
	arg, err := ctx.GetArgumentValue("in")
	return arg.(types.String), err
}

type DummyAsyncValue struct {
	V interface{}
}

func (v DummyAsyncValue) Await(context.Context) (interface{}, error) {
	return v.V, nil
}

type ErrorAsyncValue struct {
	E error
}

func (v ErrorAsyncValue) Await(context.Context) (interface{}, error) {
	return nil, v.E
}

func AsyncFooResolver(v interface{}) (interface{}, error) {
	return DummyAsyncValue{types.NewString("bar")}, nil
}

func AsyncErrorResolver(v interface{}) (interface{}, error) {
	return ErrorAsyncValue{fmt.Errorf("Test Error")}, nil
}

func FooListResolver(v interface{}) (interface{}, error) {
	return schema.ListOf(
		types.NewString("foo"),
		types.NewString("bar"),
		types.NewString("bang"),
		types.NewString("bleet"),
		types.NewString("frob"),
		types.NewString("splat"),
		types.NewString("baz"),
	), nil
}

type DummyQueryExecutionListener struct {
	BaseExecutionListener
	idleCount int
}

func (c *DummyQueryExecutionListener) NotifyIdle() {
	c.idleCount++
}

func runQuery(t *testing.T, query string, vars Variables, expectedData string, expectedIdleCount int) {
	builder := schema.NewBuilder()
	builder.AddScalarType("String", schema.EncodeScalarMarshaler, func(ctx context.Context, in schema.LiteralValue) (interface{}, error) {
		return types.NewString(string(in.(schema.LiteralString))), nil
	})
	qt := builder.AddObjectType("Query")
	qt.AddField("foo", &ast.SimpleType{Name: "String"}, schema.SimpleResolver(FooResolver))
	fooWithArg := qt.AddField("fooWithArg", &ast.SimpleType{Name: "String"}, schema.FullResolver(WithArgResolver))
	fooWithArg.AddArgument("in", &ast.SimpleType{Name: "String"}, nil)
	qt.AddField("asyncFoo", &ast.SimpleType{Name: "String"}, schema.SimpleResolver(AsyncFooResolver))
	qt.AddField("asyncFooError", &ast.SimpleType{Name: "String"}, schema.SimpleResolver(AsyncErrorResolver))
	qt.AddField("fooList", &ast.ListType{Of: &ast.SimpleType{Name: "String"}}, schema.SimpleResolver(FooListResolver))
	s := builder.MustBuild("Query")
	q, err := PrepareQuery(query, "", s)
	if err != nil {
		panic(err)
	}

	listener := &DummyQueryExecutionListener{}
	result := string(q.Execute(context.Background(), &Query{}, vars, listener))
	if result != expectedData {
		t.Errorf("Expected result %v, got %v", expectedData, result)
	}

	if listener.idleCount != expectedIdleCount {
		t.Errorf("Expected idle count %v, got %v", expectedIdleCount, listener.idleCount)
	}

}
func TestSimpleQuery(t *testing.T) {
	runQuery(t, "{foo}", nil, `{"data":{"foo":"bar"}}`, 0)
}

func TestArgQuery(t *testing.T) {
	runQuery(t, `{fooWithArg(in: "a")}`, nil, `{"data":{"fooWithArg":"a"}}`, 0)
}

func TestArgQueryVars(t *testing.T) {
	runQuery(t, `query($a: String){fooWithArg(in: $a)}`, Variables{"a": schema.LiteralString("aVar")}, `{"data":{"fooWithArg":"aVar"}}`, 0)
}

func TestAsyncQuery(t *testing.T) {
	runQuery(t, "{asyncFoo}", nil, `{"data":{"asyncFoo":"bar"}}`, 1)
}

func TestAsyncError(t *testing.T) {
	runQuery(t, "{asyncFooError}", nil, `{"data":{"asyncFooError":null},"errors":[{"message":"Test Error","path":["asyncFooError"],"locations":[{"line":2,"column":2}]}]}`, 1)
}

func TestList(t *testing.T) {
	runQuery(t, "{fooList}", nil, `{"data":{"fooList":["foo","bar","bang","bleet","frob","splat","baz"]}}`, 0)

}

func BenchmarkSimpleQuery(b *testing.B) {
	builder := schema.NewBuilder()
	builder.AddScalarType("String", schema.EncodeScalarMarshaler, nil)
	builder.AddObjectType("Query").AddField("foo", &ast.SimpleType{Name: "String"}, schema.SimpleResolver(FooResolver))
	s := builder.MustBuild("Query")
	q, err := PrepareQuery("{foo}", "", s)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		q.Execute(nil, &Query{}, nil, nil)
	}
}

func BenchmarkAsyncQuery(b *testing.B) {
	builder := schema.NewBuilder()
	builder.AddScalarType("String", schema.EncodeScalarMarshaler, nil)
	builder.AddObjectType("Query").AddField("foo", &ast.SimpleType{Name: "String"}, schema.SimpleResolver(AsyncFooResolver))
	s := builder.MustBuild("Query")
	q, err := PrepareQuery("{foo}", "", s)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		q.Execute(nil, &Query{}, nil, nil)
	}
}
