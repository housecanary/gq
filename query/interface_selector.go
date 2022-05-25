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

// An interfaceSelector maintains a list of ObjectType -> selector pairs.  Once
// a value is known, the interface selector finds the correct delegate selector
// based on the ObjectType that the value conforms to, and uses that selector to
// process the value.
type interfaceSelector struct {
	Elements map[string]selector
	Type     *schema.InterfaceType
	defaultSelector
}

func buildInterfaceSelector(cc *compileContext, typ *schema.InterfaceType, selections ast.SelectionSet) (selector, error) {
	for _, sel := range selections {
		switch v := sel.(type) {
		case *ast.FieldSelection:
			if !typ.HasField(v.Field.Name) {
				return nil, fmt.Errorf("Field %s does not exist on interface %s", v.Field.Name, typ.Name())
			}
		}
	}

	elements := make(map[string]selector)
	for _, objectType := range typ.Implementations() {
		objectSel, err := buildObjectSelector(cc, objectType, selections)
		if err != nil {
			return nil, err
		}
		elements[objectType.Name()] = objectSel
	}
	return interfaceSelector{Elements: elements, Type: typ, defaultSelector: cc.newDefaultSelector()}, nil
}

func (s interfaceSelector) apply(ctx *exeContext, value interface{}, collector collector) contFunc {
	if value == nil {
		return nil
	}

	oval, otyp := s.Type.Unwrap(ctx, value)

	if oval == nil {
		return nil
	}

	delegate, ok := s.Elements[otyp]
	if !ok {
		err := fmt.Errorf("Value %v does not conform to interface %s", value, s.Type.Name())
		ctx.listener.NotifyError(err)
		collector.Error(err, s.row, s.col)
		return nil
	}

	return delegate.apply(ctx, oval, collector)
}
