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

// An enumSelector transforms a selected value into it's serialized form
// by asking the EnumType associated with the field to encode the value
// to a LiteralString
type enumSelector struct {
	Type *schema.EnumType
	defaultSelector
}

func buildEnumSelector(cc *compileContext, typ *schema.EnumType, selections ast.SelectionSet) (selector, error) {
	return enumSelector{Type: typ, defaultSelector: cc.newDefaultSelector()}, nil
}

func (s enumSelector) apply(ctx exeContext, value interface{}, collector collector) contFunc {
	lv, err := s.Type.Encode(ctx, value)
	if err != nil {
		ctx.listener.NotifyError(err)
		collector.Error(err, s.row, s.col)
		return nil
	}

	if lv == nil {
		return nil
	}

	if sv, ok := lv.(schema.LiteralString); ok {
		collector.String(string(sv))
	} else {
		err := fmt.Errorf("Expected a string value for enum, but got %v", lv)
		ctx.listener.NotifyError(err)
		collector.Error(err, s.row, s.col)
	}

	return nil
}
