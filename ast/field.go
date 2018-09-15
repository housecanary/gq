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

import "io"

type Field struct {
	Alias        string
	Name         string
	Row          int
	Col          int
	Arguments    Arguments
	Directives   Directives
	SelectionSet SelectionSet
}

func (v *Field) MarshallGraphQL(w io.Writer) error {
	if v.Alias != v.Name {
		if _, err := w.Write([]byte(v.Alias)); err != nil {
			return err
		}
		if _, err := w.Write([]byte(": ")); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte(v.Name)); err != nil {
		return err
	}

	if err := v.Arguments.MarshallGraphQL(w); err != nil {
		return err
	}

	if err := v.Directives.MarshallGraphQL(w); err != nil {
		return err
	}

	if len(v.SelectionSet) > 0 {
		if err := v.SelectionSet.MarshallGraphQL(w); err != nil {
			return err
		}
	}

	return nil
}
