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

func parseStructField[P func(string) (*R, parser.ParseError), R any](c *buildContext, f reflect.StructField, parse P) (*R, bool, error) {
	if f.Anonymous {
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
	switch pr := (interface{})(&parseResult).(type) {
	case *ast.InputValueDefinition:
		pName = &pr.Name
		pDesc = &pr.Description
		pTyp = &pr.Type
	case *ast.FieldDefinition:
		pName = &pr.Name
		pDesc = &pr.Description
		pTyp = &pr.Type
	}

	if *pName == "" {
		*pName = kace.Camel(f.Name)
	}

	if *pDesc == "" {
		*pDesc = description
	}

	var typeWasInferred = false
	if *pTyp == nil {
		astTypeFromGo, err := c.astTypeForGoType(f.Type)
		if err != nil {
			return nil, false, err
		}
		typeWasInferred = true
		*pTyp = astTypeFromGo
	} else {
		if err := c.checkTypeCompatible(f.Type, *pTyp); err != nil {
			return nil, false, err
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
