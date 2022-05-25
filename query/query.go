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

	jsonstream "github.com/json-iterator/go"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser"
	"github.com/housecanary/gq/schema"
)

// An ExecutionListener can be used to trace the execution of a query.
// It is useful for logging, or scheduling work on idle
type ExecutionListener interface {
	// Notifies that execution is entering the selection of the specified field.
	//
	// If this method returns a ResolveCompleteCallback it will be called after
	// the value is resolved, with the result of resolving the value.
	//
	// If the error return value is non-nil, resolution of this field is skipped,
	// and the error is returned to the client.  The ResultListener is not invoked
	// in this case.
	NotifyResolve(queryField *ast.Field, schemaField *schema.FieldDescriptor) (ResolveCompleteCallback, error)

	// Notifies the listener that all synchronously resolvable query nodes have
	// been processed, and awaiting the results is about to begin.  This is a good
	// time for a listener to schedule batch loads.
	NotifyIdle()

	// Notifies the listener that an error occurred executing the query.
	NotifyError(err error)
}

// A ResolveCompleteCallback is used by ExecutionListener.  See above.
type ResolveCompleteCallback func(interface{}, error) error

// A BaseExecutionListener implements ExecutionListener to do nothing
type BaseExecutionListener struct{}

// A PreparedQuery is a compiled query that can be executed given a root value and
// query context
type PreparedQuery struct {
	root selector
}

// Execute runs this query, and returns the serialized results.
func (q *PreparedQuery) Execute(ctx context.Context, rootValue interface{}, variables Variables, listener ExecutionListener) []byte {
	if listener == nil {
		listener = BaseExecutionListener{}
	}
	cc := acquireJSONCollectorContext()
	collector := &vJSONCollector{cc: cc}
	cf := q.root.apply(&exeContext{ctx, listener, variables, nil}, rootValue, collector)
	for cont := cf; cont != nil; {
		listener.NotifyIdle()
		cont = cont()
	}

	stream := streamPool.BorrowStream(nil)

	stream.WriteObjectStart()
	stream.WriteObjectField("data")
	errors, ok := collector.serializeJSON(stream, 0)
	if !ok {
		stream.Reset(nil)
		stream.WriteObjectStart()
		stream.WriteObjectField("data")
		stream.WriteNil()
	}
	serializeErrors(stream, errors)
	stream.WriteObjectEnd()

	collector.release()
	cc.release()

	buf := stream.Buffer()
	content := make([]byte, len(buf))
	copy(content, buf)
	streamPool.ReturnStream(stream)
	return content
}

// Batch contains a batch of queries that execute together using the same context and listener
type Batch struct {
	queries    []*PreparedQuery
	variables  []Variables
	rootValues []interface{}
}

// Add adds a query to this batch
func (b *Batch) Add(q *PreparedQuery, rootValue interface{}, variables Variables) {
	b.queries = append(b.queries, q)
	b.variables = append(b.variables, variables)
	b.rootValues = append(b.rootValues, rootValue)
}

// Execute executes the entire batch of queries
func (b *Batch) Execute(ctx context.Context, listener ExecutionListener) [][]byte {
	// NOTE: This code contains a good deal of duplication with PreparedQuery.Execute
	// Need to consider if this common code can be factored out.
	if listener == nil {
		listener = BaseExecutionListener{}
	}

	// Start execution of all queries onto a consolidated worklist.  This will
	// let us group loads across all queries in the batch.
	var deferred worklist
	collectors := make([]*vJSONCollector, len(b.queries))
	for i, q := range b.queries {
		cc := acquireJSONCollectorContext()
		defer cc.release()
		collector := &vJSONCollector{cc: cc}
		collectors[i] = collector
		rootValue := b.rootValues[i]
		variables := b.variables[i]
		deferred.Add(q.root.apply(&exeContext{ctx, listener, variables, nil}, rootValue, collector))
	}

	// Drain the worklist
	if deferred != nil {
		for cont := deferred.Continue; cont != nil; {
			listener.NotifyIdle()
			cont = cont()
		}
	}

	// Prepare the results
	results := make([][]byte, len(b.queries))
	for i, collector := range collectors {
		stream := streamPool.BorrowStream(nil)

		stream.WriteObjectStart()
		stream.WriteObjectField("data")
		errors, ok := collector.serializeJSON(stream, 0)
		if !ok {
			stream.Reset(nil)
			stream.WriteObjectStart()
			stream.WriteObjectField("data")
			stream.WriteNil()
		}
		serializeErrors(stream, errors)
		stream.WriteObjectEnd()

		collector.release()

		buf := stream.Buffer()
		results[i] = make([]byte, len(buf))
		copy(results[i], buf)
		streamPool.ReturnStream(stream)
	}

	return results
}

// PrepareQuery parses the supplied query text, and compiles it to a PreparedQuery which can then
// be executed many times.  If the supplied query is invalid, nil and an error describing the problem
// are returned.
func PrepareQuery(query string, operationName string, schema *schema.Schema) (*PreparedQuery, error) {
	ast, parseErr := parser.ParseQuery(query)
	if parseErr != nil {
		return nil, parseErr
	}

	opsByName := make(map[string]bool)
	hasAnon := false
	hasNamed := false
	for _, op := range ast.OperationDefinitions {
		if _, ok := opsByName[op.Name]; ok {
			return nil, fmt.Errorf("Operation %s is defined multiple times", op.Name)
		}

		if op.Name == "" {
			hasAnon = true
		} else {
			hasNamed = true
		}
		opsByName[op.Name] = true
	}

	if hasAnon && hasNamed {
		return nil, fmt.Errorf("Cannot mix named and anonymous operations")
	}

	op := ast.LookupQueryOperation(operationName)
	if op == nil {
		return nil, fmt.Errorf("Operation %s does not exist", operationName)
	}

	cc := &compileContext{schema, ast, 0, 0}
	typ := schema.QueryType
	sel, err := buildObjectSelector(cc, typ, op.SelectionSet)
	if err != nil {
		return nil, err
	}
	return &PreparedQuery{sel}, err
}

func serializeErrors(stream *jsonstream.Stream, errors []gqlError) {
	if len(errors) == 0 {
		return
	}
	stream.WriteMore()
	stream.WriteObjectField("errors")
	stream.WriteArrayStart()
	for i, e := range errors {
		if i != 0 {
			stream.WriteMore()
		}
		stream.WriteObjectStart()
		stream.WriteObjectField("message")
		stream.WriteString(e.Error())
		stream.WriteMore()
		stream.WriteObjectField("path")
		stream.WriteArrayStart()
		for pi, pe := range e.path {
			if pi != 0 {
				stream.WriteMore()
			}
			switch pv := pe.(type) {
			case string:
				stream.WriteString(pv)
			case int:
				stream.WriteInt(pv)
			}
		}
		stream.WriteArrayEnd()

		if e.row > 0 && e.col > 0 {
			stream.WriteMore()
			stream.WriteObjectField("locations")
			stream.WriteArrayStart()
			stream.WriteObjectStart()
			stream.WriteObjectField("line")
			stream.WriteInt(e.row)
			stream.WriteMore()
			stream.WriteObjectField("column")
			stream.WriteInt(e.col)
			stream.WriteObjectEnd()
			stream.WriteArrayEnd()
		}
		stream.WriteObjectEnd()
	}
	stream.WriteArrayEnd()
}

// NotifyResolve implements ExecutionListener
func (BaseExecutionListener) NotifyResolve(queryField *ast.Field, schemaField *schema.FieldDescriptor) (ResolveCompleteCallback, error) {
	return nil, nil
}

// NotifyIdle implements ExecutionListener
func (BaseExecutionListener) NotifyIdle() {}

// NotifyError implements ExecutionListener
func (BaseExecutionListener) NotifyError(err error) {}
