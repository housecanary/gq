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
	"fmt"
	"math"
	"strconv"
	"unsafe"

	jsoniter "github.com/json-iterator/go"

	"github.com/housecanary/gq/schema"
	"github.com/housecanary/nillabletypes"
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

// Scan implements sql.Scanner
func (v *ID) Scan(src interface{}) error {
	if src == nil {
		*v = ID{present: false}
		return nil
	}
	switch t := src.(type) {
	case []byte:
		*v = ID{present: true, v: string(t)}
		return nil
	case string:
		*v = ID{present: true, v: t}
		return nil
	case int64:
		*v = ID{present: true, v: strconv.FormatInt(t, 64)}
		return nil
	case float64:
		*v = ID{present: true, v: strconv.FormatInt(int64(t), 64)}
		return nil
	}
	return fmt.Errorf("Cannot scan value %v to a string", src)
}

type ns = nillabletypes.String

// String represents the GraphQL built in String type
type String struct {
	ns
}

// NewString makes a new string
func NewString(v string) String {
	return String{nillabletypes.NewString(v)}
}

// NewStringNilEmpty makes a new string, mapping "" to the nil value
func NewStringNilEmpty(v string) String {
	if v == "" {
		return NilString()
	}
	return NewString(v)
}

// NilString makes a nil string
func NilString() String {
	return String{nillabletypes.NilString()}
}

// ToLiteralValue converts this value to a schema.LiteralValue
func (v String) ToLiteralValue() (schema.LiteralValue, error) {
	if v.Nil() {
		return nil, nil
	}
	return schema.LiteralString(v.String()), nil
}

// FromLiteralValue populates this value from a schema.LiteralValue
func (v *String) FromLiteralValue(l schema.LiteralValue) error {
	if l == nil {
		*v = NilString()
		return nil
	}

	if PermissiveInputParsing {
		switch c := l.(type) {
		case schema.LiteralBool:
			*v = NewString(fmt.Sprintf("%v", c))
			PermissiveInputCallback("String", l)
			return nil
		case schema.LiteralNumber:
			*v = NewString(fmt.Sprintf("%v", c))
			PermissiveInputCallback("String", l)
			return nil
		}
	}
	switch c := l.(type) {
	case schema.LiteralString:
		s := string(c)
		*v = NewString(s)
		return nil
	default:
		return fmt.Errorf("Literal value %v is not a string", l)
	}
}

// CollectInto implements schema.CollectableScalar
func (v String) CollectInto(col schema.ScalarCollector) {
	if !v.Nil() {
		col.String(v.String())
	}
}

type ni = nillabletypes.Int32

// Int represents the GraphQL built in Int type
type Int struct {
	ni
}

// NewInt makes a new int
func NewInt(v int32) Int {
	return Int{nillabletypes.NewInt32(v)}
}

// NilInt makes a nil int
func NilInt() Int {
	return Int{nillabletypes.NilInt32()}
}

// ToLiteralValue converts this value to a schema.LiteralValue
func (v Int) ToLiteralValue() (schema.LiteralValue, error) {
	if v.Nil() {
		return nil, nil
	}
	return schema.LiteralNumber(v.Int32()), nil
}

// FromLiteralValue populates this value from a schema.LiteralValue
func (v *Int) FromLiteralValue(l schema.LiteralValue) error {
	if l == nil {
		*v = NilInt()
		return nil
	}

	if PermissiveInputParsing {
		switch c := l.(type) {
		case schema.LiteralString:
			i, err := strconv.ParseInt(string(c), 10, 32)
			if err != nil {
				return err
			}
			*v = NewInt(int32(i))
			PermissiveInputCallback("Int", l)
			return nil
		case schema.LiteralBool:
			i := 0
			if c {
				i = 1
			}
			*v = NewInt(int32(i))
			PermissiveInputCallback("Int", l)
			return nil
		}
	}
	switch c := l.(type) {
	case schema.LiteralNumber:
		if c <= math.MaxInt32 && c >= math.MinInt32 && float64(c) == math.Trunc(float64(c)) {
			i := int32(c)
			*v = NewInt(i)
			return nil
		}
		return fmt.Errorf("Cannot convert float to int (would truncate)")
	default:
		return fmt.Errorf("Literal value %v is not an int", l)
	}
}

// CollectInto implements schema.CollectableScalar
func (v Int) CollectInto(col schema.ScalarCollector) {
	if !v.Nil() {
		col.Int(int64(v.Int32()))
	}
}

type nf = nillabletypes.Float

// Float represents the GraphQL built in Float type
type Float struct {
	nf
}

// NewFloat makes a new float
func NewFloat(v float64) Float {
	return Float{nillabletypes.NewFloat(v)}
}

// NilFloat makes a nil float
func NilFloat() Float {
	return Float{nillabletypes.NilFloat()}
}

// Float64 returns the value of this scalar as a float64. If v is nil, 0 is returned.
func (v Float) Float64() float64 {
	return v.Float()
}

// ToLiteralValue converts this value to a schema.LiteralValue
func (v Float) ToLiteralValue() (schema.LiteralValue, error) {
	if v.Nil() {
		return nil, nil
	}
	return schema.LiteralNumber(v.Float()), nil
}

// FromLiteralValue populates this value from a schema.LiteralValue
func (v *Float) FromLiteralValue(l schema.LiteralValue) error {
	if l == nil {
		*v = NilFloat()
		return nil
	}

	if PermissiveInputParsing {
		switch c := l.(type) {
		case schema.LiteralString:
			f, err := strconv.ParseFloat(string(c), 64)
			if err != nil {
				return err
			}
			*v = NewFloat(f)
			PermissiveInputCallback("Float", l)
			return nil
		case schema.LiteralBool:
			f := float64(0)
			if c {
				f = 1
			}
			*v = NewFloat(f)
			PermissiveInputCallback("Float", l)
			return nil
		}
	}

	switch c := l.(type) {
	case schema.LiteralNumber:
		f := float64(c)
		*v = NewFloat(f)
		return nil
	default:
		return fmt.Errorf("Literal value %v is not a float", l)
	}
}

// CollectInto implements schema.CollectableScalar
func (v Float) CollectInto(col schema.ScalarCollector) {
	if !v.Nil() {
		col.Float(v.Float())
	}
}

type nb = nillabletypes.Bool

// Boolean represents the GraphQL built in Boolean type
type Boolean struct {
	nb
}

// NewBoolean makes a new bool
func NewBoolean(v bool) Boolean {
	return Boolean{nillabletypes.NewBool(v)}
}

// NilBoolean makes a nil bool
func NilBoolean() Boolean {
	return Boolean{nillabletypes.NilBool()}
}

// ToLiteralValue converts this value to a schema.LiteralValue
func (v Boolean) ToLiteralValue() (schema.LiteralValue, error) {
	if v.Nil() {
		return nil, nil
	}
	return schema.LiteralBool(v.Bool()), nil
}

// FromLiteralValue populates this value from a schema.LiteralValue
func (v *Boolean) FromLiteralValue(l schema.LiteralValue) error {
	if l == nil {
		*v = NilBoolean()
		return nil
	}

	if PermissiveInputParsing {
		switch c := l.(type) {
		case schema.LiteralString:
			s := string(c)
			if s == "true" {
				*v = NewBoolean(true)
				PermissiveInputCallback("Boolean", l)
				return nil
			} else if s == "false" {
				*v = NewBoolean(false)
				PermissiveInputCallback("Boolean", l)
				return nil
			}
		case schema.LiteralNumber:
			if c == 0 {
				*v = NewBoolean(false)
			} else {
				*v = NewBoolean(true)
			}
			PermissiveInputCallback("Boolean", l)
			return nil
		}
	}

	switch c := l.(type) {
	case schema.LiteralBool:
		b := bool(c)
		*v = NewBoolean(b)
		return nil
	default:
		return fmt.Errorf("Literal value %v is not a bool", l)
	}
}

// CollectInto implements schema.CollectableScalar
func (v Boolean) CollectInto(col schema.ScalarCollector) {
	if !v.Nil() {
		col.Bool(v.Bool())
	}
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
			*((*String)(ptr)) = NilString()
		case jsoniter.StringValue:
			*((*String)(ptr)) = NewString(iter.ReadString())
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
			*((*Int)(ptr)) = NilInt()
		case jsoniter.NumberValue:
			t := iter.ReadFloat64()

			if math.Trunc(t) != float64(int32(t)) {
				iter.Error = fmt.Errorf("Value %v outside of the range of this type", t)
			}
			*((*Int)(ptr)) = NewInt(int32(t))
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
			*((*Float)(ptr)) = NilFloat()
		case jsoniter.NumberValue:
			t := iter.ReadFloat64()
			*((*Float)(ptr)) = NewFloat(t)
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
			*((*Boolean)(ptr)) = NilBoolean()
		case jsoniter.BoolValue:
			*((*Boolean)(ptr)) = NewBoolean(iter.ReadBool())
		case jsoniter.NumberValue:
			t := iter.ReadInt8()
			*((*Boolean)(ptr)) = NewBoolean(!(t == 0))
		default:
			skipped := iter.SkipAndReturnBytes()
			iter.Error = fmt.Errorf("Value is not a bool (type: %v, data: %v)", v, skipped)
		}
	})
}
