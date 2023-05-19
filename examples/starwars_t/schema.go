package starwars

import (
	"github.com/housecanary/gq/schema"
	"github.com/housecanary/gq/schema/ts"
)

func NewSchemaBuilder() (*schema.Builder, error) {
	tr, err := ts.NewTypeRegistry(ts.WithModule(starwarsModule))
	if err != nil {
		return nil, err
	}
	return tr.SchemaBuilder(), nil
}
