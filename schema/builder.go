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

package schema

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"

	"github.com/housecanary/gq/ast"
)

// BuilderSchemaElement defines common properties settable on schema elements
type BuilderSchemaElement interface {
	SetDescription(string)
	AddDirective(name string) *DirectiveBuilder
}

type builderSchemaElement struct {
	description string
	directives  []*DirectiveBuilder
}

func (b *builderSchemaElement) SetDescription(desc string) {
	b.description = desc
}

func (b *builderSchemaElement) AddDirective(name string) *DirectiveBuilder {
	db := &DirectiveBuilder{
		named: named{name},
	}

	b.directives = append(b.directives, db)
	return db
}

func (b *builderSchemaElement) toSchemaElement() schemaElement {
	directives := make([]*Directive, len(b.directives))
	for i, e := range b.directives {
		directives[i] = &Directive{
			named:     e.named,
			arguments: e.arguments,
		}
	}
	return schemaElement{b.description, directives}
}

// NewBuilder creates a new schema builder
func NewBuilder() *Builder {
	return &Builder{
		make(map[string]typeBuilder),
		make(map[string]Type),
		nil,
		nil,
		false,
	}
}

// A Builder is used to create a Schema
type Builder struct {
	typeBuilders         map[string]typeBuilder
	resolvedTypes        map[string]Type
	directives           []*DirectiveDefinitionBuilder
	deferredErrors       []error
	disableIntrospection bool
}

type typeBuilder interface {
	registerType(ctx *buildContext) buildError
	Name() string
}

// An ObjectTypeBuilder is used to construct an ObjectType
type ObjectTypeBuilder struct {
	builder *Builder
	named
	builderSchemaElement
	implements []string
	fields     []*ObjectFieldBuilder
}

// An ObjectFieldBuilder is used to configure a single object field
type ObjectFieldBuilder struct {
	named
	builderSchemaElement
	args     []*InputValueDefinitionBuilder
	typ      ast.Type
	resolver Resolver
}

// An InputValueDefinitionBuilder is used to configure a single input value
type InputValueDefinitionBuilder struct {
	named
	builderSchemaElement
	typ          ast.Type
	defaultValue ast.Value
}

// An InterfaceTypeBuilder is used to construct an InterfaceType
type InterfaceTypeBuilder struct {
	builder *Builder
	named
	builderSchemaElement
	fields []*ObjectFieldBuilder
	unwrap UnwrapInterface
}

// A UnionTypeBuilder is used to construct a UnionType
type UnionTypeBuilder struct {
	builder *Builder
	named
	builderSchemaElement
	members []string
	unwrap  UnwrapUnion
}

// A ScalarTypeBuilder is used to construct a ScalarType
type ScalarTypeBuilder struct {
	builder *Builder
	named
	builderSchemaElement
	encode      EncodeScalar
	decode      DecodeScalar
	listCreator InputListCreator
}

// A InputObjectTypeBuilder is used to construct a InputObjectType
type InputObjectTypeBuilder struct {
	builder *Builder
	named
	builderSchemaElement
	decode      DecodeInputObject
	listCreator InputListCreator
	fields      []*InputObjectFieldBuilder
}

// A InputObjectFieldBuilder is used to configure a field of an input object
type InputObjectFieldBuilder struct {
	builder *InputObjectTypeBuilder
	named
	builderSchemaElement
	typ          ast.Type
	defaultValue ast.Value
}

// A EnumTypeBuilder is used to construct a EnumType
type EnumTypeBuilder struct {
	builder *Builder
	named
	builderSchemaElement
	encode      EncodeEnum
	decode      DecodeEnum
	listCreator InputListCreator
	values      []*EnumValueBuilder
}

// An EnumValueBuilder is used to configure an enum value
type EnumValueBuilder struct {
	named
	builderSchemaElement
}

// A DirectiveDefinitionBuilder is used to construct a directive definition
type DirectiveDefinitionBuilder struct {
	named
	description string
	args        []*InputValueDefinitionBuilder
	locations   []DirectiveLocation
}

// A DirectiveBuilder is used to construct a directive
type DirectiveBuilder struct {
	named
	arguments []*DirectiveArgument
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type buildContext struct {
	pth []string
}

type buildError interface {
	error
	path() []string
}

type simpleBuildError struct {
	msg string
	pth []string
}

func (e *simpleBuildError) path() []string {
	return e.pth
}

func (e *simpleBuildError) Error() string {
	return fmt.Sprintf("Error at path %v: %s", strings.Join(e.pth, "."), e.msg)
}

func (c *buildContext) pushPathElement(text string) int {
	c.pth = append(c.pth, text)
	return len(c.pth) - 1
}

func (c *buildContext) popPathElement(to int) {
	c.pth = c.pth[:to]
}

func (c *buildContext) path() []string {
	path := make([]string, len(c.pth))
	copy(path, c.pth)
	return path
}

func (c *buildContext) error(f string, args ...interface{}) buildError {
	return &simpleBuildError{fmt.Sprintf(f, args...), c.path()}
}

// Build creates the final schema from all of the currently configured
// types
func (b *Builder) Build(queryTypeName string) (*Schema, error) {
	if len(b.deferredErrors) > 0 {
		return nil, &multierror.Error{
			Errors: b.deferredErrors,
		}
	}

	var ctx buildContext

	for k, v := range b.typeBuilders {
		if _, ok := b.resolvedTypes[k]; !ok {
			err := v.registerType(&ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	for k, v := range b.typeBuilders {
		if otb, ok := v.(*ObjectTypeBuilder); ok {
			for _, name := range otb.implements {
				if it, ok := b.resolvedTypes[name].(*InterfaceType); ok {
					ot := b.resolvedTypes[k].(*ObjectType)
					it.implementations = append(it.implementations, ot)
				} else {
					return nil, fmt.Errorf("Type %s is not an interface type", name)
				}
			}
		}
	}

	directives := make([]*DirectiveDefinition, len(b.directives))
	directivesByName := make(map[string]bool)
	for i, d := range b.directives {
		if _, ok := directivesByName[d.name]; ok {
			return nil, fmt.Errorf("Duplicate directive definition %s", d.name)
		}
		directivesByName[d.name] = true

		args := make([]*ArgumentDescriptor, len(d.args))
		argsByName := make(map[string]bool)
		for i, a := range d.args {
			if _, ok := argsByName[a.name]; ok {
				return nil, fmt.Errorf("Duplicate argument %s in directive definition %s", a.name, d.name)
			}
			argsByName[a.name] = true
			argType, err := b.resolveAstType(&ctx, a.typ)
			if err != nil {
				return nil, err
			}
			args[i] = &ArgumentDescriptor{
				named:         a.named,
				schemaElement: a.toSchemaElement(),
				typ:           argType,
				defaultValue:  a.defaultValue,
			}
		}
		directives[i] = &DirectiveDefinition{
			named:       d.named,
			description: d.description,
			arguments:   args,
			locations:   d.locations,
		}
	}

	qtc, ok := b.resolvedTypes[queryTypeName]
	if !ok {
		return nil, fmt.Errorf("Query type %s does not exist", queryTypeName)
	}
	qt, ok := qtc.(*ObjectType)
	if !ok {
		return nil, fmt.Errorf("Query type %s is not an object type", queryTypeName)
	}

	s := &Schema{QueryType: qt, allTypes: b.resolvedTypes, directives: directives}

	if !b.disableIntrospection {
		// Add in introspection meta fields
		qt.fieldsByName["__schema"] = &FieldDescriptor{
			named: named{"__schema"},
			typ:   introspectionSchemaType,
			r: SimpleResolver(func(v interface{}) (interface{}, error) {
				return s, nil
			}),
		}

		qt.fieldsByName["__type"] = &FieldDescriptor{
			named: named{"__type"},
			typ:   introspectionTypeType,
			arguments: []*ArgumentDescriptor{
				&ArgumentDescriptor{
					named: named{"name"},
					typ:   &NotNilType{introspectionStringType},
				},
			},
			r: FullResolver(func(ctx ResolverContext, v interface{}) (interface{}, error) {
				name, err := ctx.GetArgumentValue("name")
				if err != nil {
					return nil, err
				}
				return s.allTypes[name.(string)], nil
			}),
		}
	}

	return s, nil
}

// MustBuild is the same as Build, but panics on error
func (b *Builder) MustBuild(queryTypeName string) *Schema {
	s, err := b.Build(queryTypeName)
	must(err)
	return s
}

// DisableIntrospection disables introspection in this builder
func (b *Builder) DisableIntrospection() {
	b.disableIntrospection = true
}

func (b *Builder) addTypeBuilder(tb typeBuilder) {
	if _, ok := b.typeBuilders[tb.Name()]; ok {
		b.deferredErrors = append(b.deferredErrors, fmt.Errorf("Type %s already registered", tb.Name()))
		return
	}
	b.typeBuilders[tb.Name()] = tb
}

func (b *Builder) resolveAstType(ctx *buildContext, typ ast.Type) (Type, buildError) {
	switch t := typ.(type) {
	case *ast.SimpleType:
		if rt, ok := b.resolvedTypes[t.Name]; ok {
			return rt, nil
		}
		if tb, ok := b.typeBuilders[t.Name]; ok {
			err := tb.registerType(ctx)
			if err != nil {
				return nil, err
			}
			return b.resolvedTypes[t.Name], nil
		}
		return nil, ctx.error("Unknown type %s", t.Name)
	case *ast.ListType:
		of, err := b.resolveAstType(ctx, t.Of)
		if err != nil {
			return nil, err
		}
		return &ListType{of}, nil
	case *ast.NotNilType:
		of, err := b.resolveAstType(ctx, t.Of)
		if err != nil {
			return nil, err
		}
		return &NotNilType{of}, nil
	default:
		return nil, ctx.error("Unknown ast type %v", typ)
	}
}

// AddObjectType adds an object type to the schema being built
func (b *Builder) AddObjectType(name string) *ObjectTypeBuilder {
	ob := &ObjectTypeBuilder{
		builder: b,
		named:   named{name},
	}

	b.addTypeBuilder(ob)
	return ob
}

// AddField adds an field to the object being built
func (b *ObjectTypeBuilder) AddField(name string, typ ast.Type, resolver Resolver) *ObjectFieldBuilder {
	fb := &ObjectFieldBuilder{
		named:    named{name},
		typ:      typ,
		resolver: resolver,
	}

	b.fields = append(b.fields, fb)
	return fb
}

// AddArgument adds an argument to the field
func (b *ObjectFieldBuilder) AddArgument(name string, typ ast.Type, defaultValue ast.Value) *InputValueDefinitionBuilder {
	ab := &InputValueDefinitionBuilder{
		named:        named{name},
		typ:          typ,
		defaultValue: defaultValue,
	}
	b.args = append(b.args, ab)
	return ab
}

// Implements adds an interface to the object being built
func (b *ObjectTypeBuilder) Implements(name string) {
	b.implements = append(b.implements, name)
}

func (b *ObjectTypeBuilder) registerType(ctx *buildContext) buildError {
	ctxLvl := ctx.pushPathElement(fmt.Sprintf("[object %s]", b.name))
	defer func() { ctx.popPathElement(ctxLvl) }()

	ot := &ObjectType{
		named:         b.named,
		schemaElement: b.toSchemaElement(),
	}
	b.builder.resolvedTypes[b.name] = ot

	interfaces := make([]*InterfaceType, len(b.implements))
	for i, e := range b.implements {
		t, err := b.builder.resolveAstType(ctx, &ast.SimpleType{Name: e})
		if err != nil {
			return err
		}
		it, ok := t.(*InterfaceType)
		if !ok {
			return ctx.error("Type %s is not an interface type", e)
		}
		interfaces[i] = it
	}
	ot.interfaces = interfaces

	fieldsByName := make(map[string]*FieldDescriptor)
	for _, f := range b.fields {
		ctxLvl := ctx.pushPathElement(fmt.Sprintf("[field %s]", f.name))
		if _, ok := fieldsByName[f.name]; ok {
			return ctx.error("Duplicate field definition")
		}

		fieldType, err := b.builder.resolveAstType(ctx, f.typ)
		if err != nil {
			return err
		}

		argsByName := make(map[string]bool)
		args := make([]*ArgumentDescriptor, len(f.args))
		for i, a := range f.args {
			ctxLvl := ctx.pushPathElement(fmt.Sprintf("[arg %s]", a.name))
			if _, ok := argsByName[a.name]; ok {
				return ctx.error("Duplicate argument definition")
			}
			argsByName[a.name] = true
			argType, err := b.builder.resolveAstType(ctx, a.typ)
			if err != nil {
				return err
			}
			if !isValidArgumentType(argType) {
				return ctx.error("Invalid type %v, must be input object or scalar", argType)
			}
			args[i] = &ArgumentDescriptor{
				named:         a.named,
				schemaElement: a.toSchemaElement(),
				typ:           argType,
				defaultValue:  a.defaultValue,
			}
			ctx.popPathElement(ctxLvl)
		}
		fd := &FieldDescriptor{
			f.named,
			f.toSchemaElement(),
			args,
			fieldType,
			f.resolver,
		}
		fieldsByName[f.name] = fd
		ctx.popPathElement(ctxLvl)
	}

	fieldsByName["__typename"] = &FieldDescriptor{
		named{"__typename"},
		schemaElement{},
		nil,
		&NotNilType{introspectionStringType},
		SimpleResolver(func(interface{}) (interface{}, error) {
			return ot.name, nil
		}),
	}

	ot.fieldsByName = fieldsByName
	return nil
}

// AddInterfaceType adds an interface type to the schema being built
func (b *Builder) AddInterfaceType(name string, unwrap UnwrapInterface) *InterfaceTypeBuilder {
	ib := &InterfaceTypeBuilder{
		builder: b,
		named:   named{name},
		unwrap:  unwrap,
	}

	b.addTypeBuilder(ib)
	return ib
}

// AddField adds an field to the interface being built
func (b *InterfaceTypeBuilder) AddField(name string, typ ast.Type) *ObjectFieldBuilder {
	fb := &ObjectFieldBuilder{
		named: named{name},
		typ:   typ,
	}

	b.fields = append(b.fields, fb)
	return fb
}

func (b *InterfaceTypeBuilder) registerType(ctx *buildContext) buildError {
	ctxLvl := ctx.pushPathElement(fmt.Sprintf("[interface %s]", b.name))
	defer func() { ctx.popPathElement(ctxLvl) }()
	t := &InterfaceType{
		named:         named{b.name},
		schemaElement: b.toSchemaElement(),
		unwrap:        b.unwrap,
	}
	b.builder.resolvedTypes[b.name] = t

	fieldsByName := make(map[string]*FieldDescriptor)
	for _, f := range b.fields {
		ctxLvl := ctx.pushPathElement(fmt.Sprintf("[field %s]", f.name))
		fieldType, err := b.builder.resolveAstType(ctx, f.typ)
		if err != nil {
			return err
		}
		args := make([]*ArgumentDescriptor, len(f.args))
		argsByName := make(map[string]bool)
		for i, a := range f.args {
			ctxLvl := ctx.pushPathElement(fmt.Sprintf("[arg %s]", a.name))
			if _, ok := argsByName[a.name]; ok {
				return ctx.error("Duplicate argument definition")
			}
			argType, err := b.builder.resolveAstType(ctx, a.typ)
			if err != nil {
				return err
			}
			if !isValidArgumentType(argType) {
				return ctx.error("Invalid type %v, must be input object or scalar", argType)
			}
			args[i] = &ArgumentDescriptor{
				named:         a.named,
				schemaElement: a.toSchemaElement(),
				typ:           argType,
				defaultValue:  a.defaultValue,
			}
			ctx.popPathElement(ctxLvl)
		}
		fd := &FieldDescriptor{
			f.named,
			f.toSchemaElement(),
			args,
			fieldType,
			nil,
		}
		fieldsByName[f.name] = fd
		ctx.popPathElement(ctxLvl)
	}

	fieldsByName["__typename"] = &FieldDescriptor{
		named{"__typename"},
		schemaElement{},
		nil,
		&NotNilType{introspectionStringType},
		nil,
	}

	t.fields = fieldsByName
	return nil
}

// AddUnionType adds a union type to the schema being built
func (b *Builder) AddUnionType(name string, members []string, unwrap UnwrapUnion) *UnionTypeBuilder {
	ub := &UnionTypeBuilder{
		builder: b,
		named:   named{name},
		members: members,
		unwrap:  unwrap,
	}

	b.addTypeBuilder(ub)
	return ub
}

func (b *UnionTypeBuilder) registerType(ctx *buildContext) buildError {
	ctxLvl := ctx.pushPathElement(fmt.Sprintf("[union %s]", b.name))
	defer func() { ctx.popPathElement(ctxLvl) }()
	t := &UnionType{
		named:         b.named,
		schemaElement: b.toSchemaElement(),
		unwrap:        b.unwrap,
	}
	b.builder.resolvedTypes[b.name] = t

	members := make([]*ObjectType, len(b.members))
	for i, name := range b.members {
		memberType, err := b.builder.resolveAstType(ctx, &ast.SimpleType{Name: name})
		if err != nil {
			return err
		}
		if ot, ok := memberType.(*ObjectType); ok {
			members[i] = ot
		} else {
			return ctx.error("Member type %s is not an object type", name)
		}
	}

	t.members = members
	return nil
}

// AddScalarType adds a scalar type to the schema being built
func (b *Builder) AddScalarType(name string, encode EncodeScalar, decode DecodeScalar, listCreator InputListCreator) *ScalarTypeBuilder {
	sb := &ScalarTypeBuilder{
		builder:     b,
		named:       named{name},
		encode:      encode,
		decode:      decode,
		listCreator: listCreator,
	}

	b.addTypeBuilder(sb)
	return sb
}

func (b *ScalarTypeBuilder) registerType(ctx *buildContext) buildError {
	t := &ScalarType{
		named:         b.named,
		schemaElement: b.toSchemaElement(),
		encode:        b.encode,
		decode:        b.decode,
		listCreator:   b.listCreator,
	}
	b.builder.resolvedTypes[b.name] = t
	return nil
}

// AddInputObjectType adds an input object type to the schema being built
func (b *Builder) AddInputObjectType(name string, decode DecodeInputObject, listCreator InputListCreator) *InputObjectTypeBuilder {
	sb := &InputObjectTypeBuilder{
		builder:     b,
		named:       named{name},
		decode:      decode,
		listCreator: listCreator,
	}

	b.addTypeBuilder(sb)
	return sb
}

// AddField adds an field to the object being built
func (b *InputObjectTypeBuilder) AddField(name string, typ ast.Type, defaultValue ast.Value) *InputObjectFieldBuilder {
	fb := &InputObjectFieldBuilder{
		builder:      b,
		named:        named{name},
		typ:          typ,
		defaultValue: defaultValue,
	}
	b.fields = append(b.fields, fb)
	return fb
}

func (b *InputObjectTypeBuilder) registerType(ctx *buildContext) buildError {
	ctxLvl := ctx.pushPathElement(fmt.Sprintf("[input %s]", b.name))
	defer func() { ctx.popPathElement(ctxLvl) }()
	t := &InputObjectType{
		named:         b.named,
		schemaElement: b.toSchemaElement(),
		decode:        b.decode,
		listCreator:   b.listCreator,
	}
	b.builder.resolvedTypes[b.name] = t
	fieldsByName := make(map[string]*InputObjectFieldDescriptor)
	for _, f := range b.fields {
		ctxLvl := ctx.pushPathElement(fmt.Sprintf("[field %s]", f.name))
		if _, ok := fieldsByName[b.name]; ok {
			return ctx.error("Duplicate field definition")
		}
		fieldType, err := b.builder.resolveAstType(ctx, f.typ)
		if err != nil {
			return err
		}

		decoder, decoderErr := inputObjectElementDecoder(fieldType.(InputableType))
		if decoderErr != nil {
			return ctx.error("%v", decoderErr)
		}

		fd := &InputObjectFieldDescriptor{
			f.named,
			f.toSchemaElement(),
			fieldType,
			f.defaultValue,
			decoder,
		}
		fieldsByName[f.name] = fd
		ctx.popPathElement(ctxLvl)
	}

	t.fields = fieldsByName
	return nil
}

func isValidArgumentType(typ Type) bool {
	switch t := typ.(type) {
	case WrappedType:
		return isValidArgumentType(t.Unwrap())
	case InputableType:
		return true
	}

	return false
}

// AddEnumType adds an enum type to the schema being built
func (b *Builder) AddEnumType(name string, encode EncodeEnum, decode DecodeEnum, listCreator InputListCreator) *EnumTypeBuilder {
	sb := &EnumTypeBuilder{
		builder:     b,
		named:       named{name},
		encode:      encode,
		decode:      decode,
		listCreator: listCreator,
	}

	b.addTypeBuilder(sb)
	return sb
}

// AddValue adds an allowed value to the enum being built
func (b *EnumTypeBuilder) AddValue(val string) *EnumValueBuilder {
	v := &EnumValueBuilder{
		named: named{val},
	}
	b.values = append(b.values, v)
	return v
}

func (b *EnumTypeBuilder) registerType(ctx *buildContext) buildError {
	ctxLvl := ctx.pushPathElement(fmt.Sprintf("[enum %s]", b.name))
	defer func() { ctx.popPathElement(ctxLvl) }()

	t := &EnumType{
		named:         b.named,
		schemaElement: b.toSchemaElement(),
		encode:        b.encode,
		decode:        b.decode,
		listCreator:   b.listCreator,
	}
	b.builder.resolvedTypes[b.name] = t

	values := make(map[LiteralString]*enumValueDescriptor)
	for _, e := range b.values {
		if _, ok := values[LiteralString(e.name)]; ok {
			return ctx.error("Duplicate enum value %s", e.name)
		}
		values[LiteralString(e.name)] = &enumValueDescriptor{
			named:         e.named,
			schemaElement: e.toSchemaElement(),
		}
	}
	t.values = values
	return nil
}

// AddDirectiveDefinition adds a directive definition to the schema
func (b *Builder) AddDirectiveDefinition(name string, locations ...DirectiveLocation) *DirectiveDefinitionBuilder {
	db := &DirectiveDefinitionBuilder{
		named:     named{name},
		locations: locations,
	}
	b.directives = append(b.directives, db)
	return db
}

// AddArgument adds an argument to the directive
func (b *DirectiveDefinitionBuilder) AddArgument(name string, typ ast.Type, defaultValue ast.Value) *InputValueDefinitionBuilder {
	ab := &InputValueDefinitionBuilder{
		named:        named{name},
		typ:          typ,
		defaultValue: defaultValue,
	}
	b.args = append(b.args, ab)
	return ab
}

// SetDescription sets the description of this directive
func (b *DirectiveDefinitionBuilder) SetDescription(desc string) {
	b.description = desc
}

// AddArgument adds an argument to the directive
func (b *DirectiveBuilder) AddArgument(name string, value ast.Value) {
	b.arguments = append(b.arguments, &DirectiveArgument{
		named: named{name},
		value: value,
	})
}
