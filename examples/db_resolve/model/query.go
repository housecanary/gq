package model

import (
	"github.com/housecanary/gq/schema"
	ss "github.com/housecanary/gq/schema/structschema"
	"github.com/housecanary/gq/types"

	"github.com/housecanary/gq/examples/db_resolve/loader"
)

type Query struct {
	ss.Meta `{
		users(from: ID, count: Int): [User]
		posts(from: ID, count: Int): [Post]
		user(id: ID): User
		post(id: ID): Post
	}`
}

func (q *Query) ResolveUsers(rc schema.ResolverContext, l *loader.DBLoader, from types.ID, count types.Int) func() ([]*User, error) {
	c := l.FetchUsers(from, count, collectFields(rc))
	return usersMapper(c)
}

func (q *Query) ResolvePosts(rc schema.ResolverContext, l *loader.DBLoader, from types.ID, count types.Int) func() ([]*Post, error) {
	c := l.FetchPosts(from, count, collectFields(rc))
	return postsMapper(c)
}

func (q *Query) ResolveUser(rc schema.ResolverContext, l *loader.DBLoader, id types.ID) func() (*User, error) {
	c := l.FetchUserByID(id, collectFields(rc))
	return userMapper(c)
}

func (q *Query) ResolvePost(rc schema.ResolverContext, l *loader.DBLoader, id types.ID) func() (*Post, error) {
	c := l.FetchPostByID(id, collectFields(rc))
	return postMapper(c)
}
