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

type TypeRegistry struct {
	sb          *schema.Builder
	objectTypes map[reflect.Type]map[string]*fieldRuntimeInfo
}

// NewTypeRegistry creates a schema.Builder from a set of ts modules.
func NewTypeRegistry(opts ...TypeRegistryOption) (*TypeRegistry, error) {
	cnf := typeRegistryConfig{
		providers: make(map[reflect.Type]func(QueryInfo) any),
	}

	WithProvider(func(qi QueryInfo) QueryInfo {
		return qi
	}).apply(&cnf)

	for _, opt := range opts {
		opt.apply(&cnf)
	}

	bc := newBuildContext(cnf.providers)
	sb := schema.NewBuilder()
	var allTypes []builderType
	for _, mc := range append(cnf.mods, &ModuleConfig{mod: BuiltinTypes}) {
		types := mc.mod.elements
		for _, bt := range types {
			gqlType, goType, err := bt.parse(mc.mod.typePrefix)
			if err != nil {
				return nil, fmt.Errorf("cannot parse type %s: %w", bt.describe(), err)
			}
			bc.goTypeToSchemaType[goType] = gqlType
			bc.goTypeToBuilder[goType] = bt
			if _, excluded := mc.exclude[gqlType.astType.Signature()]; !excluded {
				allTypes = append(allTypes, bt)
			}
		}
	}

	for _, bt := range allTypes {
		err := bt.build(bc, sb)
		if err != nil {
			return nil, fmt.Errorf("cannot build type %s: %w", bt.describe(), err)
		}
	}

	return &TypeRegistry{sb, bc.objectTypes}, nil
}

// BuildSchema creates a new GQL schema from this type registry
func (ts *TypeRegistry) BuildSchema(queryTypeName string) (*schema.Schema, error) {
	return ts.sb.Build(queryTypeName)
}

// MustBuildSchema creates a new GQL schema from this type registry, panicing on error
func (ts *TypeRegistry) MustBuildSchema(queryTypeName string) *schema.Schema {
	return ts.sb.MustBuild(queryTypeName)
}

// SchemaBuilder creates a new GQL schema builder from this type registry
func (ts *TypeRegistry) SchemaBuilder() *schema.Builder {
	return ts.sb
}

// QueryField queries the given gql field on an object with a type registered as a GQL object type.
//
// The return value is the raw value returned from the field: either a Result[T] or the raw value of a struct field
func (ts *TypeRegistry) QueryField(ctx context.Context, o any, name string, args map[string]interface{}, childSelections ...*QueryFieldSelection) (interface{}, error) {
	fi, err := ts.getFieldRuntimeInfo(reflect.TypeOf(o), name)
	if err != nil {
		return nil, err
	}
	return fi.invoker(&invokeQueryInfo{
		ctx:      ctx,
		args:     args,
		children: childSelections,
	}, o), nil
}

// GetBackingField returns a reflect.Value that points to the struct field that backs a GQL field.
//
// If the given field is not backed by a struct field, or the field is only traversable via a nil pointer, an error is returned
func (ts *TypeRegistry) GetBackingField(o any, name string) (reflect.Value, error) {
	fi, err := ts.getFieldRuntimeInfo(reflect.TypeOf(o), name)
	if err != nil {
		return reflect.ValueOf(nil), err
	}
	if fi.sourceField.Index == nil {
		return reflect.ValueOf(nil), fmt.Errorf("field %s is virtual", name)
	}

	return reflect.ValueOf(o).Elem().FieldByIndexErr(fi.sourceField.Index)
}

// GetFieldTag returns a reflect.StructTag from the the struct field that backs a GQL field.
//
// If the given field is not backed by a struct field, an empty struct tag is returned
func (ts *TypeRegistry) GetFieldTag(typ reflect.Type, name string) (reflect.StructTag, error) {
	fi, err := ts.getFieldRuntimeInfo(typ, name)
	if err != nil {
		return "", err
	}
	return fi.sourceField.Tag, nil
}

// GetFields returns all the fields registered on a type
func (ts *TypeRegistry) GetFields(typ reflect.Type) ([]string, error) {
	fMap, ok := ts.objectTypes[typ]
	if !ok {
		return nil, fmt.Errorf("type %v does not match any registered types", typ)
	}

	fields := make([]string, 0, len(fMap))
	for k := range fMap {
		fields = append(fields, k)
	}
	return fields, nil
}

func (ts *TypeRegistry) getFieldRuntimeInfo(typ reflect.Type, name string) (*fieldRuntimeInfo, error) {
	fMap, ok := ts.objectTypes[typ]
	if !ok {
		return nil, fmt.Errorf("type %v does not match any registered types", typ)
	}

	fi, ok := fMap[name]
	if !ok {
		return nil, fmt.Errorf("field %s does not exist", name)
	}
	return fi, nil
}

type TypeRegistryOption interface {
	apply(config *typeRegistryConfig)
}

type typeRegistryConfig struct {
	mods      []*ModuleConfig
	providers map[reflect.Type]func(QueryInfo) any
}

type typeRegistryOptionFunc func(config *typeRegistryConfig)

func (f typeRegistryOptionFunc) apply(config *typeRegistryConfig) {
	f(config)
}

// WithModule adds a module to the type registry
func WithModule(mod *Module) *ModuleConfig {
	return &ModuleConfig{
		mod:     mod,
		exclude: make(map[string]struct{}),
	}
}

type ModuleConfig struct {
	mod     *Module
	exclude map[string]struct{}
}

func (mc *ModuleConfig) apply(config *typeRegistryConfig) {
	config.mods = append(config.mods, mc)
}

func (mc *ModuleConfig) Excluding(names ...string) *ModuleConfig {
	for _, name := range names {
		mc.exclude[name] = struct{}{}
	}
	return mc
}

// WithProvider adds a provider for type T
func WithProvider[T any](f func(QueryInfo) T) TypeRegistryOption {
	return typeRegistryOptionFunc(func(config *typeRegistryConfig) {
		config.providers[typeOf[T]()] = func(qi QueryInfo) any {
			return f(qi)
		}
	})
}

type QueryFieldSelection struct {
	Selection   *ast.Field
	SchemaField *schema.FieldDescriptor
	Children    []*QueryFieldSelection
}

type fieldInvoker func(q *invokeQueryInfo, o any) any

type fieldRuntimeInfo struct {
	sourceField reflect.StructField
	invoker     fieldInvoker
}

type invokeQueryInfo struct {
	ctx      context.Context
	args     map[string]any
	children []*QueryFieldSelection
}

func (q *invokeQueryInfo) ArgumentValue(name string) (any, error) {
	return q.args[name], nil
}

func (q *invokeQueryInfo) QueryContext() context.Context {
	return q.ctx
}

func (q *invokeQueryInfo) ChildFieldsIterator() schema.FieldSelectionIterator {
	return &queryFieldSelectionIterator{}
}

func (q *invokeQueryInfo) setArgumentValue(name string, dest reflect.Value, converter inputConverter) error {
	av := q.args[name]
	rv := reflect.ValueOf(av)
	if !rv.Type().AssignableTo(dest.Type()) && av != nil {
		return fmt.Errorf("cannot assign a value of type %T to a destination of type %T for argument %s", av, dest.Interface(), name)
	}
	dest.Set(rv)
	return nil
}

type queryFieldSelectionIterator struct {
	head *QueryFieldSelection
	tail []*QueryFieldSelection
}

func (i *queryFieldSelectionIterator) Next() bool {
	if len(i.tail) == 0 {
		return false
	}
	i.head = i.tail[0]
	i.tail = i.tail[1:]
	return true
}

func (i *queryFieldSelectionIterator) Selection() *ast.Field {
	return i.head.Selection
}
func (i *queryFieldSelectionIterator) SchemaField() *schema.FieldDescriptor {
	return i.head.SchemaField
}
func (i *queryFieldSelectionIterator) ChildFieldsIterator() schema.FieldSelectionIterator {
	return &queryFieldSelectionIterator{nil, i.head.Children}
}
