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

	jsoniter "github.com/json-iterator/go"

	"github.com/housecanary/gq/schema"
)

// Variables represents variables passed to the query
type Variables schema.LiteralObject

// NewVariablesFromJSON creates a Variables struct from JSON text
func NewVariablesFromJSON(j []byte) (Variables, error) {
	if len(j) == 0 {
		return nil, nil
	}
	itr := iterPool.BorrowIterator(j)
	defer func() { iterPool.ReturnIterator(itr) }()

	t := itr.WhatIsNext()

	if t == jsoniter.ObjectValue {
		lo, ok := decodeJSONObject(itr)
		if !ok {
			return nil, itr.Error
		}
		return Variables(lo), nil
	} else if t == jsoniter.NilValue {
		return nil, nil
	} else {
		return nil, fmt.Errorf("Invalid variables input %s", string(j))
	}
}

var iterPool = jsoniter.NewIterator(
	jsoniter.Config{
		UseNumber: false,
	}.Froze(),
).Pool()

func decodeJSONValue(itr *jsoniter.Iterator) (schema.LiteralValue, bool) {
	switch itr.WhatIsNext() {
	case jsoniter.InvalidValue:
		return nil, false
	case jsoniter.StringValue:
		return schema.LiteralString(itr.ReadStringAsSlice()), true
	case jsoniter.NumberValue:
		return schema.LiteralNumber(itr.ReadFloat64()), true
	case jsoniter.NilValue:
		itr.ReadNil()
		return nil, true
	case jsoniter.BoolValue:
		return schema.LiteralBool(itr.ReadBool()), true
	case jsoniter.ArrayValue:
		return decodeJSONArray(itr)
	case jsoniter.ObjectValue:
		return decodeJSONObject(itr)
	}
	itr.ReportError("decodeJSONValue", fmt.Sprintf("Invalid token type %v", itr.WhatIsNext()))
	return nil, false
}

func decodeJSONObject(itr *jsoniter.Iterator) (schema.LiteralObject, bool) {
	r := make(schema.LiteralObject)
	ok := itr.ReadObjectCB(func(itr *jsoniter.Iterator, field string) bool {
		v, ok := decodeJSONValue(itr)
		if !ok {
			return false
		}
		r[field] = v
		return true
	})
	return r, ok
}

func decodeJSONArray(itr *jsoniter.Iterator) (schema.LiteralArray, bool) {
	r := make(schema.LiteralArray, 0)
	ok := itr.ReadArrayCB(func(itr *jsoniter.Iterator) bool {
		v, ok := decodeJSONValue(itr)
		if !ok {
			return false
		}
		r = append(r, v)
		return true
	})
	return r, ok
}
