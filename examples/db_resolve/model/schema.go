package model

import (
	"context"

	"github.com/housecanary/gq/schema"
	ss "github.com/housecanary/gq/schema/structschema"
)

func NewSchemaBuilder(createDBLoader func(context.Context) interface{}) (*schema.Builder, error) {
	builder := &ss.Builder{Types: []interface{}{&Query{}}}
	builder.RegisterArgProvider("*loader.DBLoader", createDBLoader)
	return builder.SchemaBuilder()
}
