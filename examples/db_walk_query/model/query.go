package model

import (
	"database/sql"
	"fmt"

	"github.com/housecanary/gq/schema"
	ss "github.com/housecanary/gq/schema/structschema"
	"github.com/housecanary/gq/types"

	loaders "github.com/housecanary/gq/examples/db_walk_query/loader"
)

type Query struct {
	ss.Meta `{
		users(from: ID, count: Int): [User]
		posts(from: ID, count: Int, authorId: ID): [Post]
		user(id: ID): User
		post(id: ID): Post
	}`

	DB *sql.DB `gq:"-"`
}

func createModel(name string) loaders.DBModel {
	switch name {
	case "User":
		return &User{}
	case "Post":
		return &Post{}
	case "Like":
		return &Like{}
	}
	panic("Requested unknown model " + name)
}

func (q *Query) ResolveUsers(ctx schema.ResolverContext, from types.ID, count types.Int) func() ([]*User, error) {
	loader := loaders.NewDBLoader(q.DB, createModel, "User", "user", "id")
	ctx.WalkChildSelections(loader.WalkQuery)
	if !count.Nil() {
		var fromVal interface{}
		if !from.Nil() {
			fromVal = from.String()
		}
		loader.Page(fromVal, int(count.Int32()))
	}
	return func() ([]*User, error) {
		results, err := loader.Load()

		if err != nil {
			return nil, err
		}

		n := len(results)
		if !count.Nil() && int(count.Int32()) < n {
			n = int(count.Int32())
		}

		r := make([]*User, n)
		for i, u := range results {
			if i >= n {
				break
			}
			r[i] = u.(*User)
		}

		return r, nil
	}

}

func (q *Query) ResolvePosts(ctx schema.ResolverContext, from types.ID, count types.Int, authorID types.ID) func() ([]*Post, error) {
	loader := loaders.NewDBLoader(q.DB, createModel, "Post", "post", "id")
	ctx.WalkChildSelections(loader.WalkQuery)
	if !count.Nil() {
		var fromVal interface{}
		if !from.Nil() {
			fromVal = from.String()
		}
		loader.Page(fromVal, int(count.Int32()))
	}

	if !authorID.Nil() {
		loader.Where("author_id", "=", authorID.String())
	}

	return func() ([]*Post, error) {
		results, err := loader.Load()

		if err != nil {
			return nil, err
		}

		n := len(results)
		if !count.Nil() && int(count.Int32()) < n {
			n = int(count.Int32())
		}

		r := make([]*Post, n)
		for i, u := range results {
			if i >= n {
				break
			}
			r[i] = u.(*Post)
		}

		return r, nil
	}

}

func (q *Query) ResolveUser(ctx schema.ResolverContext, id types.ID) func() (*User, error) {
	loader := loaders.NewDBLoader(q.DB, createModel, "User", "user", "id")
	ctx.WalkChildSelections(loader.WalkQuery)
	loader.Where("id", "=", id.String())
	return func() (*User, error) {
		results, err := loader.Load()

		if err != nil {
			return nil, err
		}

		n := len(results)
		if n == 0 {
			return nil, nil
		} else if n > 1 {
			return nil, fmt.Errorf("Load by id %v matched more than 1 row", id.String())
		}

		return results[0].(*User), nil
	}

}

func (q *Query) ResolvePost(ctx schema.ResolverContext, id types.ID) func() (*Post, error) {
	loader := loaders.NewDBLoader(q.DB, createModel, "Post", "post", "id")
	ctx.WalkChildSelections(loader.WalkQuery)
	loader.Where("id", "=", id.String())
	return func() (*Post, error) {
		results, err := loader.Load()

		if err != nil {
			return nil, err
		}

		n := len(results)
		if n == 0 {
			return nil, nil
		} else if n > 1 {
			return nil, fmt.Errorf("Load by id %v matched more than 1 row", id.String())
		}

		return results[0].(*Post), nil
	}

}
