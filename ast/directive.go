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

type Directives []*Directive

func (v Directives) MarshalGraphQL(w io.Writer) error {
	for _, d := range v {
		if err := d.MarshalGraphQL(w); err != nil {
			return err
		}
	}

	return nil
}

type Directive struct {
	Name      string
	Arguments Arguments
}

func (v *Directive) MarshalGraphQL(w io.Writer) error {
	if _, err := w.Write([]byte(" @")); err != nil {
		return err
	}

	if _, err := w.Write([]byte(v.Name)); err != nil {
		return err
	}

	if err := v.Arguments.MarshalGraphQL(w); err != nil {
		return err
	}
	return nil
}
