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
func NewTypeRegistry(mods ...*ModuleType) (*TypeRegistry, error) {
	bc := newBuildContext()
	sb := schema.NewBuilder()
	var allTypes []builderType
	for _, mod := range append(mods, BuiltinTypes) {
		types := mod.elements
		for _, bt := range types {
			gqlType, goType, err := bt.parse(mod.typePrefix)
			if err != nil {
				return nil, fmt.Errorf("cannot parse type %s: %w", bt.describe(), err)
			}
			bc.goTypeToSchemaType[goType] = gqlType
		}
		allTypes = append(allTypes, types...)
	}

	// TODO: should we allow the same type to be registered in multiple modules?

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
// If the given field is not backed by a struct field, an error is returned
func (ts *TypeRegistry) GetBackingField(o any, name string) (reflect.Value, error) {
	fi, err := ts.getFieldRuntimeInfo(reflect.TypeOf(o), name)
	if err != nil {
		return reflect.ValueOf(nil), err
	}
	if fi.sourceField.Index == nil {
		return reflect.ValueOf(nil), fmt.Errorf("field %s is virtual", name)
	}

	return reflect.ValueOf(o).Elem().FieldByIndex(fi.sourceField.Index), nil
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

type QueryFieldSelection struct {
	Selection   *ast.Field
	SchemaField *schema.FieldDescriptor
	Children    []*QueryFieldSelection
}

type fieldInvoker func(q QueryInfo, o interface{}) interface{}

type fieldRuntimeInfo struct {
	sourceField reflect.StructField
	invoker     fieldInvoker
}

type invokeQueryInfo struct {
	ctx      context.Context
	args     map[string]interface{}
	children []*QueryFieldSelection
}

func (q *invokeQueryInfo) ArgumentValue(name string) (interface{}, error) {
	return q.args[name], nil
}

func (q *invokeQueryInfo) QueryContext() context.Context {
	return q.ctx
}

func (q *invokeQueryInfo) ChildFieldsIterator() schema.FieldSelectionIterator {
	return &queryFieldSelectionIterator{}
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
