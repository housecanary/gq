package starwars

import (
	"context"
	"testing"

	"github.com/housecanary/gq/query"
)

func BenchmarkQuery(b *testing.B) {
	builder, err := NewSchemaBuilder()
	if err != nil {
		panic(err)
	}
	schema := builder.MustBuild("Query")

	pq, _ := query.PrepareQuery(`
	{
		droid(lookup: {
			id: "d1"
		}) {
			id
			name
			primaryFunction
			secretBackstory
		}
		
		random {
			__typename
			... on Human {
				homePlanet
				id
				name
				secretBackstory
			}
		}
		
		hero(episode: NEWHOPE) {
			id
			name
		}
	}`, "", schema)

	for i := 0; i < b.N; i++ {
		pq.Execute(context.Background(), &Query{}, nil, nil)
	}
}
