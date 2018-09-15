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

package gen

import (
	"fmt"
	"go/types"

	"github.com/codemodus/kace"

	"github.com/housecanary/gq/internal/pkg/parser"
)

func (c *genCtx) processInterfaceType(typ *types.Named) (*interfaceMeta, error) {
	// Find and parse the meta field that contains partial GraphQL definition
	// of this type
	var err error
	f := fieldByName(typ, "Interface")
	gqlTypeDef, err := parser.ParsePartialInterfaceTypeDefinition(string(f.tag))
	if err != nil {
		return nil, fmt.Errorf("Cannot parse GQL metadata for interface %s: %v", typ.Obj().Id(), err)
	}
	td := gqlTypeDef

	// Assign name if not defined in GraphQL
	if td.Name == "" {
		td.Name = kace.Pascal(typ.Obj().Name())
	}

	if err := c.checkRegistration(td.Name, typ); err != nil {
		return nil, err
	}

	if existing, ok := c.meta[td.Name]; ok {
		return existing.(*interfaceMeta), nil
	}

	meta := &interfaceMeta{
		baseMeta: baseMeta{
			name:      td.Name,
			namedType: typ,
		},
		GQL:           td,
		InterfaceType: f.field.Type(),
	}
	c.meta[td.Name] = meta
	return meta, nil
}
