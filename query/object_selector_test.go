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
	"fmt"
	"testing"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
	"github.com/housecanary/gq/types"
)

func buildFieldDescriptor(resolver schema.Resolver) *schema.FieldDescriptor {
	builder := schema.NewBuilder()
	builder.AddScalarType("String", schema.EncodeScalarMarshaler, func(ctx context.Context, in schema.LiteralValue) (interface{}, error) {
		return types.NewString(string(in.(schema.LiteralString))), nil
	}, stringInputListCreator{})
	qt := builder.AddObjectType("Query")
	qt.AddField("theField", &ast.SimpleType{Name: "String"}, resolver)
	return builder.MustBuild("Query").QueryType.Field("theField")
}

func TestObjectSelectorSimple(t *testing.T) {
	queryField := &ast.Field{
		Alias: "outName",
	}
	schemaField := buildFieldDescriptor(stringResolver("Foo"))
	os := &objectSelector{
		Fields: []*objectSelectorField{
			&objectSelectorField{
				AstField: queryField,
				Sel:      dummySelector{},
				Field:    schemaField,
			},
		},
	}
	col := &testCollector{}
	listener := &assertExecutionListener{
		assertions: []executionListenerAssertion{
			resolveAssertion{queryField: queryField, schemaField: schemaField},
		},
	}
	os.apply(exeContext{
		listener: listener,
	}, struct{}{}, col)
	listener.assertDone()
}

func TestObjectSelectorSimpleSafe(t *testing.T) {
	queryField := &ast.Field{
		Alias: "outName",
	}
	schemaField := buildFieldDescriptor(schema.MarkSafe(stringResolver("Foo")))
	os := &objectSelector{
		Fields: []*objectSelectorField{
			&objectSelectorField{
				AstField: queryField,
				Sel:      dummySelector{},
				Field:    schemaField,
			},
		},
	}
	col := &testCollector{}
	listener := &assertExecutionListener{
		assertions: []executionListenerAssertion{
			resolveAssertion{queryField: queryField, schemaField: schemaField},
		},
	}
	os.apply(exeContext{
		listener: listener,
	}, struct{}{}, col)
	listener.assertDone()
}

func TestObjectSelectorSimpleError(t *testing.T) {
	err := fmt.Errorf("Test error")
	queryField := &ast.Field{
		Alias: "outName",
	}
	schemaField := buildFieldDescriptor(errorResolver(err))
	os := &objectSelector{
		Fields: []*objectSelectorField{
			&objectSelectorField{
				AstField: queryField,
				Sel:      dummySelector{},
				Field:    schemaField,
			},
		},
	}
	col := &testCollector{}
	listener := &assertExecutionListener{
		assertions: []executionListenerAssertion{
			resolveAssertion{queryField: queryField, schemaField: schemaField},
			errorAssertion{err: err},
		},
	}
	os.apply(exeContext{
		listener: listener,
	}, struct{}{}, col)
	listener.assertDone()
}

func TestObjectSelectorSimplePanic(t *testing.T) {
	err := fmt.Errorf("Test error")
	queryField := &ast.Field{
		Alias: "outName",
	}
	schemaField := buildFieldDescriptor(panicResolver(err))
	os := &objectSelector{
		Fields: []*objectSelectorField{
			&objectSelectorField{
				AstField: queryField,
				Sel:      dummySelector{},
				Field:    schemaField,
			},
		},
	}
	col := &testCollector{}
	listener := &assertExecutionListener{
		assertions: []executionListenerAssertion{
			resolveAssertion{queryField: queryField, schemaField: schemaField},
			errorAssertion{err: err},
		},
	}
	os.apply(exeContext{
		listener: listener,
	}, struct{}{}, col)
	listener.assertDone()
}

func TestObjectSelectorSimpleSafeError(t *testing.T) {
	err := fmt.Errorf("Test error")
	queryField := &ast.Field{
		Alias: "outName",
	}
	schemaField := buildFieldDescriptor(schema.MarkSafe(errorResolver(err)))
	os := &objectSelector{
		Fields: []*objectSelectorField{
			&objectSelectorField{
				AstField: queryField,
				Sel:      dummySelector{},
				Field:    schemaField,
			},
		},
	}
	col := &testCollector{}
	listener := &assertExecutionListener{
		assertions: []executionListenerAssertion{
			resolveAssertion{queryField: queryField, schemaField: schemaField},
			errorAssertion{err: err},
		},
	}
	os.apply(exeContext{
		listener: listener,
	}, struct{}{}, col)
	listener.assertDone()
}

func TestObjectSelectorAsync(t *testing.T) {
	queryField := &ast.Field{
		Alias: "outName",
	}
	schemaField := buildFieldDescriptor(asyncResolver(stringResolver("Foo")))
	os := &objectSelector{
		Fields: []*objectSelectorField{
			&objectSelectorField{
				AstField: queryField,
				Sel:      dummySelector{},
				Field:    schemaField,
			},
		},
	}
	col := &testCollector{}
	listener := &assertExecutionListener{
		assertions: []executionListenerAssertion{
			resolveAssertion{queryField: queryField, schemaField: schemaField},
		},
	}
	cont := os.apply(exeContext{
		listener: listener,
	}, struct{}{}, col)
	cont()
	listener.assertDone()
}

func TestObjectSelectorAsyncErr(t *testing.T) {
	err := fmt.Errorf("Test error")
	queryField := &ast.Field{
		Alias: "outName",
	}
	schemaField := buildFieldDescriptor(asyncResolver(errorResolver(err)))
	os := &objectSelector{
		Fields: []*objectSelectorField{
			&objectSelectorField{
				AstField: queryField,
				Sel:      dummySelector{},
				Field:    schemaField,
			},
		},
	}
	col := &testCollector{}
	listener := &assertExecutionListener{
		assertions: []executionListenerAssertion{
			resolveAssertion{queryField: queryField, schemaField: schemaField},
			errorAssertion{err: err},
		},
	}
	cont := os.apply(exeContext{
		listener: listener,
	}, struct{}{}, col)
	cont()
	listener.assertDone()
}

func TestObjectSelectorAsyncPanic(t *testing.T) {
	err := fmt.Errorf("Test error")
	queryField := &ast.Field{
		Alias: "outName",
	}
	schemaField := buildFieldDescriptor(asyncResolver(panicResolver(err)))
	os := &objectSelector{
		Fields: []*objectSelectorField{
			&objectSelectorField{
				AstField: queryField,
				Sel:      dummySelector{},
				Field:    schemaField,
			},
		},
	}
	col := &testCollector{}
	listener := &assertExecutionListener{
		assertions: []executionListenerAssertion{
			resolveAssertion{queryField: queryField, schemaField: schemaField},
			errorAssertion{err: err},
		},
	}
	cont := os.apply(exeContext{
		listener: listener,
	}, struct{}{}, col)
	cont()
	listener.assertDone()
}
