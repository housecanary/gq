package ts

import (
	"testing"
)

var testMod = NewModule()

type testObject struct {
	value string
}

var objectType = NewObjectType[testObject](testMod, ``)

type testEnum string

var enumType = NewEnumType[testEnum](testMod, ``)

var testEnumA = enumType.Value(`A`)
var testEnumB = enumType.Value(`B`)

func mustBuildTypes(t *testing.T, mods ...*Module) *TypeRegistry {
	var opts []TypeRegistryOption
	for _, mod := range mods {
		opts = append(opts, WithModule(mod))
	}
	opts = append(opts, WithModule(testMod))
	tr, err := NewTypeRegistry(opts...)
	if err != nil {
		t.Fatal(err)
	}
	return tr
}

type mustBox[T any] struct {
	v   T
	err error
}

func (b *mustBox[T]) Get(t *testing.T) T {
	if b.err != nil {
		t.Fatal(b.err)
	}
	return b.v
}

func must[T any](v T, err error) *mustBox[T] {
	return &mustBox[T]{v, err}
}
