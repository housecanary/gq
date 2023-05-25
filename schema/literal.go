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

package schema

import (
	"github.com/housecanary/gq/ast"
)

// LiteralValue is a marker interface for all known serializable types
type LiteralValue interface {
	isLiteralValue()
}

// LiteralString represents a GraphQL String value
type LiteralString string

func (LiteralString) isLiteralValue() {}

// LiteralNumber represents a GraphQL Float or Int value
type LiteralNumber float64

func (LiteralNumber) isLiteralValue() {}

// LiteralBool represents a GraphQL Bool value
type LiteralBool bool

func (LiteralBool) isLiteralValue() {}

// LiteralObject represents an object composed of GraphQL values
type LiteralObject map[string]LiteralValue

func (LiteralObject) isLiteralValue() {}

// LiteralArray represents an array composed of GraphQL values
type LiteralArray []LiteralValue

func (LiteralArray) isLiteralValue() {}

func LiteralValueFromAstValue(v ast.Value) LiteralValue {
	if v == nil {
		return nil
	}
	switch tv := v.(type) {
	case ast.EnumValue:
		return LiteralString(tv.V)
	case ast.StringValue:
		return LiteralString(tv.V)
	case ast.IntValue:
		return LiteralNumber(tv.V)
	case ast.FloatValue:
		return LiteralNumber(tv.V)
	case ast.BooleanValue:
		return LiteralBool(tv.V)
	case ast.ObjectValue:
		lv := make(LiteralObject)
		for k, v := range tv.V {
			lv[k] = LiteralValueFromAstValue(v)
		}
		return lv
	case ast.ArrayValue:
		lv := make(LiteralArray, len(tv.V))
		for i, v := range tv.V {
			lv[i] = LiteralValueFromAstValue(v)
		}
		return lv
	}
	panic("Cannot create literal value from ast value")
}
