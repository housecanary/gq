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

// An unionSelector maintains a list of ObjectType -> selector pairs.  Once
// a value is known, the union selector finds the correct delegate selector
// based on the ObjectType that the value conforms to, and uses that selector to
// process the value.
type unionSelector struct {
	Elements map[string]selector
	Type     *schema.UnionType
	defaultSelector
}

func buildUnionSelector(cc *compileContext, typ *schema.UnionType, selections ast.SelectionSet) (selector, error) {
	for _, sel := range selections {
		switch v := sel.(type) {
		case *ast.FieldSelection:
			if v.Field.Name != "__typename" {
				return nil, fmt.Errorf("Cannot select field %s on union", v.Field.Name)
			}
		}
	}

	elements := make(map[string]selector)
	for _, objectType := range typ.Members() {
		objectSel, err := buildObjectSelector(cc, objectType, selections)
		if err != nil {
			return nil, err
		}
		elements[objectType.Name()] = objectSel
	}
	return unionSelector{Elements: elements, Type: typ, defaultSelector: cc.newDefaultSelector()}, nil
}

func (s unionSelector) apply(ctx exeContext, value interface{}, collector collector) contFunc {
	if value == nil {
		return nil
	}

	oval, otyp := s.Type.Unwrap(ctx, value)

	if oval == nil {
		return nil
	}

	delegate, ok := s.Elements[otyp]
	if !ok {
		err := fmt.Errorf("Value %v does not conform to any member of union %s", value, s.Type.Name())
		ctx.listener.NotifyError(err)
		collector.Error(err, s.row, s.col)
		return nil
	}

	return delegate.apply(ctx, oval, collector)
}
