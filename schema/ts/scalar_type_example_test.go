// Copyright 2023 HouseCanary, Inc.
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

package ts_test

import (
	"fmt"
	"time"

	"github.com/housecanary/gq/schema"
	"github.com/housecanary/gq/schema/ts"
)

var scalarModType = ts.NewModule()

// To create an scalar type, first define the type that will hold your value.
// The type must implement schema.ScalarMarshaler for the value and schema.ScalarUnmarshaler
// for a pointer to the value

type DateTime time.Time

func (v DateTime) ToLiteralValue() (schema.LiteralValue, error) {
	formatted := time.Time(v).Format(time.RFC3339Nano)
	return schema.LiteralString(formatted), nil
}

func (v *DateTime) FromLiteralValue(l schema.LiteralValue) error {
	if l == nil {
		*v = DateTime{}
		return nil
	}
	switch c := l.(type) {
	case schema.LiteralString:
		parsed, err := time.Parse(time.RFC3339Nano, string(c))
		if err != nil {
			return fmt.Errorf("invalid datetime %s: %w", c, err)
		}
		*v = DateTime(parsed)
		return nil
	default:
		return fmt.Errorf("Literal value %v is not a string", l)
	}
}

// Next, construct the GQL type using the ts.NewScalarType function
var dateType = ts.NewScalarType[DateTime](scalarModType, `"An ISO format datetime."`)

func ExampleNewScalarType() {
	// Once the scalar type is registered, it can be used in arguments, as a struct field, etc
	// just like any of the built in types
}
