package ts_test

import (
	"fmt"
	"time"

	"github.com/housecanary/gq/schema"
	"github.com/housecanary/gq/schema/ts"
)

var scalarModType = ts.NewModule()

// To create an scalar type, first define the type that will hold your value.
// The type must implement schema.ScalarMarshaler for the value and schema.ScalarUnmarshaler
// for a pointer to the value

type DateTime time.Time

func (v DateTime) ToLiteralValue() (schema.LiteralValue, error) {
	formatted := time.Time(v).Format(time.RFC3339Nano)
	return schema.LiteralString(formatted), nil
}

func (v *DateTime) FromLiteralValue(l schema.LiteralValue) error {
	if l == nil {
		*v = DateTime{}
		return nil
	}
	switch c := l.(type) {
	case schema.LiteralString:
		parsed, err := time.Parse(time.RFC3339Nano, string(c))
		if err != nil {
			return fmt.Errorf("invalid datetime %s: %w", c, err)
		}
		*v = DateTime(parsed)
		return nil
	default:
		return fmt.Errorf("Literal value %v is not a string", l)
	}
}

// Next, construct the GQL type using the ts.NewScalarType function
var dateType = ts.NewScalarType[DateTime](scalarModType, `"An ISO format datetime."`)

func ExampleNewScalarType() {
	// Once the scalar type is registered, it can be used in arguments, as a struct field, etc
	// just like any of the built in types
}
