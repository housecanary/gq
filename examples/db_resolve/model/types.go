package model

import (
	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
	ss "github.com/housecanary/gq/schema/structschema"
	"github.com/housecanary/gq/types"

	"github.com/housecanary/gq/examples/db_resolve/loader"
)

type User struct {
	ss.Meta `{
		posts: [Post]
		likes: [Like]
	}`
	UserData
}

type UserData struct {
	ID   types.ID `gq:"id"`
	Name types.String
}

func (u *User) ResolvePosts(rc schema.ResolverContext, l *loader.DBLoader) func() ([]*Post, error) {
	c := l.FetchPostsByUserID(u.ID, collectFields(rc))
	return postsMapper(c)
}

func (u *User) ResolveLikes(rc schema.ResolverContext, l *loader.DBLoader) func() ([]*Like, error) {
	c := l.FetchLikesByUserID(u.ID, collectFields(rc))
	return likesMapper(c)
}

type Post struct {
	ss.Meta `{
		author: User
		likes: [Like]
	}`
	PostData
}

type PostData struct {
	ID       types.ID `gq:"id"`
	AuthorID types.ID `gq:"authorId"`
	Title    types.String
	Body     types.String
}

func (p *Post) ResolveAuthor(rc schema.ResolverContext, l *loader.DBLoader) func() (*User, error) {
	c := l.FetchUserByID(p.AuthorID, collectFields(rc))
	return userMapper(c)
}

func (p *Post) ResolveLikes(rc schema.ResolverContext, l *loader.DBLoader) func() ([]*Like, error) {
	c := l.FetchLikesByPostID(p.ID, collectFields(rc))
	return likesMapper(c)
}

type Like struct {
	ss.Meta `{
		user: User
		post: Post
	}`
	LikeData
}

type LikeData struct {
	ID     types.ID `gq:"id"`
	UserID types.ID `gq:"userId"`
	PostID types.ID `gq:"postId"`
}

func (lk *Like) ResolveUser(rc schema.ResolverContext, l *loader.DBLoader) func() (*User, error) {
	c := l.FetchUserByID(lk.UserID, collectFields(rc))
	return userMapper(c)
}

func (lk *Like) ResolvePost(rc schema.ResolverContext, l *loader.DBLoader) func() (*Post, error) {
	c := l.FetchPostByID(lk.PostID, collectFields(rc))
	return postMapper(c)
}

func collectFields(rc schema.ResolverContext) []string {
	var fields []string
	rc.WalkChildSelections(func(selection *ast.Field, field *schema.FieldDescriptor, walker schema.ChildWalker) bool {
		fields = append(fields, field.Name())
		return false
	})
	return fields
}

func userMapper(c <-chan loader.UserFetchResult) func() (*User, error) {
	return func() (*User, error) {
		r := <-c
		if r.Error != nil || r.User == nil {
			return nil, r.Error
		}

		return &User{
			UserData: UserData(*r.User),
		}, nil
	}
}

func usersMapper(c <-chan loader.UsersFetchResult) func() ([]*User, error) {
	return func() ([]*User, error) {
		r := <-c
		if r.Error != nil {
			return nil, r.Error
		}

		users := make([]*User, len(r.Users))
		for i, user := range r.Users {
			users[i] = &User{
				UserData: UserData(*user),
			}
		}
		return users, nil
	}
}

func postMapper(c <-chan loader.PostFetchResult) func() (*Post, error) {
	return func() (*Post, error) {
		r := <-c
		if r.Error != nil || r.Post == nil {
			return nil, r.Error
		}

		return &Post{
			PostData: PostData(*r.Post),
		}, nil
	}
}

func postsMapper(c <-chan loader.PostsFetchResult) func() ([]*Post, error) {
	return func() ([]*Post, error) {
		r := <-c
		if r.Error != nil {
			return nil, r.Error
		}

		posts := make([]*Post, len(r.Posts))
		for i, post := range r.Posts {
			posts[i] = &Post{
				PostData: PostData(*post),
			}
		}
		return posts, nil
	}
}

func likesMapper(c <-chan loader.LikesFetchResult) func() ([]*Like, error) {
	return func() ([]*Like, error) {
		r := <-c
		if r.Error != nil {
			return nil, r.Error
		}

		likes := make([]*Like, len(r.Likes))
		for i, like := range r.Likes {
			likes[i] = &Like{
				LikeData: LikeData(*like),
			}
		}
		return likes, nil
	}
}
