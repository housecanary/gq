package ts

import (
	"context"
	"testing"
)

func TestObjectFieldResolvers(t *testing.T) {
	m := Module()

	ot := Object[struct {
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
