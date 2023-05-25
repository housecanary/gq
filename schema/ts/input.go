package ts

import (
	"fmt"
	"reflect"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

type inputConverter func(value schema.LiteralValue, dest reflect.Value) error

type inputConverterProvider interface {
	makeInputConverter(c *buildContext) inputConverter
}

func makeInputConverterForType(c *buildContext, gqlType ast.Type, goType reflect.Type) inputConverter {
	switch gqlType := gqlType.(type) {
	case *ast.ListType:
		elementConverter := makeInputConverterForType(c, gqlType.Of, goType.Elem())
		return func(value schema.LiteralValue, dest reflect.Value) error {
			if value == nil {
				dest.Set(reflect.Zero(goType))
			}

			var inputAry schema.LiteralArray
			if la, ok := value.(schema.LiteralArray); ok {
				inputAry = la
			} else {
				inputAry = schema.LiteralArray{value}
			}

			slice := reflect.MakeSlice(goType, len(inputAry), len(inputAry))
			for i := 0; i < len(inputAry); i++ {
				target := slice.Index(i)
				if err := elementConverter(inputAry[i], target); err != nil {
					return fmt.Errorf("cannot convert input list element %d: %w", i, err)
				}
			}
			dest.Set(slice)
			return nil
		}
	case *ast.NotNilType:
		elementConverter := makeInputConverterForType(c, gqlType.Of, goType)
		return func(value schema.LiteralValue, dest reflect.Value) error {
			if value == nil {
				return fmt.Errorf("value is required, but was not provided")
			}
			return elementConverter(value, dest)
		}
	default:
		builder := c.goTypeToBuilder[goType]
		if icp, ok := builder.(inputConverterProvider); ok {
			return icp.makeInputConverter(c)
		}

		// field is not a valid input type, this will be reported elsewhere
		return nil
	}
}
