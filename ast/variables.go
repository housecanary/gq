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
)

type VariableDefinitions []*VariableDefinition

func (v VariableDefinitions) MarshallGraphQL(w io.Writer) error {
	if len(v) > 0 {
		if _, err := w.Write([]byte("(")); err != nil {
			return err
		}

		for i, def := range v {
			if i > 0 {
				if _, err := w.Write([]byte(", ")); err != nil {
					return err
				}
			}

			if err := def.MarshallGraphQL(w); err != nil {
				return err
			}
		}

		if _, err := w.Write([]byte(")")); err != nil {
			return err
		}

	}
	return nil
}

type VariableDefinition struct {
	VariableName string
	Type         Type
	DefaultValue Value
}

func (v *VariableDefinition) MarshallGraphQL(w io.Writer) error {
	if _, err := w.Write([]byte("$")); err != nil {
		return err
	}

	if _, err := w.Write([]byte(v.VariableName)); err != nil {
		return err
	}

	if _, err := w.Write([]byte(": ")); err != nil {
		return err
	}

	if _, err := w.Write([]byte(v.Type.Signature())); err != nil {
		return err
	}

	if v.DefaultValue != nil {
		if _, err := w.Write([]byte(" = ")); err != nil {
			return err
		}

		if _, err := w.Write([]byte(v.DefaultValue.Representation())); err != nil {
			return err
		}
	}
	return nil
}
