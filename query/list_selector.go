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
	"fmt"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

// A listSelector has a delegate selector that it uses to process each element of a
// list of values.
type listSelector struct {
	ElementSelector selector
	defaultSelector
}

func buildListSelector(cc *compileContext, typ *schema.ListType, selections ast.SelectionSet) (selector, error) {
	elementSelector, err := buildSelector(cc, typ.Unwrap(), selections)
	if err != nil {
		return nil, err
	}
	return listSelector{ElementSelector: elementSelector, defaultSelector: cc.newDefaultSelector()}, nil
}

func (s listSelector) apply(ctx exeContext, value interface{}, collector collector) contFunc {
	if value == nil {
		return nil
	}

	var deferred worklist
	lv, ok := value.(schema.ListValue)
	if !ok {
		err := fmt.Errorf("Value %v is not a list value", value)
		ctx.listener.NotifyError(err)
		collector.Error(err, s.row, s.col)
		return nil
	}

	valueCollector := collector.Array(lv.Len())
	lv.ForEachElement(func(item interface{}) {
		elementCollector := valueCollector.Item()
		s.ElementSelector.prepareCollector(elementCollector)
		deferred.Add(s.ElementSelector.apply(ctx, item, elementCollector))
	})

	if len(deferred) > 0 {
		return deferred.Continue
	}

	return nil
}
