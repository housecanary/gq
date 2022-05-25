package starwars

import (
	"fmt"

	"github.com/housecanary/gq/schema"
	"github.com/housecanary/gq/schema/ts"
)

func NewSchemaBuilder() (*schema.Builder, error) {
	fmt.Println("Using reflective schema")
	return ts.NewSchemaBuilder(modType)
}
