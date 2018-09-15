package main

import (
	"log"
	"net/http"

	starwars "github.com/housecanary/gq/examples/starwars"
	gqserver "github.com/housecanary/gq/server"
)

func main() {
	builder, err := starwars.NewSchemaBuilder()
	if err != nil {
		panic(err)
	}
	schema := builder.MustBuild("Query")

	handler := gqserver.NewGraphQLHandler(schema, &gqserver.GraphQLHandlerConfig{
		RootObject: &starwars.Query{},
	})

	http.Handle("/graphql", handler)

	log.Fatal(http.ListenAndServe(":3000", nil))
}
