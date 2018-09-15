package starwars

import (
	"fmt"

	"github.com/housecanary/gq/schema"
	ss "github.com/housecanary/gq/schema/structschema"
)

func NewSchemaBuilder() (*schema.Builder, error) {
	fmt.Println("Using reflective schema")
	builder := &ss.Builder{Types: []interface{}{&Query{}}}
	return builder.SchemaBuilder()
}
