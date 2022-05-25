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

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/housecanary/gq/query"
	"github.com/housecanary/gq/schema"
)

// A QueryBuilder is used to transform a query text and operation name into a PreparedQuery.
//
// The default implementation recompiles the query each request, a production configuration should
// use some sort of cache to reuse previously compiled queries
type QueryBuilder func(schema *schema.Schema, text string, operationName string) (*query.PreparedQuery, error)

// DefaultQueryBuilder is a simple QueryBuilder that recompiles the query each request.
var DefaultQueryBuilder QueryBuilder = func(schema *schema.Schema, text string, operationName string) (*query.PreparedQuery, error) {
	return query.PrepareQuery(text, operationName, schema)
}

// A QueryExecutor runs a prepared query. Implementations of this may add variables to the context,
// set up callbacks, and set up tracing. This method is deprecated, use a QueryExecutionWrapper instead
type QueryExecutor func(q *query.PreparedQuery, req *http.Request, vars query.Variables, responseHeaders http.Header) []byte

// QueryInfo supplies information about the queries being executed to a QueryExecutionWrapper
type QueryInfo interface {
	GetNQueries() int
	GetQuery(n int) *query.PreparedQuery
	GetVariables(n int) query.Variables
	GetRootObject(n int) interface{}
}

// A QueryExecutionWrapper wraps execution of a query or batch of queries. Implementations of this may add variables to the context,
// set up callbacks, and set up tracing.
type QueryExecutionWrapper func(queryInfo QueryInfo, req *http.Request, responseHeaders http.Header, proceed func(context.Context, query.ExecutionListener))

// A RootObjectProvider is used to create root objects for queries
type RootObjectProvider func(req *http.Request) interface{}

// Default maximum size of request body.
const defaultMaxRequestBodySize = 100 * 1024 // 100kb

var _ http.Handler = &GraphQLHandler{}

// A GraphQLHandler is a http.Handler that fulfills GraphQL requests
type GraphQLHandler struct {
	schema             *schema.Schema
	queryBuilder       QueryBuilder
	queryExecutor      QueryExecutor
	executionWrapper   QueryExecutionWrapper
	rootObjectProvider RootObjectProvider
	maxRequestBodySize int64
	disableGraphiQL    bool
}

// A GraphQLHandlerConfig supplies configuration parameters to NewGraphQLHandler
type GraphQLHandlerConfig struct {
	// Callback to build queries.  Can be used to implement query caching or additional validation.
	QueryBuilder QueryBuilder

	// Callback to execute queries.  Can be used to inject request specific items (loggers, listeners, context variables, etc),
	// as well as for logging
	QueryExecutor QueryExecutor

	// Callback to wrap execution of queries.  Can be used to inject request specific items (loggers, listeners, context variables, etc),
	// as well as for logging
	QueryExecutionWrapper QueryExecutionWrapper

	// Root object to use.
	RootObject interface{}

	// Provider for root objects
	RootObjectProvider RootObjectProvider

	// By default GraphiQL is enabled.  This can be used to disable it.
	DisableGraphiQL bool

	// Max size of request body.  If -1, no limit.  Server will respond with 413
	// if the size is exceeded
	MaxRequestBodySize int64
}

// NewGraphQLHandler creates a new GraphQLHandler with the specified configuration
func NewGraphQLHandler(s *schema.Schema, config *GraphQLHandlerConfig) *GraphQLHandler {
	qb := config.QueryBuilder
	if qb == nil {
		qb = DefaultQueryBuilder
	}

	rop := config.RootObjectProvider
	if rop == nil {
		rop = func(req *http.Request) interface{} {
			return config.RootObject
		}
	}

	qe := config.QueryExecutor
	if qe == nil {
		execWrapper := config.QueryExecutionWrapper
		qe = func(q *query.PreparedQuery, req *http.Request, vars query.Variables, responseHeaders http.Header) []byte {
			root := rop(req)
			var result []byte
			if execWrapper != nil {
				execWrapper(singleQueryInfo{q, vars, root}, req, responseHeaders, func(ctx context.Context, ql query.ExecutionListener) {
					result = q.Execute(ctx, root, vars, ql)
				})
			} else {
				result = q.Execute(req.Context(), root, vars, nil)
			}
			return result
		}
	}

	maxRequestBodySize := config.MaxRequestBodySize
	if maxRequestBodySize == 0 {
		maxRequestBodySize = defaultMaxRequestBodySize
	}

	return &GraphQLHandler{
		schema:             s,
		queryBuilder:       qb,
		queryExecutor:      qe,
		executionWrapper:   config.QueryExecutionWrapper,
		rootObjectProvider: rop,
		maxRequestBodySize: maxRequestBodySize,
		disableGraphiQL:    config.DisableGraphiQL,
	}
}

func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	qs := req.URL.Query()
	switch req.Method {
	case http.MethodGet:
		msg := graphQLRequest{}
		msg.Query = qs.Get("query")
		msg.OperationName = qs.Get("operationName")
		msg.Variables = json.RawMessage(qs.Get("variables"))
		h.executeSingle(w, req, &msg)
	case http.MethodPost:
		body := req.Body
		var err error
		if body != nil {
			defer body.Close()
			var data []byte
			if h.maxRequestBodySize != -1 {
				limitReader := &io.LimitedReader{R: body, N: h.maxRequestBodySize}
				data, err = ioutil.ReadAll(body)
				if limitReader.N <= 0 {
					h.writeError(http.StatusRequestEntityTooLarge, "Request body too large", w)
					return
				}
			} else {
				data, err = ioutil.ReadAll(body)
			}

			if err != nil {
				// This is inevitably a network error.  Try to write a message
				// to the client just in case, but no need to log.
				h.writeError(http.StatusBadRequest, "Request body too large", w)
				return
			}

			switch req.Header.Get("Content-Type") {
			case "application/json":
				itr := iterPool.BorrowIterator(data)
				defer iterPool.ReturnIterator(itr)

				next := itr.WhatIsNext()
				if next == jsoniter.ArrayValue {
					var requests []*graphQLRequest
					itr.ReadVal(&requests)
					if itr.Error != nil {
						h.writeError(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", itr.Error), w)
						return
					}
					h.executeBatch(w, req, requests)
				} else if next == jsoniter.ObjectValue {
					var msg graphQLRequest
					itr.ReadVal(&msg)
					if itr.Error != nil {
						h.writeError(http.StatusBadRequest, fmt.Sprintf("Bad request: %v", itr.Error), w)
						return
					}

					if len(msg.Extensions.VariablesList) > 0 {
						if len(msg.Variables) != 0 {
							h.writeError(http.StatusBadRequest, "Bad request: cannot use variables and variablesList extension together", w)
							return
						}
						h.executeBatch(w, req, msg.splitVariablesList())
					} else {
						h.executeSingle(w, req, &msg)
					}
				} else {
					h.writeError(http.StatusBadRequest, "Unparsable response body: must be an object or array", w)
				}
			case "application/graphql":
				msg := graphQLRequest{
					Query: string(data),
				}
				h.executeSingle(w, req, &msg)
			default:
				h.writeError(http.StatusUnsupportedMediaType, fmt.Sprintf("Unsupported media type %s", req.Header.Get("Content-Type")), w)
			}
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func serializeError(err error) []byte {
	errObj := make(map[string]interface{})
	errObj["message"] = err.Error()
	b, _ := json.Marshal(struct {
		Errors []interface{} `json:"errors"`
	}{
		Errors: []interface{}{
			struct {
				Message string `json:"message"`
			}{Message: err.Error()},
		},
	})

	return b
}

func (h *GraphQLHandler) writeError(statusCode int, msg string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	if msg != "" {
		w.Write([]byte(msg))
	}
}

func (h *GraphQLHandler) executeSingle(w http.ResponseWriter, req *http.Request, request *graphQLRequest) {
	if request.Query == "" {
		h.writeSingleRequestResult(w, req, request, []byte(""))
		return
	}

	vars, err := query.NewVariablesFromJSON(request.Variables)
	if err != nil {
		h.writeSingleRequestResult(w, req, request, serializeError(err))
		return
	}
	q, err := h.queryBuilder(h.schema, request.Query, request.OperationName)
	if err != nil {
		h.writeSingleRequestResult(w, req, request, serializeError(err))
		return
	}

	result := h.queryExecutor(q, req, vars, w.Header())
	h.writeSingleRequestResult(w, req, request, result)
}

func (h *GraphQLHandler) executeBatch(w http.ResponseWriter, req *http.Request, requests []*graphQLRequest) {
	results := make([][]byte, len(requests))
	toExecute := make([]batchQueryItem, 0, len(requests))
	for i, request := range requests {
		vars, err := query.NewVariablesFromJSON(request.Variables)
		if err != nil {
			results[i] = serializeError(err)
			continue
		}
		q, err := h.queryBuilder(h.schema, request.Query, request.OperationName)
		if err != nil {
			results[i] = serializeError(err)
			continue
		}
		toExecute = append(toExecute, batchQueryItem{
			query:       q,
			vars:        vars,
			rootObject:  h.rootObjectProvider(req),
			resultIndex: i,
		})
	}

	batch := &query.Batch{}
	for _, q := range toExecute {
		batch.Add(q.query, q.rootObject, q.vars)
	}

	var batchResults [][]byte
	if h.executionWrapper != nil {
		h.executionWrapper(batchQueryInfo(toExecute), req, w.Header(), func(ctx context.Context, ql query.ExecutionListener) {
			batchResults = batch.Execute(ctx, ql)
		})
	} else {
		batchResults = batch.Execute(nil, nil)
	}

	for i, qi := range toExecute {
		results[qi.resultIndex] = batchResults[i]
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.Write(startArray)
	for i, result := range results {
		if i != 0 {
			w.Write(comma)
		}
		w.Write(result)
	}
	w.Write(endArray)
}

func (h *GraphQLHandler) writeSingleRequestResult(w http.ResponseWriter, req *http.Request, gqlRequest *graphQLRequest, body []byte) {
	isGraphiQL := !h.disableGraphiQL &&
		req.Method == http.MethodGet &&
		!hasParam(req, "raw") &&
		(strings.Contains(req.Header.Get("Accept"), "text/html") ||
			strings.Contains(req.Header.Get("Accept"), "*/*"))

	isPretty := isGraphiQL || hasParam(req, "pretty")

	if isPretty {
		buf := &bytes.Buffer{}
		err := json.Indent(buf, []byte(body), "", "  ")
		if err == nil { // if the response cannot be prettified, just use the unprettied version
			body = buf.Bytes()
		}
	}

	if isGraphiQL {
		w.Header().Set("Content-Type", "text/html;charset=utf-8")
		graphiQLTemplate.Execute(w, &graphiQLParams{
			GraphiQLVersion: "0.12.0",
			Query:           gqlRequest.Query,
			OperationName:   gqlRequest.OperationName,
			Variables:       string(gqlRequest.Variables),
			Response:        string(body),
		})
	} else {
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		w.Write(body)
	}
}

func hasParam(req *http.Request, name string) bool {
	_, ok := req.URL.Query()[name]
	return ok
}

type graphQLRequest struct {
	Query         string          `json:"query"`
	Variables     json.RawMessage `json:"variables"`
	OperationName string          `json:"operationName"`
	Extensions    struct {
		VariablesList []json.RawMessage `json:"variablesList"`
	}
}

type graphQLBatchRequest struct {
	Query         string `json:"query"`
	OperationName string `json:"operationName"`
}

func (r *graphQLRequest) splitVariablesList() []*graphQLRequest {
	result := make([]*graphQLRequest, len(r.Extensions.VariablesList))
	for i, v := range r.Extensions.VariablesList {
		result[i] = &graphQLRequest{
			Query:         r.Query,
			Variables:     v,
			OperationName: r.OperationName,
		}
	}
	return result
}

var startArray = []byte("[")
var endArray = []byte("]")
var comma = []byte(",")

type batchQueryItem struct {
	query       *query.PreparedQuery
	vars        query.Variables
	rootObject  interface{}
	resultIndex int
}

type batchQueryInfo []batchQueryItem

func (b batchQueryInfo) GetNQueries() int {
	return len(b)
}

func (b batchQueryInfo) GetQuery(n int) *query.PreparedQuery {
	return b[n].query
}

func (b batchQueryInfo) GetVariables(n int) query.Variables {
	return b[n].vars
}

func (b batchQueryInfo) GetRootObject(n int) interface{} {
	return b[n].rootObject
}

type singleQueryInfo struct {
	query      *query.PreparedQuery
	vars       query.Variables
	rootObject interface{}
}

func (i singleQueryInfo) GetNQueries() int {
	return 1
}

func (i singleQueryInfo) GetQuery(n int) *query.PreparedQuery {
	return i.query
}

func (i singleQueryInfo) GetVariables(n int) query.Variables {
	return i.vars
}

func (i singleQueryInfo) GetRootObject(n int) interface{} {
	return i.rootObject
}

var iterPool = jsoniter.NewIterator(
	jsoniter.Config{
		UseNumber: false,
	}.Froze(),
).Pool()
