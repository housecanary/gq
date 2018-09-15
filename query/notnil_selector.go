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
	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

// A notNilSelector wraps a delegate selector and enforces that the returned
// value is not nil.  Most of the work to do so is in the collector, the notNilSelector
// simply marks the collector as required in its prepareCollector method.
type notNilSelector struct {
	Delegate selector
	defaultSelector
}

func buildNotNilSelector(cc *compileContext, typ *schema.NotNilType, selections ast.SelectionSet) (selector, error) {
	s, err := buildSelector(cc, typ.Unwrap(), selections)
	if err != nil {
		return nil, err
	}
	return notNilSelector{Delegate: s, defaultSelector: cc.newDefaultSelector()}, nil
}

func (s notNilSelector) prepareCollector(c collector) {
	c.Required(s.row, s.col)
}

func (s notNilSelector) apply(ctx exeContext, value interface{}, collector collector) contFunc {
	return s.Delegate.apply(ctx, value, collector)
}
