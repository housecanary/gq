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
	"math"
	"reflect"
	"testing"

	"github.com/housecanary/gq/schema"
	"github.com/housecanary/nillabletypes"
)

func TestNewStringNilEmpty(t *testing.T) {
	tests := []struct {
		name string
		give string
		want String
	}{
		{
			name: "Empty",
			give: "",
			want: NilString(),
		},
		{
			name: "Not Empty",
			give: "a",
			want: NewString("a"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStringNilEmpty(tt.give); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStringNilEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_ToLiteralValue(t *testing.T) {
	tests := []struct {
		name    string
		give    nillabletypes.String
		want    schema.LiteralValue
		wantErr bool
	}{
		{
			name:    "Nil",
			give:    nillabletypes.NilString(),
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Not Nil",
			give:    nillabletypes.NewString("abcd"),
			want:    schema.LiteralString("abcd"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := String{
				ns: tt.give,
			}
			got, err := v.ToLiteralValue()
			if (err != nil) != tt.wantErr {
				t.Errorf("String.ToLiteralValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String.ToLiteralValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt_ToLiteralValue(t *testing.T) {
	tests := []struct {
		name    string
		give    nillabletypes.Int32
		want    schema.LiteralValue
		wantErr bool
	}{
		{
			name:    "Nil",
			give:    nillabletypes.NilInt32(),
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Not Nil",
			give:    nillabletypes.NewInt32(17),
			want:    schema.LiteralNumber(17),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Int{
				ni: tt.give,
			}
			got, err := v.ToLiteralValue()
			if (err != nil) != tt.wantErr {
				t.Errorf("Int.ToLiteralValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Int.ToLiteralValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat_ToLiteralValue(t *testing.T) {
	tests := []struct {
		name    string
		give    nillabletypes.Float
		want    schema.LiteralValue
		wantErr bool
	}{
		{
			name:    "Nil",
			give:    nillabletypes.NilFloat(),
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Not Nil",
			give:    nillabletypes.NewFloat(20324.5),
			want:    schema.LiteralNumber(20324.5),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Float{
				nf: tt.give,
			}
			got, err := v.ToLiteralValue()
			if (err != nil) != tt.wantErr {
				t.Errorf("Float.ToLiteralValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Float.ToLiteralValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoolean_ToLiteralValue(t *testing.T) {
	tests := []struct {
		name    string
		give    nillabletypes.Bool
		want    schema.LiteralValue
		wantErr bool
	}{
		{
			name:    "Nil",
			give:    nillabletypes.NilBool(),
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Not Nil",
			give:    nillabletypes.NewBool(true),
			want:    schema.LiteralBool(true),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Boolean{
				nb: tt.give,
			}
			got, err := v.ToLiteralValue()
			if (err != nil) != tt.wantErr {
				t.Errorf("Boolean.ToLiteralValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Boolean.ToLiteralValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_FromLiteralValue(t *testing.T) {
	tests := []struct {
		name                   string
		give                   schema.LiteralValue
		permissiveInputParsing bool
		want                   String
		wantErr                bool
	}{
		{
			name:                   "From Nil",
			give:                   nil,
			permissiveInputParsing: false,
			want:                   NilString(),
			wantErr:                false,
		},
		{
			name:                   "From String",
			give:                   schema.LiteralString("abcd"),
			permissiveInputParsing: false,
			want:                   NewString("abcd"),
			wantErr:                false,
		},
		{
			name:                   "From Bool w/out Permissive Input Parsing",
			give:                   schema.LiteralBool(true),
			permissiveInputParsing: false,
			want:                   String{},
			wantErr:                true,
		},
		{
			name:                   "From Number w/out Permissive Input Parsing",
			give:                   schema.LiteralNumber(1234),
			permissiveInputParsing: false,
			want:                   String{},
			wantErr:                true,
		},
		{
			name:                   "From Bool w/ Permissive Input Parsing",
			give:                   schema.LiteralBool(true),
			permissiveInputParsing: true,
			want:                   NewString("true"),
			wantErr:                false,
		},
		{
			name:                   "From Number w/ Permissive Input Parsing",
			give:                   schema.LiteralNumber(1234),
			permissiveInputParsing: true,
			want:                   NewString("1234"),
			wantErr:                false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origPermissiveInputParsing := PermissiveInputParsing
			PermissiveInputParsing = tt.permissiveInputParsing
			got := &String{}
			if err := got.FromLiteralValue(tt.give); (err != nil) != tt.wantErr {
				t.Errorf("String.FromLiteralValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, &(tt.want)) {
				t.Errorf("String.FromLiteralValue() = %#v, want %#v", got, &(tt.want))
			}
			PermissiveInputParsing = origPermissiveInputParsing
		})
	}
}

func TestInt_FromLiteralValue(t *testing.T) {
	tests := []struct {
		name                   string
		give                   schema.LiteralValue
		permissiveInputParsing bool
		want                   Int
		wantErr                bool
	}{
		{
			name:                   "From Nil",
			give:                   nil,
			permissiveInputParsing: false,
			want:                   NilInt(),
			wantErr:                false,
		},
		{
			name:                   "From Number",
			give:                   schema.LiteralNumber(7654),
			permissiveInputParsing: false,
			want:                   NewInt(7654),
			wantErr:                false,
		},
		{
			name:                   "From Huge Number",
			give:                   schema.LiteralNumber(math.MaxInt32 + 1),
			permissiveInputParsing: false,
			want:                   Int{},
			wantErr:                true,
		},
		{
			name:                   "From String w/out Permissive Input Parsing",
			give:                   schema.LiteralString("7654"),
			permissiveInputParsing: false,
			want:                   Int{},
			wantErr:                true,
		},
		{
			name:                   "From Bool w/out Permissive Input Parsing",
			give:                   schema.LiteralBool(true),
			permissiveInputParsing: false,
			want:                   Int{},
			wantErr:                true,
		},
		{
			name:                   "From Valid String w/ Permissive Input Parsing",
			give:                   schema.LiteralString("7654"),
			permissiveInputParsing: true,
			want:                   NewInt(7654),
			wantErr:                false,
		},
		{
			name:                   "From Invalid String w/ Permissive Input Parsing",
			give:                   schema.LiteralString("123abc"),
			permissiveInputParsing: true,
			want:                   Int{},
			wantErr:                true,
		},
		{
			name:                   "From Bool w/ Permissive Input Parsing",
			give:                   schema.LiteralBool(true),
			permissiveInputParsing: true,
			want:                   NewInt(1),
			wantErr:                false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origPermissiveInputParsing := PermissiveInputParsing
			PermissiveInputParsing = tt.permissiveInputParsing
			got := &Int{}
			if err := got.FromLiteralValue(tt.give); (err != nil) != tt.wantErr {
				t.Errorf("Int.FromLiteralValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, &(tt.want)) {
				t.Errorf("Int.FromLiteralValue() = %#v, want %#v", got, &(tt.want))
			}
			PermissiveInputParsing = origPermissiveInputParsing
		})
	}
}

func TestFloat_FromLiteralValue(t *testing.T) {
	tests := []struct {
		name                   string
		give                   schema.LiteralValue
		permissiveInputParsing bool
		want                   Float
		wantErr                bool
	}{
		{
			name:                   "From Nil",
			give:                   nil,
			permissiveInputParsing: false,
			want:                   NilFloat(),
			wantErr:                false,
		},
		{
			name:                   "From Number",
			give:                   schema.LiteralNumber(7654),
			permissiveInputParsing: false,
			want:                   NewFloat(7654),
			wantErr:                false,
		},
		{
			name:                   "From String w/out Permissive Input Parsing",
			give:                   schema.LiteralString("7654.5674"),
			permissiveInputParsing: false,
			want:                   Float{},
			wantErr:                true,
		},
		{
			name:                   "From Bool w/out Permissive Input Parsing",
			give:                   schema.LiteralBool(true),
			permissiveInputParsing: false,
			want:                   Float{},
			wantErr:                true,
		},
		{
			name:                   "From Valid String w/ Permissive Input Parsing",
			give:                   schema.LiteralString("7654.5674"),
			permissiveInputParsing: true,
			want:                   NewFloat(7654.5674),
			wantErr:                false,
		},
		{
			name:                   "From Invalid String w/ Permissive Input Parsing",
			give:                   schema.LiteralString("duck"),
			permissiveInputParsing: true,
			want:                   Float{},
			wantErr:                true,
		},
		{
			name:                   "From Bool w/ Permissive Input Parsing",
			give:                   schema.LiteralBool(true),
			permissiveInputParsing: true,
			want:                   NewFloat(1),
			wantErr:                false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origPermissiveInputParsing := PermissiveInputParsing
			PermissiveInputParsing = tt.permissiveInputParsing
			got := &Float{}
			if err := got.FromLiteralValue(tt.give); (err != nil) != tt.wantErr {
				t.Errorf("Float.FromLiteralValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, &(tt.want)) {
				t.Errorf("Float.FromLiteralValue() = %#v, want %#v", got, &(tt.want))
			}
			PermissiveInputParsing = origPermissiveInputParsing
		})
	}
}

func TestBoolean_FromLiteralValue(t *testing.T) {
	tests := []struct {
		name                   string
		give                   schema.LiteralValue
		permissiveInputParsing bool
		want                   Boolean
		wantErr                bool
	}{
		{
			name:                   "From Nil",
			give:                   nil,
			permissiveInputParsing: false,
			want:                   NilBoolean(),
			wantErr:                false,
		},
		{
			name:                   "From Bool",
			give:                   schema.LiteralBool(true),
			permissiveInputParsing: false,
			want:                   NewBoolean(true),
			wantErr:                false,
		},
		{
			name:                   "From String w/out Permissive Input Parsing",
			give:                   schema.LiteralString("false"),
			permissiveInputParsing: false,
			want:                   Boolean{},
			wantErr:                true,
		},
		{
			name:                   "From Number w/out Permissive Input Parsing",
			give:                   schema.LiteralNumber(1),
			permissiveInputParsing: false,
			want:                   Boolean{},
			wantErr:                true,
		},
		{
			name:                   "From 'true' String w/ Permissive Input Parsing",
			give:                   schema.LiteralString("true"),
			permissiveInputParsing: true,
			want:                   NewBoolean(true),
			wantErr:                false,
		},
		{
			name:                   "From 'false' String w/ Permissive Input Parsing",
			give:                   schema.LiteralString("false"),
			permissiveInputParsing: true,
			want:                   NewBoolean(false),
			wantErr:                false,
		},
		{
			name:                   "From Invalid String w/ Permissive Input Parsing",
			give:                   schema.LiteralString("duck"),
			permissiveInputParsing: true,
			want:                   Boolean{},
			wantErr:                true,
		},
		{
			name:                   "From Number == 0 w/ Permissive Input Parsing",
			give:                   schema.LiteralNumber(0),
			permissiveInputParsing: true,
			want:                   NewBoolean(false),
			wantErr:                false,
		},
		{
			name:                   "From Number != 0 w/ Permissive Input Parsing",
			give:                   schema.LiteralNumber(-1),
			permissiveInputParsing: true,
			want:                   NewBoolean(true),
			wantErr:                false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origPermissiveInputParsing := PermissiveInputParsing
			PermissiveInputParsing = tt.permissiveInputParsing
			got := &Boolean{}
			if err := got.FromLiteralValue(tt.give); (err != nil) != tt.wantErr {
				t.Errorf("Boolean.FromLiteralValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, &(tt.want)) {
				t.Errorf("Boolean.FromLiteralValue() = %#v, want %#v", got, &(tt.want))
			}
			PermissiveInputParsing = origPermissiveInputParsing
		})
	}
}
