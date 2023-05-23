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
	"context"
	"fmt"
	"reflect"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

type unionTypeBuilder[U UnionT] struct {
	ut  *UnionType[U]
	def *ast.BasicTypeDefinition
}

func (b *unionTypeBuilder[U]) describe() string {
	typ := typeOf[U]()
	return fmt.Sprintf("union %s", typeDesc(typ))
}

func (b *unionTypeBuilder[U]) parse(namePrefix string) (*gqlTypeInfo, reflect.Type, error) {
	return parseTypeDef[U, U](kindUnion, b.ut.def, namePrefix, &b.def)
}

func (b *unionTypeBuilder[U]) build(c *buildContext, sb *schema.Builder) error {
	typeNameMap := make(map[reflect.Type]string)
	var members []string
	for _, t := range b.ut.members {
		st, err := c.astTypeForGoType(t)
		if err != nil {
			return err
		}
		typeNameMap[t] = st.Signature()
		members = append(members, st.Signature())
	}

	tb := sb.AddUnionType(b.def.Name, members, func(ctx context.Context, value interface{}) (interface{}, string) {
		ub := (Union)(value.(U))
		if ub.objectType == nil {
			return nil, ""
		}
		return ub.unionElement, typeNameMap[ub.objectType]
	})
	setSchemaElementProps(tb, b.def.Description, b.def.Directives)
	return nil
}
