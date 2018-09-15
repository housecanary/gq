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
	"io"

	text "github.com/tonnerre/golang-text"
)

type SelectionSet []Selection

func (v SelectionSet) MarshallGraphQL(w io.Writer) error {
	if _, err := w.Write([]byte(" {")); err != nil {
		return err
	}

	iw := text.NewIndentWriter(w, []byte(""), []byte("  "))
	for _, sel := range v {
		if _, err := iw.Write([]byte("\n")); err != nil {
			return err
		}

		if err := sel.MarshallGraphQL(iw); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte("\n}")); err != nil {
		return err
	}

	return nil
}

type Selection interface {
	GraphQLMarshaller
}

type FieldSelection struct {
	Field Field
}

func (v *FieldSelection) MarshallGraphQL(w io.Writer) error {
	return v.Field.MarshallGraphQL(w)
}

type FragmentSpreadSelection struct {
	FragmentName string
	Directives   Directives
}

func (v *FragmentSpreadSelection) MarshallGraphQL(w io.Writer) error {
	if _, err := w.Write([]byte("...")); err != nil {
		return err
	}

	if _, err := w.Write([]byte(v.FragmentName)); err != nil {
		return err
	}

	if err := v.Directives.MarshallGraphQL(w); err != nil {
		return err
	}
	return nil
}

type InlineFragmentSelection struct {
	OnType       string
	Directives   Directives
	SelectionSet SelectionSet
}

func (v *InlineFragmentSelection) MarshallGraphQL(w io.Writer) error {
	if _, err := w.Write([]byte("... on ")); err != nil {
		return err
	}

	if _, err := w.Write([]byte(v.OnType)); err != nil {
		return err
	}

	if err := v.SelectionSet.MarshallGraphQL(w); err != nil {
		return err
	}
	return nil
}
