package model

import (
	"github.com/housecanary/gq/schema"
	ss "github.com/housecanary/gq/schema/structschema"
)

func NewSchemaBuilder() (*schema.Builder, error) {
	builder := &ss.Builder{Types: []interface{}{&Query{}}}
	return builder.SchemaBuilder()
}
