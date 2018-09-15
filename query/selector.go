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

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

type exeContext struct {
	context.Context
	listener  ExecutionListener
	variables Variables
}

// A contFunc represents remaining work that a selector needs to perform.
//
// Code using a selector is expected to invoke any contFunc returned by a selector
// repeatedly until the contFunc returns nil
type contFunc func() contFunc

// A worklist is used for asynchronous execution.
//
// The worklist maintains a list of continuation function that are used to advance
// execution of a query.  The expected use of the worklist is to add a number of
// continuation functions using Add(), and then call Continue repeatedly until
// Continue() returns nil.  Once Continue returns nil, the worklist is complete,
// and all work has been performed.
type worklist []contFunc

func (w *worklist) Add(c contFunc) {
	if c == nil {
		return
	}
	*w = append(*w, c)
}

func (w *worklist) Continue() contFunc {
	i := 0
	for _, f := range *w {
		c := f()
		if c != nil {
			(*w)[i] = c
			i++
		}
	}
	*w = (*w)[0:i]
	if i == 0 {
		return nil
	}
	return w.Continue
}

// A selector is the runtime peer of an element in the query tree.
//
// Selectors are responsible for extracting the data specified by a query
// Selection Set from an input value, and providing the results to a collector.
type selector interface {
	prepareCollector(collector collector)
	apply(ctx exeContext, value interface{}, collector collector) contFunc
}

type defaultSelector struct {
	row int
	col int
}

func (defaultSelector) prepareCollector(collector collector) {
}

// buildSelector builds a selector for an input type and selection set.  If the query is invalid according
// to the schema it may return an error.
func buildSelector(cc *compileContext, typ schema.Type, selections ast.SelectionSet) (selector, error) {
	switch t := typ.(type) {
	case *schema.ObjectType:
		return buildObjectSelector(cc, t, selections)
	case *schema.NotNilType:
		return buildNotNilSelector(cc, t, selections)
	case *schema.ListType:
		return buildListSelector(cc, t, selections)
	case *schema.InterfaceType:
		return buildInterfaceSelector(cc, t, selections)
	case *schema.UnionType:
		return buildUnionSelector(cc, t, selections)
	case *schema.ScalarType:
		return buildScalarSelector(cc, t, selections)
	case *schema.EnumType:
		return buildEnumSelector(cc, t, selections)
	}

	return nil, fmt.Errorf("Invalid type %v", typ)
}
