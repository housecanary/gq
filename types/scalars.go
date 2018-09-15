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

package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/housecanary/gq/schema"
)

// PermissiveInputParsing allows more input value conversions than allowed by the spec if set to true
var PermissiveInputParsing = false

// PermissiveInputCallback is a callback that is notified when a value is accepted via permissive parsing
var PermissiveInputCallback = func(typ string, value schema.LiteralValue) {}

// ID represents the GraphQL built in ID type
type ID struct {
	v       string
	present bool
}

// NewID creates a new ID
func NewID(v string) ID {
	return ID{v, true}
}

// NilID creates a nil ID
func NilID() ID {
	return ID{"", false}
}

// String returns the raw value of this scalar as a string
func (v ID) String() string {
	return v.v
}

// Nil returns whether this scalar is nil
func (v ID) Nil() bool {
	return !v.present
}

// ToLiteralValue converts this value to a schema.LiteralValue
func (v ID) ToLiteralValue() (schema.LiteralValue, error) {
	if !v.present {
		return nil, nil
	}
	return schema.LiteralString(v.v), nil
}

// FromLiteralValue populates this value from a schema.LiteralValue
func (v *ID) FromLiteralValue(l schema.LiteralValue) error {
	if l == nil {
		*v = ID{"", false}
		return nil
	}

	switch c := l.(type) {
	case schema.LiteralString:
		s := string(c)
		*v = ID{s, true}
		return nil
	case schema.LiteralNumber:
		s := strconv.Itoa(int(c))
		*v = ID{s, true}
		return nil
	default:
		return fmt.Errorf("Literal value %v is not a string or number", l)
	}
}

// CollectInto implements schema.CollectableScalar
func (v ID) CollectInto(col schema.ScalarCollector) {
	if v.present {
		col.String(v.v)
	}
}

// String represents the GraphQL built in String type
type String struct {
	v       string
	present bool
}

// NewString makes a new string
func NewString(v string) String {
	return String{v, true}
}

// NewStringNilEmpty makes a new string, mapping "" to the nil value
func NewStringNilEmpty(v string) String {
	if v == "" {
		return String{"", false}
	}
	return String{v, true}
}

// NilString makes a nil string
func NilString() String {
	return String{"", false}
}

// String returns the string value.  If v is nil, "" is returned.
func (v String) String() string {
	return v.v
}

// Nil returns whether this scalar is nil
func (v String) Nil() bool {
	return !v.present
}

// ToLiteralValue converts this value to a schema.LiteralValue
func (v String) ToLiteralValue() (schema.LiteralValue, error) {
	if !v.present {
		return nil, nil
	}
	return schema.LiteralString(v.v), nil
}

// FromLiteralValue populates this value from a schema.LiteralValue
func (v *String) FromLiteralValue(l schema.LiteralValue) error {
	if l == nil {
		*v = String{"", false}
		return nil
	}

	if PermissiveInputParsing {
		switch c := l.(type) {
		case schema.LiteralBool:
			*v = String{fmt.Sprintf("%v", c), true}
			PermissiveInputCallback("String", l)
			return nil
		case schema.LiteralNumber:
			*v = String{fmt.Sprintf("%v", c), true}
			PermissiveInputCallback("String", l)
			return nil
		}
	}
	switch c := l.(type) {
	case schema.LiteralString:
		s := string(c)
		*v = String{s, true}
		return nil
	default:
		return fmt.Errorf("Literal value %v is not a string", l)
	}
}

// CollectInto implements schema.CollectableScalar
func (v String) CollectInto(col schema.ScalarCollector) {
	if v.present {
		col.String(v.v)
	}
}

// UnmarshalJSON implements json.Unmarshaller
func (v *String) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte{'n', 'u', 'l', 'l'}) {
		v.present = false
		return nil
	}
	err := json.Unmarshal(data, &v.v)
	if err != nil {
		return err
	}
	v.present = true
	return nil
}

// Scan implements sql.Scanner
func (v *String) Scan(src interface{}) error {
	if src == nil {
		*v = String{present: false}
		return nil
	}
	switch t := src.(type) {
	case []byte:
		*v = String{present: true, v: string(t)}
		return nil
	case string:
		*v = String{present: true, v: t}
		return nil
	}
	return fmt.Errorf("Cannot scan value %v to a string", src)
}

// Int represents the GraphQL built in Int type
type Int struct {
	v       int32
	present bool
}

// NewInt makes a new int
func NewInt(v int32) Int {
	return Int{v, true}
}

// NilInt makes a nil int
func NilInt() Int {
	return Int{0, false}
}

// Int32 returns the value of this scalar as an int32. If v is nil, 0 is returned.
func (v Int) Int32() int32 {
	return v.v
}

// Nil returns whether this scalar is nil
func (v Int) Nil() bool {
	return !v.present
}

// ToLiteralValue converts this value to a schema.LiteralValue
func (v Int) ToLiteralValue() (schema.LiteralValue, error) {
	if !v.present {
		return nil, nil
	}
	return schema.LiteralNumber(v.v), nil
}

// FromLiteralValue populates this value from a schema.LiteralValue
func (v *Int) FromLiteralValue(l schema.LiteralValue) error {
	if l == nil {
		*v = Int{0, false}
		return nil
	}

	if PermissiveInputParsing {
		switch c := l.(type) {
		case schema.LiteralString:
			i, err := strconv.ParseInt(string(c), 10, 32)
			if err != nil {
				return err
			}
			*v = Int{int32(i), true}
			PermissiveInputCallback("Int", l)
			return nil
		case schema.LiteralBool:
			i := 0
			if c {
				i = 1
			}
			*v = Int{int32(i), true}
			PermissiveInputCallback("Int", l)
			return nil
		}
	}
	switch c := l.(type) {
	case schema.LiteralNumber:
		if c <= math.MaxInt32 && c >= math.MinInt32 && float64(c) == math.Trunc(float64(c)) {
			i := int32(c)
			*v = Int{i, true}
			return nil
		}
		return fmt.Errorf("Cannot convert float to int (would truncate)")
	default:
		return fmt.Errorf("Literal value %v is not an int", l)
	}
}

// CollectInto implements schema.CollectableScalar
func (v Int) CollectInto(col schema.ScalarCollector) {
	if v.present {
		col.Int(int64(v.v))
	}
}

// UnmarshalJSON implements json.Unmarshaller
func (v *Int) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte{'n', 'u', 'l', 'l'}) {
		v.present = false
		return nil
	}
	var t float64
	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	if math.Trunc(t) != float64(int32(t)) {
		return fmt.Errorf("Value %v outside of the range of this type", t)
	}
	v.v = int32(t)
	v.present = true
	return nil
}

// Scan implements sql.Scanner
func (v *Int) Scan(src interface{}) error {
	if src == nil {
		*v = Int{present: false}
		return nil
	}
	switch t := src.(type) {
	case int64:
		*v = Int{present: true, v: int32(t)}
		return nil
	case float64:
		*v = Int{present: true, v: int32(t)}
		return nil
	case bool:
		if t {
			*v = Int{present: true, v: 1}
		} else {
			*v = Int{present: true, v: 0}
		}
	}
	return fmt.Errorf("Cannot scan value %v to an int", src)
}

// Float represents the GraphQL built in Float type
type Float struct {
	v       float64
	present bool
}

// NewFloat makes a new float
func NewFloat(v float64) Float {
	return Float{v, true}
}

// NilFloat makes a nil float
func NilFloat() Float {
	return Float{0, false}
}

// Float64 returns the value of this scalar as a float64. If v is nil, 0 is returned.
func (v Float) Float64() float64 {
	return v.v
}

// Nil returns whether this scalar is nil
func (v Float) Nil() bool {
	return !v.present
}

// ToLiteralValue converts this value to a schema.LiteralValue
func (v Float) ToLiteralValue() (schema.LiteralValue, error) {
	if !v.present {
		return nil, nil
	}
	return schema.LiteralNumber(v.v), nil
}

// FromLiteralValue populates this value from a schema.LiteralValue
func (v *Float) FromLiteralValue(l schema.LiteralValue) error {
	if l == nil {
		*v = Float{0, false}
		return nil
	}

	if PermissiveInputParsing {
		switch c := l.(type) {
		case schema.LiteralString:
			f, err := strconv.ParseFloat(string(c), 64)
			if err != nil {
				return err
			}
			*v = Float{f, true}
			PermissiveInputCallback("Float", l)
			return nil
		case schema.LiteralBool:
			f := float64(0)
			if c {
				f = 1
			}
			*v = Float{f, true}
			PermissiveInputCallback("Float", l)
			return nil
		}
	}

	switch c := l.(type) {
	case schema.LiteralNumber:
		f := float64(c)
		*v = Float{f, true}
		return nil
	default:
		return fmt.Errorf("Literal value %v is not a float", l)
	}
}

// CollectInto implements schema.CollectableScalar
func (v Float) CollectInto(col schema.ScalarCollector) {
	if v.present {
		col.Float(v.v)
	}
}

// UnmarshalJSON implements json.Unmarshaller
func (v *Float) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte{'n', 'u', 'l', 'l'}) {
		v.present = false
		return nil
	}
	err := json.Unmarshal(data, &v.v)
	if err != nil {
		return err
	}
	v.present = true
	return nil
}

// Scan implements sql.Scanner
func (v *Float) Scan(src interface{}) error {
	if src == nil {
		*v = Float{present: false}
		return nil
	}
	switch t := src.(type) {
	case float64:
		*v = Float{present: true, v: t}
		return nil
	case int64:
		*v = Float{present: true, v: float64(t)}
	case bool:
		if t {
			*v = Float{present: true, v: 1}
		} else {
			*v = Float{present: true, v: 0}
		}
		return nil
	}
	return fmt.Errorf("Cannot scan value %v to a float", src)
}

// Boolean represents the GraphQL built in Boolean type
type Boolean struct {
	v       bool
	present bool
}

// NewBoolean makes a new bool
func NewBoolean(v bool) Boolean {
	return Boolean{v, true}
}

// NilBoolean makes a nil bool
func NilBoolean() Boolean {
	return Boolean{false, false}
}

// Bool returns the value of this scalar as a bool. If v is nil, false is returned.
func (v Boolean) Bool() bool {
	return v.v
}

// Nil returns whether this scalar is nil
func (v Boolean) Nil() bool {
	return !v.present
}

// ToLiteralValue converts this value to a schema.LiteralValue
func (v Boolean) ToLiteralValue() (schema.LiteralValue, error) {
	if !v.present {
		return nil, nil
	}
	return schema.LiteralBool(v.v), nil
}

// FromLiteralValue populates this value from a schema.LiteralValue
func (v *Boolean) FromLiteralValue(l schema.LiteralValue) error {
	if l == nil {
		*v = Boolean{false, false}
		return nil
	}

	if PermissiveInputParsing {
		switch c := l.(type) {
		case schema.LiteralString:
			s := string(c)
			if s == "true" {
				*v = Boolean{true, true}
				PermissiveInputCallback("Boolean", l)
				return nil
			} else if s == "false" {
				*v = Boolean{false, true}
				PermissiveInputCallback("Boolean", l)
				return nil
			}
		case schema.LiteralNumber:
			if c == 0 {
				*v = Boolean{false, true}
			} else {
				*v = Boolean{true, true}
			}
			PermissiveInputCallback("Boolean", l)
			return nil
		}
	}

	switch c := l.(type) {
	case schema.LiteralBool:
		b := bool(c)
		*v = Boolean{b, true}
		return nil
	default:
		return fmt.Errorf("Literal value %v is not a bool", l)
	}
}

// CollectInto implements schema.CollectableScalar
func (v Boolean) CollectInto(col schema.ScalarCollector) {
	if v.present {
		col.Bool(v.v)
	}
}

// UnmarshalJSON implements json.Unmarshaller
func (v *Boolean) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte{'n', 'u', 'l', 'l'}) {
		v.present = false
		return nil
	}
	err := json.Unmarshal(data, &v.v)
	if err != nil {
		return err
	}
	v.present = true
	return nil
}

// Scan implements sql.Scanner
func (v *Boolean) Scan(src interface{}) error {
	if src == nil {
		*v = Boolean{present: false}
		return nil
	}
	switch t := src.(type) {
	case bool:
		*v = Boolean{present: true, v: t}
		return nil
	case int64:
		if t == 0 {
			*v = Boolean{present: true, v: false}
		} else {
			*v = Boolean{present: true, v: true}
		}
		return nil
	case float64:
		if t == 0 {
			*v = Boolean{present: true, v: false}
		} else {
			*v = Boolean{present: true, v: true}
		}
		return nil
	}
	return fmt.Errorf("Cannot scan value %v to a bool", src)
}

func init() {
	jsoniter.RegisterTypeDecoderFunc("types.ID", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		v := iter.WhatIsNext()
		switch v {
		case jsoniter.NilValue:
			iter.ReadNil()
			*((*ID)(ptr)) = ID{present: false}
		case jsoniter.StringValue:
			*((*ID)(ptr)) = ID{v: iter.ReadString(), present: true}
		case jsoniter.NumberValue:
			*((*ID)(ptr)) = ID{v: string(iter.ReadNumber()), present: true}
		default:
			skipped := iter.SkipAndReturnBytes()
			iter.Error = fmt.Errorf("Value is not a string or number (type: %v, data: %v)", v, skipped)
		}
	})

	jsoniter.RegisterTypeDecoderFunc("types.String", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		v := iter.WhatIsNext()
		switch v {
		case jsoniter.NilValue:
			iter.ReadNil()
			*((*String)(ptr)) = String{present: false}
		case jsoniter.StringValue:
			*((*String)(ptr)) = String{v: iter.ReadString(), present: true}
		default:
			skipped := iter.SkipAndReturnBytes()
			iter.Error = fmt.Errorf("Value is not a string (type: %v, data: %v)", v, skipped)
		}
	})

	jsoniter.RegisterTypeDecoderFunc("types.Int", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		v := iter.WhatIsNext()
		switch v {
		case jsoniter.NilValue:
			iter.ReadNil()
			*((*Int)(ptr)) = Int{present: false}
		case jsoniter.NumberValue:
			t := iter.ReadFloat64()

			if math.Trunc(t) != float64(int32(t)) {
				iter.Error = fmt.Errorf("Value %v outside of the range of this type", t)
			}
			*((*Int)(ptr)) = Int{v: int32(t), present: true}
		default:
			skipped := iter.SkipAndReturnBytes()
			iter.Error = fmt.Errorf("Value is not a number (type: %v, data: %v)", v, skipped)
		}
	})

	jsoniter.RegisterTypeDecoderFunc("types.Float", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		v := iter.WhatIsNext()
		switch v {
		case jsoniter.NilValue:
			iter.ReadNil()
			*((*Float)(ptr)) = Float{present: false}
		case jsoniter.NumberValue:
			t := iter.ReadFloat64()
			*((*Float)(ptr)) = Float{v: t, present: true}
		default:
			skipped := iter.SkipAndReturnBytes()
			iter.Error = fmt.Errorf("Value is not a number (type: %v, data: %v)", v, skipped)
		}
	})

	jsoniter.RegisterTypeDecoderFunc("types.Boolean", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		v := iter.WhatIsNext()
		switch v {
		case jsoniter.NilValue:
			iter.ReadNil()
			*((*Boolean)(ptr)) = Boolean{present: false}
		case jsoniter.BoolValue:
			*((*Boolean)(ptr)) = Boolean{v: iter.ReadBool(), present: true}
		case jsoniter.NumberValue:
			t := iter.ReadInt8()
			*((*Boolean)(ptr)) = Boolean{v: !(t == 0), present: true}
		default:
			skipped := iter.SkipAndReturnBytes()
			iter.Error = fmt.Errorf("Value is not a bool (type: %v, data: %v)", v, skipped)
		}
	})
}
