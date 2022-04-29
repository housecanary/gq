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

type Document struct {
	OperationDefinitions []*OperationDefinition
	FragmentDefinitions  []*FragmentDefinition
	queryOpsByName       map[string]*OperationDefinition
	fragmentDefsByName   map[string]*FragmentDefinition
}

func (d *Document) AddOperationDefinition(op *OperationDefinition) {
	d.OperationDefinitions = append(d.OperationDefinitions, op)
	if d.queryOpsByName == nil {
		d.queryOpsByName = make(map[string]*OperationDefinition)
	}
	if op.OperationType == OperationTypeQuery {
		d.queryOpsByName[op.Name] = op
	}
}

func (d *Document) AddFragmentDefinition(frag *FragmentDefinition) {
	d.FragmentDefinitions = append(d.FragmentDefinitions, frag)
	if d.fragmentDefsByName == nil {
		d.fragmentDefsByName = make(map[string]*FragmentDefinition)
	}
	d.fragmentDefsByName[frag.Name] = frag
}

func (d *Document) LookupQueryOperation(name string) *OperationDefinition {
	v, ok := d.queryOpsByName[name]

	if ok {
		return v
	}

	if len(d.queryOpsByName) == 1 && name == "" {
		for _, v := range d.queryOpsByName {
			return v
		}
	}

	return nil
}

func (d *Document) LookupFragmentDefinition(name string) *FragmentDefinition {
	v, ok := d.fragmentDefsByName[name]

	if ok {
		return v
	}

	return nil
}

func (d *Document) MarshalGraphQL(w io.Writer) error {
	for _, op := range d.OperationDefinitions {
		if err := op.MarshalGraphQL(w); err != nil {
			return err
		}
	}
	return nil
}
