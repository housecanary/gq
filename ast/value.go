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

package ast

import (
	"fmt"
	"strconv"
	"strings"
)

type Value interface {
	Representation() string
	isValue()
}

type StringValue struct {
	V string
}

func (StringValue) isValue() {}

func (v StringValue) Representation() string {
	s := v.V
	sb := strings.Builder{}
	sb.Grow(len(s) + 2)
	sb.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < ' ' || c == '\\' || c == '"' {
			switch c {
			case '\\', '"':
				sb.WriteByte('\\')
				sb.WriteByte(c)
			case '\b':
				sb.WriteByte('\\')
				sb.WriteByte('b')
			case '\f':
				sb.WriteByte('\\')
				sb.WriteByte('f')
			case '\n':
				sb.WriteByte('\\')
				sb.WriteByte('n')
			case '\r':
				sb.WriteByte('\\')
				sb.WriteByte('r')
			case '\t':
				sb.WriteByte('\t')
			default:
				sb.WriteByte('\\')
				sb.WriteByte('u')
				formatted := strconv.FormatUint(uint64(c), 16)
				for i := 0; i < 4-len(formatted); i++ {
					sb.WriteByte('0')
				}
				for i := 0; i < len(formatted); i++ {
					sb.WriteByte(formatted[i])
				}
			}
		} else {
			sb.WriteByte(c)
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

var _ Value = (*StringValue)(nil)

type IntValue struct {
	V int64
}

func (IntValue) isValue() {}

func (v IntValue) Representation() string {
	return fmt.Sprintf("%v", v.V)
}

var _ Value = (*IntValue)(nil)

type FloatValue struct {
	V float64
}

func (FloatValue) isValue() {}

func (v FloatValue) Representation() string {
	return fmt.Sprintf("%v", v.V)
}

var _ Value = (*FloatValue)(nil)

type BooleanValue struct {
	V bool
}

func (BooleanValue) isValue() {}

func (v BooleanValue) Representation() string {
	if v.V {
		return "true"
	}
	return "false"
}

var _ Value = (*BooleanValue)(nil)

type NilValue struct {
}

func (NilValue) isValue() {}

func (NilValue) Representation() string {
	return "null"
}

var _ Value = (*NilValue)(nil)

type EnumValue struct {
	V string
}

func (EnumValue) isValue() {}

func (v EnumValue) Representation() string {
	return v.V
}

var _ Value = (*EnumValue)(nil)

type ArrayValue struct {
	V []Value
}

func (ArrayValue) isValue() {}

func (v ArrayValue) Representation() string {
	var s strings.Builder
	s.WriteString("[")
	for i, e := range v.V {
		if i > 0 {
			s.WriteString(", ")
		}
		s.WriteString(e.Representation())
	}
	s.WriteString("]")
	return s.String()
}

var _ Value = (*ArrayValue)(nil)

type ObjectValue struct {
	V map[string]Value
}

func (ObjectValue) isValue() {}
func (v ObjectValue) Representation() string {
	var s strings.Builder
	s.WriteString("{")
	var i = 0
	for k, v := range v.V {
		if i > 0 {
			s.WriteString(", ")
		}
		i++
		s.WriteString(k)
		s.WriteString(": ")
		s.WriteString(v.Representation())
	}
	s.WriteString("}")
	return s.String()
}

var _ Value = (*ObjectValue)(nil)

type ReferenceValue struct {
	Name string
}

func (v ReferenceValue) isValue() {}

func (v ReferenceValue) Representation() string {
	return fmt.Sprintf("$%s", v.Name)
}

var _ Value = (*ReferenceValue)(nil)
