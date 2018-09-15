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
)

type Kind int

const (
	KindSimple Kind = iota
	KindList
	KindNotNil
)

type Type interface {
	Kind() Kind
	Signature() string
	ContainedType() Type
}

type SimpleType struct {
	Name string
}

func (t *SimpleType) Kind() Kind {
	return KindSimple
}

func (t *SimpleType) Signature() string {
	return t.Name
}

func (t *SimpleType) ContainedType() Type {
	return nil
}

type ListType struct {
	Of Type
}

func (t *ListType) Kind() Kind {
	return KindList
}

func (t *ListType) Signature() string {
	return fmt.Sprintf("[%s]", t.Of.Signature())
}

func (t *ListType) ContainedType() Type {
	return t.Of
}

type NotNilType struct {
	Of Type
}

func (t *NotNilType) Kind() Kind {
	return KindNotNil
}

func (t *NotNilType) Signature() string {
	return fmt.Sprintf("%s!", t.Of.Signature())
}

func (t *NotNilType) ContainedType() Type {
	return t.Of
}
