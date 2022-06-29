package structschema

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/codemodus/kace"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/internal/pkg/parser"
)

type StructFieldMetadata struct {
	Name         string
	Description  string
	Directives   ast.Directives
	ReflectField reflect.StructField
}

// GetStructFieldMetadata loads GQL information about the fields of a struct from
// the annotations on the struct.
//
// Information such as the resolved GQL type of the field is not available as that
// would require a proper builder to resolve the types.
func GetStructFieldMetadata(typ reflect.Type) ([]*StructFieldMetadata, error) {
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("invalid type: expected a struct type, not a %v", typ.Kind())
	}

	// Find and parse the meta field that contains partial GraphQL definition
	// of this type
	td := &ast.ObjectTypeDefinition{}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Type == schemaMetaType {
			gqlTypeDef, err := parser.ParsePartialObjectTypeDefinition(string(f.Tag))
			if err != nil {
				return nil, fmt.Errorf("Cannot parse GQL metadata for object %s: %v", typ.Name(), err)
			}
			td = gqlTypeDef
			break
		}
	}

	// Merge in field definition data defined in GQL with data
	// discovered by reflecting over the fields of the struct
	fieldsByName := make(map[string]*fieldMeta)
	for _, fd := range td.FieldsDefinition {
		fieldsByName[fd.Name] = &fieldMeta{
			Name:     fd.Name,
			GqlField: fd,
		}
	}

	var result []*StructFieldMetadata
	for _, f := range flatFields(typ) {
		fieldDef := &ast.FieldDefinition{}

		if tag, ok := f.Tag.Lookup("gq"); ok {
			tag := strings.TrimSpace(tag)
			parts := strings.SplitN(tag, ";", 2)
			gql := strings.TrimSpace(parts[0])
			doc := ""
			if len(parts) > 1 {
				doc = parts[1]
			}

			if len(gql) > 0 {
				if strings.HasPrefix(gql, "-") {
					continue
				} else {
					gqlFieldDef, err := parser.ParsePartialFieldDefinition(gql)
					if err != nil {
						return nil, fmt.Errorf("Cannot parse GQL metadata for object %s, field %s: %v", typ.Name(), f.Name, err)
					}
					fieldDef = gqlFieldDef
				}
			}

			if doc != "" {
				fieldDef.Description = doc
			}
		}

		if fieldDef.Name == "" {
			fieldDef.Name = kace.Camel(f.Name)
		}

		if existingMeta, ok := fieldsByName[fieldDef.Name]; ok {
			mergeFieldDef(fieldDef, existingMeta.GqlField)
		}

		result = append(result, &StructFieldMetadata{
			Name:         fieldDef.Name,
			Description:  fieldDef.Description,
			Directives:   fieldDef.Directives,
			ReflectField: f,
		})
	}

	return result, nil
}
