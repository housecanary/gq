package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/jmoiron/sqlx"

	// SQLite driver
	_ "github.com/mattn/go-sqlite3"

	"github.com/housecanary/gq/examples/db_resolve/loader"
	"github.com/housecanary/gq/examples/db_resolve/model"
	"github.com/housecanary/gq/query"
	"github.com/housecanary/gq/server"
)

type dbLoaderKeyType struct{}

var dbLoaderKey dbLoaderKeyType

func main() {
	db, err := sqlx.Open("sqlite3", "file::memory:?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}

	populateDB(db)

	builder, err := model.NewSchemaBuilder(func(c context.Context) interface{} {
		return c.Value(dbLoaderKey)
	})

	if err != nil {
		panic(err)
	}

	executeQuery := func(q *query.PreparedQuery, req *http.Request, vars query.Variables, responseHeaders http.Header) []byte {
		l := loader.NewDBLoader(db)
		ctx := context.WithValue(req.Context(), dbLoaderKey, l)
		ql := &queryListener{
			ctx:    ctx,
			loader: l,
		}
		return q.Execute(ctx, &model.Query{}, vars, ql)
	}

	schema := builder.MustBuild("Query")

	handler := server.NewGraphQLHandler(schema, &server.GraphQLHandlerConfig{
		RootObject:    &model.Query{},
		QueryExecutor: executeQuery,
	})

	http.Handle("/graphql", handler)

	log.Fatal(http.ListenAndServe(":3000", nil))
}

type queryListener struct {
	query.BaseExecutionListener
	ctx    context.Context
	loader *loader.DBLoader
}

func (l *queryListener) NotifyIdle() {
	l.loader.ExecuteBatches()
}

func populateDB(db *sqlx.DB) {
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

	for i := 0; i < 1000; i++ {
		userID := rand.Intn(100)
		postID := rand.Intn(1000)
		mustExec(db.Exec(`
			INSERT INTO like(id, user_id, post_id) VALUES (?, ?, ?)
		`, fmt.Sprintf("%v", i), fmt.Sprintf("%v", userID), fmt.Sprintf("%v", postID)))
	}
}

func mustExec(_ sql.Result, err error) {
	if err != nil {
		panic(err)
	}
}
