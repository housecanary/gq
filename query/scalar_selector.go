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

// A scalarSelector transforms a selected value into it's serialized form
// by asking the ScalarType associated with the field to encode the value
// to a ScalarValue
type scalarSelector struct {
	Type *schema.ScalarType
	defaultSelector
}

func buildScalarSelector(cc *compileContext, typ *schema.ScalarType, selections ast.SelectionSet) (selector, error) {
	return scalarSelector{Type: typ, defaultSelector: cc.newDefaultSelector()}, nil
}

func (s scalarSelector) apply(ctx *exeContext, value interface{}, collector collector) contFunc {
	if value == nil {
		return nil
	}

	if collectable, ok := value.(schema.CollectableScalar); ok {
		collectable.CollectInto(collector)
		return nil
	}

	sv, err := s.Type.Encode(ctx, value)
	if err != nil {
		collector.Error(err, s.row, s.col)
		return nil
	}
	encodeScalarValue(sv, collector)
	return nil
}

func encodeScalarValue(v schema.LiteralValue, c collector) {
	switch sv := v.(type) {
	case schema.LiteralString:
		c.String(string(sv))
	case schema.LiteralNumber:
		c.Float(float64(sv))
	case schema.LiteralBool:
		c.Bool(bool(sv))
	case schema.LiteralObject:
		oc := c.Object(len(sv))
		for k, v := range sv {
			encodeScalarValue(v, oc.Field(k))
		}
	case schema.LiteralArray:
		ac := c.Array(len(sv))
		for _, v := range sv {
			encodeScalarValue(v, ac.Item())
		}
	}
}
