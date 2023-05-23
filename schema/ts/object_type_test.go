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

package ts

import (
	"context"
	"testing"

	"github.com/housecanary/gq/schema/ts/result"
	"github.com/housecanary/gq/types"
)

func TestObjectFieldResolvers(t *testing.T) {
	m := NewModule()

	ot := NewObjectType[struct {
		AsStruct testObject
		AsPtr    *testObject
		Enum     testEnum
	}](m, "x")

	instance := ot.NewInstance()
	instance.AsPtr = &testObject{"a"}
	instance.AsStruct = testObject{"b"}
	instance.Enum = testEnumA

	tr := mustBuildTypes(t, m)
	structResult := must(tr.QueryField(context.Background(), instance, "asStruct", nil)).Get(t)
	if to, ok := structResult.(*testObject); ok {
		if to.value != "b" {
			t.Fatal("expected value to be b")
		}
	} else {
		t.Fatal("expected struct result to be *testObject")
	}

	ptrResult := must(tr.QueryField(context.Background(), instance, "asPtr", nil)).Get(t)
	if to, ok := ptrResult.(*testObject); ok {
		if to.value != "a" {
			t.Fatal("expected value to be a")
		}
	} else {
		t.Fatal("expected ptr result to be *testObject")
	}

	enumResult := must(tr.QueryField(context.Background(), instance, "enum", nil)).Get(t)
	if enm, ok := enumResult.(testEnum); ok {
		if enm != testEnumA {
			t.Fatal("expected value to be a")
		}
	} else {
		t.Fatal("expected enum result to be testEnum")
	}

}

func TestEmbeddedObjectFieldResolvers(t *testing.T) {
	m := NewModule()

	type t1 struct {
		F2 types.String
	}
	type t2 struct {
		t1
		F1 types.String
	}
	type t3 struct {
		*t1
		F1 types.String
	}

	ot1 := NewObjectType[t1](m, "t1")
	AddField(ot1, "f1", func(p *t1) Result[types.String] {
		return result.Of(types.NewString("t1f1"))
	})
	AddField(ot1, "f2", func(p *t1) Result[types.String] {
		return result.Of(types.NewString("t1f2"))
	})

	ot2 := NewObjectType[t2](m, "t2")
	ot3 := NewObjectType[t3](m, "t3")

	o2 := ot2.NewInstance()
	o2.F1 = types.NewString("t2f1")
	o3 := ot3.NewInstance()
	o3.F1 = types.NewString("t3f1")

	tr := mustBuildTypes(t, m)
	o2f1Result := must(tr.QueryField(context.Background(), o2, "f1", nil)).Get(t)
	if s, ok := o2f1Result.(types.String); ok {
		if s.String() != "t2f1" {
			t.Fatal("expected value to be t2f1")
		}
	} else {
		t.Fatal("expected types.String")
	}

	o2f2Result := must(tr.QueryField(context.Background(), o2, "f2", nil)).Get(t)
	if s, ok := o2f2Result.(Result[types.String]); ok {
		s, _, _ := s.UnpackResult()
		if s.String() != "t1f2" {
			t.Fatal("expected value to be t1f2")
		}
	} else {
		t.Fatal("expected types.String")
	}

	o3f1Result := must(tr.QueryField(context.Background(), o3, "f1", nil)).Get(t)
	if s, ok := o3f1Result.(types.String); ok {
		if s.String() != "t3f1" {
			t.Fatal("expected value to be t3f1")
		}
	} else {
		t.Fatal("expected types.String")
	}

	o3f2Result := must(tr.QueryField(context.Background(), o3, "f2", nil)).Get(t)
	if s, ok := o3f2Result.(Result[types.String]); ok {
		s, _, _ := s.UnpackResult()
		if s.String() != "t1f2" {
			t.Fatal("expected value to be t1f2")
		}
	} else {
		t.Fatal("expected types.String")
	}

}
