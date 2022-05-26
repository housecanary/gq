package ts

import "testing"

var testMod = Module()

type testObject struct {
	value string
}

var objectType = Object[testObject](testMod, ``)

type testEnum string

var enumType = Enum[testEnum](testMod, ``)

var testEnumA = enumType.Value(`A`)
var testEnumB = enumType.Value(`B`)

func mustBuildTypes(t *testing.T, mods ...*ModuleType) *TypeRegistry {
	tr, err := NewTypeRegistry(append(mods, testMod)...)
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
