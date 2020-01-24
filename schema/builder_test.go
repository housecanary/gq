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

package schema_test

import (
	"context"
	"fmt"
	"os"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

func ExampleBuilder_object() {
	// Example struct that will be the root object of the query
	type Foo struct {
		barField string
	}

	b := schema.NewBuilder()

	b.AddScalarType("String", nil, nil, nil)
	b.AddDirectiveDefinition("exampleDirective", schema.DirectiveLocationObject, schema.DirectiveLocationFieldDefinition).AddArgument("options", &ast.ListType{Of: &ast.SimpleType{Name: "String"}}, nil)

	// Register an object type
	ot := b.AddObjectType("Foo")
	ot.SetDescription("Foo is the root query type")
	ot.AddDirective("exampleDirective").AddArgument("options", ast.ArrayValue{V: []ast.Value{ast.StringValue{V: "Test"}}})

	// Add a field to the object with a resolver that accesses data from the container
	fd := ot.AddField("bar", &ast.SimpleType{Name: "String"}, schema.SimpleResolver(func(container interface{}) (interface{}, error) {
		return container.(*Foo).barField, nil
	}))
	fd.AddDirective("exampleDirective").AddArgument("options", ast.ArrayValue{V: []ast.Value{ast.StringValue{V: "Test2"}}})
	fd.SetDescription("bar contains ... data")

	// Add a field to the object with a resolver that accesses arguments
	fd = ot.AddField("hello", &ast.SimpleType{Name: "String"}, schema.FullResolver(func(ctx schema.ResolverContext, container interface{}) (interface{}, error) {
		v, err := ctx.GetArgumentValue("name")
		if err != nil {
			return nil, err
		}
		return fmt.Sprintf("Hello %v", v), nil
	}))

	ad := fd.AddArgument("name", &ast.SimpleType{Name: "String"}, ast.StringValue{V: "World"})
	ad.SetDescription("The name of the person to say hello to")

	fd.SetDescription("hello says hello")

	s := b.MustBuild("Foo")
	s.WriteDefinition(os.Stdout)
	// Output:
	// schema {
	//   query: Foo
	//
	//   directive @exampleDirective(
	//     options: [String]
	//   ) on OBJECT FIELD_DEFINITION
	//
	//   "Foo is the root query type"
	//   object Foo @exampleDirective(options: ["Test"]) {
	//     "bar contains ... data"
	//     bar: String @exampleDirective(options: ["Test2"])
	//
	//     "hello says hello"
	//     hello (
	//       "The name of the person to say hello to"
	//       name: String = "World"
	//     ): String
	//   }
	// }
}

func ExampleBuilder_interface() {
	type Foo struct {
		barField string
	}

	b := schema.NewBuilder()
	b.AddScalarType("String", nil, nil, nil) // Add non-working string scalar for example purposes

	ot := b.AddObjectType("Foo")
	ot.SetDescription("Foo is the root query type")
	ot.Implements("Helloer")

	fd := ot.AddField("bar", &ast.SimpleType{Name: "String"}, schema.SimpleResolver(func(container interface{}) (interface{}, error) {
		return container.(*Foo).barField, nil
	}))
	fd.SetDescription("bar contains ... data")

	fd = ot.AddField("hello", &ast.SimpleType{Name: "String"}, schema.FullResolver(func(ctx schema.ResolverContext, container interface{}) (interface{}, error) {
		v, err := ctx.GetArgumentValue("name")
		if err != nil {
			return nil, err
		}
		return fmt.Sprintf("Hello %v", v), nil
	}))

	ad := fd.AddArgument("name", &ast.SimpleType{Name: "String"}, ast.StringValue{V: "World"})
	ad.SetDescription("The name of the person to say hello to")

	fd.SetDescription("hello says hello")

	it := b.AddInterfaceType("Helloer", func(ctx context.Context, v interface{}) (interface{}, string) {
		return v, "Foo"
	})
	it.SetDescription("Objects that know how to say hello")

	ifd := it.AddField("hello", &ast.SimpleType{Name: "String"})

	ad = ifd.AddArgument("name", &ast.SimpleType{Name: "String"}, ast.StringValue{V: "World"})
	ad.SetDescription("The name of the person to say hello to")

	ifd.SetDescription("hello says hello")

	s := b.MustBuild("Foo")
	s.WriteDefinition(os.Stdout)
	// Output:
	// schema {
	//   query: Foo
	//
	//   "Foo is the root query type"
	//   object Foo implements & Helloer {
	//     "bar contains ... data"
	//     bar: String
	//
	//     "hello says hello"
	//     hello (
	//       "The name of the person to say hello to"
	//       name: String = "World"
	//     ): String
	//   }
	//
	//   "Objects that know how to say hello"
	//   interface Helloer {
	//     "hello says hello"
	//     hello (
	//       "The name of the person to say hello to"
	//       name: String = "World"
	//     ): String
	//   }
	// }
}

func ExampleBuilder_union() {
	b := schema.NewBuilder()

	b.AddObjectType("Foo")

	b.AddObjectType("Bar")

	ut := b.AddUnionType("FooOrBar", []string{"Foo", "Bar"}, func(ctx context.Context, v interface{}) (interface{}, string) {
		return nil, ""
	})
	ut.SetDescription("A Foo or a Bar")

	s := b.MustBuild("Foo")
	s.WriteDefinition(os.Stdout)
	// Output:
	// schema {
	//   query: Foo
	//
	//   object Bar
	//
	//   object Foo
	//
	//   "A Foo or a Bar"
	//   union FooOrBar = | Bar | Foo
	// }
}

func ExampleBuilder_enum() {
	b := schema.NewBuilder()

	type MyEnumType string

	b.AddObjectType("Foo")

	et := b.AddEnumType("Enm", func(ctx context.Context, v interface{}) (schema.LiteralValue, error) {
		enmVal := v.(MyEnumType)
		if enmVal == MyEnumType("") {
			return nil, nil
		}
		return schema.LiteralString(enmVal), nil
	}, func(ctx context.Context, v schema.LiteralValue) (interface{}, error) {
		if v == nil {
			return MyEnumType(""), nil
		}

		if s, ok := v.(schema.LiteralString); ok {
			return MyEnumType(s), nil
		}

		return nil, fmt.Errorf("Invalid literal value: not a string")
	}, nil)

	et.AddValue("VALUE1").SetDescription("enum value 1")
	et.AddValue("VALUE2").SetDescription("enum value 2")
	et.AddValue("VALUE3").SetDescription("enum value 3")

	et.SetDescription("Enumerated value")

	s := b.MustBuild("Foo")
	s.WriteDefinition(os.Stdout)
	// Output:
	// schema {
	//   query: Foo
	//
	//   "Enumerated value"
	//   enum Enm {
	//     "enum value 1"
	//     VALUE1
	//
	//     "enum value 2"
	//     VALUE2
	//
	//     "enum value 3"
	//     VALUE3
	//   }
	//
	//   object Foo
	// }
}

func ExampleBuilder_scalar() {
	b := schema.NewBuilder()

	type Date string

	b.AddObjectType("Foo")

	st := b.AddScalarType("Date", func(ctx context.Context, v interface{}) (schema.LiteralValue, error) {
		dateVal := v.(Date)
		if dateVal == Date("") {
			return nil, nil
		}
		return schema.LiteralString(dateVal), nil
	}, func(ctx context.Context, v schema.LiteralValue) (interface{}, error) {
		if v == nil {
			return Date(""), nil
		}

		if s, ok := v.(schema.LiteralString); ok {
			return Date(s), nil
		}

		return nil, fmt.Errorf("Invalid literal value: not a string")
	}, nil)

	st.SetDescription("A date encoded as yyyy-mm-ddd")

	s := b.MustBuild("Foo")
	s.WriteDefinition(os.Stdout)
	// Output:
	// schema {
	//   query: Foo
	//
	//   "A date encoded as yyyy-mm-ddd"
	//   scalar Date
	//
	//   object Foo
	// }
}
