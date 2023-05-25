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
	"fmt"
	"reflect"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser"
	"github.com/housecanary/gq/schema"
)

type inputObjectTypeBuilder[O any] struct {
	it             *InputObjectType[O]
	def            *ast.BasicTypeDefinition
	inputConverter inputConverter
}

func (b *inputObjectTypeBuilder[O]) describe() string {
	typ := typeOf[O]()
	return fmt.Sprintf("input object %s", typeDesc(typ))
}

func (b *inputObjectTypeBuilder[O]) parse(namePrefix string) (*gqlTypeInfo, reflect.Type, error) {
	typ := typeOf[O]()
	if typ.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("Input objects must be represented by a struct, not a %v", typ.Kind())
	}

	return parseTypeDef[O, *O](kindInputObject, b.it.def, namePrefix, &b.def)
}

func (b *inputObjectTypeBuilder[O]) build(c *buildContext, sb *schema.Builder) error {
	decoder, err := b.makeDecoder(c)
	if err != nil {
		return err
	}
	tb := sb.AddInputObjectType(
		b.def.Name,
		decoder,
		reflectionInputListCreator{typeOf[*O]()},
	)

	setSchemaElementProps(tb, b.def.Description, b.def.Directives)

	return b.mapFields(c, tb)
}

func (b *inputObjectTypeBuilder[O]) makeDecoder(c *buildContext) (schema.DecodeInputObject, error) {
	typ := typeOf[O]()

	fieldMap := make(map[string][]int)
	for _, field := range reflect.VisibleFields(typ) {
		if !field.IsExported() {
			continue
		}
		valueDef, _, err := parseStructField(c, field, parser.ParsePartialInputValueDefinition)
		if err != nil {
			return nil, fmt.Errorf("error processing field %s: %w", field.Name, err)
		}
		if valueDef == nil {
			continue
		}
		fieldMap[valueDef.Name] = field.Index
	}

	return func(ctx schema.InputObjectDecodeContext) (interface{}, error) {
		if ctx.IsNil() {
			return (*O)(nil), nil
		}
		var target O
		rv := reflect.ValueOf(&target).Elem()
		for k, v := range fieldMap {
			value, err := ctx.GetFieldValue(k)
			if err != nil {
				return nil, err
			}
			fieldV, err := rv.FieldByIndexErr(v)
			if err != nil {
				return nil, err
			}
			fieldV.Set(reflect.ValueOf(value))
		}
		return &target, nil
	}, nil
}

func (b *inputObjectTypeBuilder[O]) mapFields(c *buildContext, tb *schema.InputObjectTypeBuilder) error {
	typ := typeOf[O]()

	for _, field := range reflect.VisibleFields(typ) {
		if !field.IsExported() {
			continue
		}
		valueDef, _, err := parseStructField(c, field, parser.ParsePartialInputValueDefinition)
		if err != nil {
			return fmt.Errorf("error processing field %s: %w", field.Name, err)
		}
		if valueDef == nil {
			continue
		}
		fb := tb.AddField(valueDef.Name, valueDef.Type, valueDef.DefaultValue)
		setSchemaElementProps(fb, valueDef.Description, valueDef.Directives)
	}

	return nil
}

func (b *inputObjectTypeBuilder[O]) makeInputConverter(c *buildContext) inputConverter {
	if b.inputConverter != nil {
		return b.inputConverter
	}

	typ := typeOf[O]()

	type fieldConversionData struct {
		Name         string
		Index        []int
		Converter    inputConverter
		DefaultValue schema.LiteralValue
	}

	var fields []*fieldConversionData

	b.inputConverter = func(value schema.LiteralValue, dest reflect.Value) error {
		if value == nil {
			dest.Set(reflect.Zero(reflect.PointerTo(typ)))
		}

		var target O
		rv := reflect.ValueOf(&target).Elem()
		if obj, ok := value.(schema.LiteralObject); ok {
			for _, f := range fields {
				target, err := rv.FieldByIndexErr(f.Index)
				if err != nil {
					return err
				}

				var fieldVal schema.LiteralValue
				if fv, ok := obj[f.Name]; ok {
					fieldVal = fv
				} else {
					fieldVal = f.DefaultValue
				}
				if err := f.Converter(fieldVal, target); err != nil {
					return fmt.Errorf("invalid input for field %s: %w", f.Name, err)
				}
			}
		}

		dest.Set(reflect.ValueOf(&target))
		return nil
	}

	for _, field := range reflect.VisibleFields(typ) {
		if !field.IsExported() {
			continue
		}

		valueDef, _, err := parseStructField(c, field, parser.ParsePartialInputValueDefinition)
		if err != nil {
			// We'll log an error when we build this field into the schema later on, so no need to report now
			return nil
		}

		if valueDef == nil {
			continue
		}

		inputConverter := makeInputConverterForType(c, valueDef.Type, field.Type)
		fields = append(fields, &fieldConversionData{
			Name:         valueDef.Name,
			Index:        field.Index,
			Converter:    inputConverter,
			DefaultValue: schema.LiteralValueFromAstValue(valueDef.DefaultValue),
		})
	}

	return b.inputConverter
}
