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
