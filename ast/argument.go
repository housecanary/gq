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

type Arguments []*Argument

func (v Arguments) MarshalGraphQL(w io.Writer) error {
	if len(v) > 0 {
		if _, err := w.Write([]byte("(")); err != nil {
			return err
		}

		for i, a := range v {
			if i > 0 {
				if _, err := w.Write([]byte(", ")); err != nil {
					return err
				}
			}

			if err := a.MarshalGraphQL(w); err != nil {
				return err
			}
		}

		if _, err := w.Write([]byte(")")); err != nil {
			return err
		}
	}
	return nil
}

// ByName finds the argument with the given name, setting found to indicate if it was found
func (v Arguments) ByName(name string) (argument *Argument, found bool) {
	for _, a := range v {
		if a.Name == name {
			argument = a
			found = true
			return
		}
	}
	return
}

type Argument struct {
	Name  string
	Value Value
}

func (v *Argument) MarshalGraphQL(w io.Writer) error {
	if _, err := w.Write([]byte(v.Name)); err != nil {
		return err
	}

	if _, err := w.Write([]byte(": ")); err != nil {
		return err
	}

	if _, err := w.Write([]byte(v.Value.Representation())); err != nil {
		return err
	}
	return nil
}

type ArgumentsDefinition []*InputValueDefinition

type InputValueDefinition struct {
	Description  string
	Name         string
	Type         Type
	DefaultValue Value
	Directives   Directives
}
