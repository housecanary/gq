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
	"encoding/json"
	"fmt"
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

// A QueryExecutor runs a prepared query.  Implementations of this may add variables to the context,
// set up callbacks, and set up tracing.
type QueryExecutor func(q *query.PreparedQuery, req *http.Request, vars query.Variables, responseHeaders http.Header) []byte

// A BatchQueryExecutor runs a batch of prepared queries.  Implementations of this may add variables to the context,
// set up callbacks, and set up tracing.
type BatchQueryExecutor func(q []BatchQueryItem, req *http.Request, responseHeaders http.Header) [][]byte

// A BatchQueryItem is a single item in a batch of queries
type BatchQueryItem struct {
	Query       *query.PreparedQuery
	Vars        query.Variables
	resultIndex int
}

var _ http.Handler = &GraphQLHandler{}

// A GraphQLHandler is a http.Handler that fulfills GraphQL requests
type GraphQLHandler struct {
	schema             *schema.Schema
	queryBuilder       QueryBuilder
	queryExecutor      QueryExecutor
	batchQueryExecutor BatchQueryExecutor
	disableGraphiQL    bool
}

// A GraphQLHandlerConfig supplies configuration parameters to NewGraphQLHandler
type GraphQLHandlerConfig struct {
	// Callback to build queries.  Can be used to implement query caching or additional validation.
	QueryBuilder QueryBuilder

	// Callback to execute queries.  Can be used to inject request specific items (loggers, listeners, context variables, etc),
	// as well as for logging
	QueryExecutor QueryExecutor

	// Callback to execute batched queries.  Can be used to inject request specific items (loggers, listeners, context variables, etc),
	// as well as for logging
	BatchQueryExecutor BatchQueryExecutor

	// Root object to use.
	RootObject interface{}

	// By default GraphiQL is enabled.  This can be used to disable it.
	DisableGraphiQL bool
}

// NewGraphQLHandler creates a new GraphQLHandler with the specified configuration
func NewGraphQLHandler(s *schema.Schema, config *GraphQLHandlerConfig) *GraphQLHandler {
	qb := config.QueryBuilder
	if qb == nil {
		qb = DefaultQueryBuilder
	}

	qe := config.QueryExecutor
	if qe == nil {
		qe = func(q *query.PreparedQuery, req *http.Request, vars query.Variables, responseHeaders http.Header) []byte {
			return q.Execute(nil, config.RootObject, vars, nil)
		}
	}

	return &GraphQLHandler{
		schema:             s,
		queryBuilder:       qb,
		queryExecutor:      qe,
		batchQueryExecutor: config.BatchQueryExecutor,
		disableGraphiQL:    config.DisableGraphiQL,
	}
}

func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	qs := req.URL.Query()
	msg := graphQLRequest{}
	msg.Query = qs.Get("query")
	msg.OperationName = qs.Get("operationName")
	msg.Variables = json.RawMessage(qs.Get("variables"))
	switch req.Method {
	case http.MethodGet:
		break
	case http.MethodPost:
		body := req.Body
		if body != nil {
			defer body.Close()
			data, err := ioutil.ReadAll(body)
			if err != nil {
				// This is inevitably a network error.  Try to write a message
				// to the client just in case, but no need to log.
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Body truncated"))
				return
			}

			switch req.Header.Get("Content-Type") {
			case "application/json":
				itr := iterPool.BorrowIterator(data)
				defer func() { iterPool.ReturnIterator(itr) }()
				next := itr.WhatIsNext()

				if next == jsoniter.ArrayValue {
					var requests []*graphQLRequest
					itr.ReadVal(&requests)

					if itr.Error != nil {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(fmt.Sprintf("Bad request: %v", itr.Error)))
						return
					}

					h.executeBatch(w, req, requests)
					return
				}

				itr.ReadVal(&msg)

				if itr.Error != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("Bad request: %v", itr.Error)))
					return
				}
			case "application/graphql":
				msg.Query = string(data)
			default:
				w.WriteHeader(http.StatusUnsupportedMediaType)
				w.Write([]byte(fmt.Sprintf("Unsupported media type %s", req.Header.Get("Content-Type"))))
				return
			}
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if msg.Query == "" {
		h.writeResult(w, req, &msg, []byte(""))
		return
	}

	vars, err := query.NewVariablesFromJSON(msg.Variables)
	if err != nil {
		h.writeResult(w, req, &msg, serializeError(err))
		return
	}
	q, err := h.queryBuilder(h.schema, msg.Query, msg.OperationName)
	if err != nil {
		h.writeResult(w, req, &msg, serializeError(err))
		return
	}

	result := h.queryExecutor(q, req, vars, w.Header())
	h.writeResult(w, req, &msg, result)
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

var startArray = []byte("[")
var endArray = []byte("]")
var comma = []byte(",")

func (h *GraphQLHandler) executeBatch(w http.ResponseWriter, req *http.Request, requests []*graphQLRequest) {
	results := make([][]byte, len(requests))
	toExecute := make([]BatchQueryItem, 0, len(requests))
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
		toExecute = append(toExecute, BatchQueryItem{
			Query:       q,
			Vars:        vars,
			resultIndex: i,
		})
	}
	if h.batchQueryExecutor == nil {
		for _, qi := range toExecute {
			result := h.queryExecutor(qi.Query, req, qi.Vars, w.Header())
			results[qi.resultIndex] = result
		}
	} else {
		batchResults := h.batchQueryExecutor(toExecute, req, w.Header())
		for i, qi := range toExecute {
			results[qi.resultIndex] = batchResults[i]
		}
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

func (h *GraphQLHandler) writeResult(w http.ResponseWriter, req *http.Request, gqlRequest *graphQLRequest, body []byte) {
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
}

var iterPool = jsoniter.NewIterator(
	jsoniter.Config{
		UseNumber: false,
	}.Froze(),
).Pool()
