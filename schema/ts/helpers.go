package ts

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/codemodus/kace"
	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser"
	"github.com/housecanary/gq/schema"
)

type reflectionInputListCreator struct {
	typ reflect.Type
}

func (r reflectionInputListCreator) NewList(size int, get func(i int) (interface{}, error)) (interface{}, error) {
	lst := reflect.MakeSlice(reflect.SliceOf(r.typ), size, size)
	for i := 0; i < size; i++ {
		v, err := get(i)
		if err != nil {
			return nil, err
		}
		lst.Index(i).Set(reflect.ValueOf(v))
	}
	return lst.Interface(), nil
}

func (r reflectionInputListCreator) Creator() schema.InputListCreator {
	return reflectionInputListCreator{reflect.SliceOf(r.typ)}
}

func typeOf[T any]() reflect.Type {
	var empty T
	return reflect.TypeOf(&empty).Elem()
}

type structParseResult interface {
	ast.InputValueDefinition | ast.FieldDefinition
}

func parseStructField[P func(string) (*R, parser.ParseError), R structParseResult](c *buildContext, f reflect.StructField, parse P) (*R, bool, error) {
	if f.Anonymous || !f.IsExported() {
		return nil, false, nil
	}
	var parseResult R
	var description string
	if gq, ok := f.Tag.Lookup("gq"); ok {
		parts := strings.SplitN(gq, ";", 2)
		if parts[0] == "-" {
			return nil, false, nil
		}
		if parts[0] != "" {
			pr, err := parse(parts[0])
			if err != nil {
				return pr, false, err
			}
			parseResult = *pr
		}
		if len(parts) == 2 {
			description = parts[1]
		}
	}

	var pName *string
	var pDesc *string
	var pTyp *ast.Type
	tryPtrType := false
	allowInputObject := true
	switch pr := (interface{})(&parseResult).(type) {
	case *ast.InputValueDefinition:
		pName = &pr.Name
		pDesc = &pr.Description
		pTyp = &pr.Type
	case *ast.FieldDefinition:
		pName = &pr.Name
		pDesc = &pr.Description
		pTyp = &pr.Type
		tryPtrType = true
		allowInputObject = false
	}

	if *pName == "" {
		*pName = kace.Camel(f.Name)
	}

	if *pDesc == "" {
		*pDesc = description
	}

	var typeWasInferred = false

	// If tryPtrType is set check and see if T is a struct and *T is in the
	// type registry as an object type. If so, use that to calculate the
	// field type instead.
	fTyp := f.Type
	if tryPtrType && fTyp.Kind() == reflect.Struct {
		if gqlType, ok := c.goTypeToSchemaType[reflect.PointerTo(fTyp)]; ok && gqlType.kind == kindObject {
			fTyp = reflect.PointerTo(fTyp)
		}
	}
	if *pTyp == nil {
		astTypeFromGo, err := c.astTypeForGoType(fTyp)
		if err != nil {
			return nil, false, err
		}
		typeWasInferred = true
		*pTyp = astTypeFromGo
	} else {
		if err := c.checkTypeCompatible(fTyp, *pTyp); err != nil {
			return nil, false, err
		}
	}

	if !allowInputObject {
		gqlType := c.goTypeToSchemaType[fTyp]
		if gqlType.kind == kindInputObject {
			return nil, false, fmt.Errorf("input object type %s not allowed as a GQL object field", gqlType.astType.Signature())
		}
	}

	return &parseResult, typeWasInferred, nil
}

func setSchemaElementProps(e schema.BuilderSchemaElement, desc string, directives ast.Directives) {
	e.SetDescription(desc)
	for _, d := range directives {
		db := e.AddDirective(d.Name)
		for _, a := range d.Arguments {
			db.AddArgument(a.Name, a.Value)
		}

	}
}

func typeDesc(t reflect.Type) string {
	orig := t
	prefix := ""
	for {
		if t.Kind() == reflect.Pointer {
			prefix += "*"
			t = t.Elem()
			continue
		}

		if t.Kind() == reflect.Slice {
			prefix += "[]"
			t = t.Elem()
			continue
		}

		break
	}

	pkgName := t.PkgPath()
	name := t.Name()
	if pkgName == "" && name == "" {
		return orig.String()
	}

	return fmt.Sprintf("%s{%s}{%s}", prefix, pkgName, name)
}

func parseTypeDef[T any, OT any](kind gqlTypeKind, gqlText, namePrefix string, outTypeDef **ast.BasicTypeDefinition) (*gqlTypeInfo, reflect.Type, error) {
	typ := typeOf[T]()

	typeDef, err := parser.ParseTSTypeDefinition(gqlText)
	if err != nil {
		return nil, nil, err
	}

	name := typeDef.Name
	if name == "" {
		name = kace.Pascal(typ.Name())
	}
	if name == "" {
		return nil, nil, fmt.Errorf("name cannot be inferred from type %v, provide a name in the metadata", typeDesc(typ))
	}
	name = namePrefix + name
	typeDef.Name = name
	*outTypeDef = typeDef

	return &gqlTypeInfo{&ast.SimpleType{Name: name}, kind}, typeOf[OT](), nil
}
