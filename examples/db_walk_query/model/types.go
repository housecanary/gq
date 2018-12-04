package model

import (
	ss "github.com/housecanary/gq/schema/structschema"
	"github.com/housecanary/gq/types"
)

type User struct {
	ss.Meta `@dbTable(name:"user")`
	ID      types.ID     `gq:"id: ID @dbColumn"`
	Name    types.String `gq:"@dbColumn"`
	Posts   []*Post      `gq:"@dbJoin(from: \"id\", to: \"author_id\")"`
	Likes   []*Like      `gq:"@dbJoin(from: \"id\", to: \"user_id\")"`
}

func (u *User) PrepareScan(fieldNames []string) []interface{} {
	var targets []interface{}
	for _, fieldName := range fieldNames {
		switch fieldName {
		case "id":
			targets = append(targets, &u.ID)
		case "name":
			targets = append(targets, &u.Name)
		}
	}
	return targets
}

func (u *User) SetJoinedModels(fieldName string, values []interface{}) {
	switch fieldName {
	case "posts":
		u.Posts = make([]*Post, len(values))
		for i, post := range values {
			u.Posts[i] = post.(*Post)
		}
	case "likes":
		u.Likes = make([]*Like, len(values))
		for i, like := range values {
			u.Likes[i] = like.(*Like)
		}
	}
}

type Post struct {
	ss.Meta  `@dbTable(name:"post")`
	ID       types.ID     `gq:"id: ID @dbColumn"`
	AuthorID types.ID     `gq:"authorId: ID @dbColumn(name: \"author_id\")"`
	Author   *User        `gq:"@dbJoin(from: \"author_id\", to: \"id\")"`
	Title    types.String `gq:"@dbColumn"`
	Body     types.String `gq:"@dbColumn"`
	Likes    []*Like      `gq:"@dbJoin(from: \"id\", to: \"post_id\")"`
}

func (p *Post) PrepareScan(fieldNames []string) []interface{} {
	var targets []interface{}
	for _, fieldName := range fieldNames {
		switch fieldName {
		case "id":
			targets = append(targets, &p.ID)
		case "authorId":
			targets = append(targets, &p.AuthorID)
		case "title":
			targets = append(targets, &p.Title)
		case "body":
			targets = append(targets, &p.Body)
		}
	}
	return targets
}

func (p *Post) SetJoinedModels(fieldName string, values []interface{}) {
	switch fieldName {
	case "author":
		if len(values) == 0 {
			p.Author = nil
		} else {
			p.Author = values[0].(*User)
		}
	case "likes":
		p.Likes = make([]*Like, len(values))
		for i, like := range values {
			p.Likes[i] = like.(*Like)
		}
	}
}

type Like struct {
	ss.Meta `@dbTable(name:"like")`
	ID      types.ID `gq:"id: ID @dbColumn"`
	UserID  types.ID `gq:"userId: ID @dbColumn(name: \"user_id\")"`
	User    *User    `gq:"@dbJoin(from: \"user_id\", to: \"id\")"`
	PostID  types.ID `gq:"postId: ID @dbColumn(name: \"post_id\")"`
	Post    *Post    `gq:"@dbJoin(from: \"post_id\", to: \"id\")"`
}

func (l *Like) PrepareScan(fieldNames []string) []interface{} {
	var targets []interface{}
	for _, fieldName := range fieldNames {
		switch fieldName {
		case "id":
			targets = append(targets, &l.ID)
		case "userId":
			targets = append(targets, &l.UserID)
		case "postId":
			targets = append(targets, &l.UserID)
		}
	}
	return targets
}

func (l *Like) SetJoinedModels(fieldName string, values []interface{}) {
	switch fieldName {
	case "user":
		if len(values) == 0 {
			l.User = nil
		} else {
			l.User = values[0].(*User)
		}
	case "post":
		if len(values) == 0 {
			l.Post = nil
		} else {
			l.Post = values[0].(*Post)
		}
	}
}
