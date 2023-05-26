package ts

import (
	"context"
	"fmt"
	"testing"

	"github.com/housecanary/gq/query"
	"github.com/housecanary/gq/types"
)

func BenchmarkFieldQuery(b *testing.B) {
	mod := NewModule()
	type Object1 struct {
		Field types.String
	}
	NewObjectType[Object1](mod, ``)
	tr, err := NewTypeRegistry(WithModule(mod))
	if err != nil {
		b.Fatal(err)
	}

	schema := tr.MustBuildSchema("Object1")
	queryText := "{"
	for i := 0; i < 50; i++ {
		queryText += fmt.Sprintf("field%d: field\n", i)
	}
	queryText += "}"
	pq, err := query.PrepareQuery(queryText, "", schema)
	if err != nil {
		b.Fatal(err)
	}

	root := &Object1{
		Field: types.NewString("test"),
	}
	for i := 0; i < b.N; i++ {
		pq.Execute(context.Background(), root, nil, nil)
	}

}
