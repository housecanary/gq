package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	// SQLite driver
	_ "github.com/mattn/go-sqlite3"

	model "github.com/housecanary/gq/examples/db_walk_query/model"
	gqserver "github.com/housecanary/gq/server"
)

func main() {
	db, err := sql.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}

	populateDB(db)

	builder, err := model.NewSchemaBuilder()
	if err != nil {
		panic(err)
	}
	schema := builder.MustBuild("Query")

	handler := gqserver.NewGraphQLHandler(schema, &gqserver.GraphQLHandlerConfig{
		RootObject: &model.Query{
			DB: db,
		},
	})

	http.Handle("/graphql", handler)

	log.Fatal(http.ListenAndServe(":3000", nil))
}

func populateDB(db *sql.DB) {
	mustExec(db.Exec(`
		CREATE TABLE user (
			id TEXT PRIMARY KEY,
			name TEXT
		)
	`))

	for i := 0; i < 100; i++ {
		mustExec(db.Exec(`
			INSERT INTO user(id, name) VALUES (?, ?)
		`, fmt.Sprintf("%v", i), fmt.Sprintf("User %v", i)))
	}

	mustExec(db.Exec(`
		CREATE TABLE post (
			id TEXT PRIMARY KEY,
			author_id TEXT,
			title TEXT,
			body TEXT
		)
	`))

	for i := 0; i < 1000; i++ {
		mustExec(db.Exec(`
			INSERT INTO post(id, author_id, title, body) VALUES (?, ?, ?, ?)
		`, fmt.Sprintf("%v", i), fmt.Sprintf("%v", i%100), fmt.Sprintf("Post %v", i), fmt.Sprintf("Lorem ipsum %v", i)))
	}

	mustExec(db.Exec(`
		CREATE TABLE like (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			post_id TEXT
		)
	`))
}

func mustExec(_ sql.Result, err error) {
	if err != nil {
		panic(err)
	}
}
