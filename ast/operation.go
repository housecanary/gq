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
	"io"
)

type OperationType string

const (
	OperationTypeQuery        OperationType = "query"
	OperationTypeMutation     OperationType = "mutation"
	OperationTypeSubscription OperationType = "subscription"
)

type OperationDefinition struct {
	OperationType       OperationType
	Name                string
	VariableDefinitions VariableDefinitions
	Directives          Directives
	SelectionSet        SelectionSet
}

func (o *OperationDefinition) MarshallGraphQL(w io.Writer) error {
	lenName := len(o.Name)
	if lenName > 0 {
		lenName++
	}
	if _, err := w.Write([]byte(fmt.Sprintf(`%s%*s`, o.OperationType, lenName, o.Name))); err != nil {
		return err
	}
	if err := o.VariableDefinitions.MarshallGraphQL(w); err != nil {
		return err
	}

	if err := o.Directives.MarshallGraphQL(w); err != nil {
		return err
	}

	if err := o.SelectionSet.MarshallGraphQL(w); err != nil {
		return err
	}
	return nil
}
