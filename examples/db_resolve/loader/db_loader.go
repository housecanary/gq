package loader

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/housecanary/gq/types"
)

// A DBLoader can fetch a graph of objects from the database
type DBLoader struct {
	db      *sqlx.DB
	pending map[loadGroup][](*loadItem)
}

// NewDBLoader makes a new DBLoader
func NewDBLoader(db *sqlx.DB) *DBLoader {
	return &DBLoader{
		db:      db,
		pending: make(map[loadGroup][](*loadItem)),
	}
}

type UserFetchResult struct {
	Error error
	User  *UserColumns
}

type UsersFetchResult struct {
	Error error
	Users []*UserColumns
}

type UserColumns struct {
	ID   types.ID
	Name types.String
}

var userFieldToColumn = map[string]string{
	"name": "name",
}

type PostFetchResult struct {
	Error error
	Post  *PostColumns
}

type PostsFetchResult struct {
	Error error
	Posts []*PostColumns
}

type PostColumns struct {
	ID       types.ID
	AuthorID types.ID `db:"author_id"`
	Title    types.String
	Body     types.String
}

var postFieldToColumn = map[string]string{
	"authorId": "author_id",
	"author":   "author_id",
	"title":    "title",
	"bpdy":     "body",
}

type LikesFetchResult struct {
	Error error
	Likes []*LikeColumns
}

type LikeColumns struct {
	ID     types.ID
	UserID types.ID `db:"user_id"`
	PostID types.ID `db:"post_id"`
}

var likeFieldToColumn = map[string]string{
	"userId": "user_id",
	"user":   "user_id",
	"postId": "post_id",
	"post":   "post_id",
}

func (l *DBLoader) FetchUsers(from types.ID, count types.Int, fieldNames []string) <-chan UsersFetchResult {
	c := make(chan UsersFetchResult, 1)
	k := userByPageLoadGroup{from, count}
	var itemKey interface{}
	if !from.Nil() {
		itemKey = from.String()
	}
	l.pending[k] = append(l.pending[k], &loadItem{
		itemKey,
		fieldNames,
		func(data interface{}, err error) {
			if err != nil || data == nil {
				c <- UsersFetchResult{Error: err}
			} else {
				c <- UsersFetchResult{Users: data.([]*UserColumns)}
			}
		},
	})
	return c
}

func (l *DBLoader) FetchUserByID(userID types.ID, fieldNames []string) <-chan UserFetchResult {
	c := make(chan UserFetchResult, 1)
	k := userByIDLoadGroup{}
	l.pending[k] = append(l.pending[k], &loadItem{
		userID.String(),
		fieldNames,
		func(data interface{}, err error) {
			if err != nil || data == nil {
				c <- UserFetchResult{Error: err}
			} else {
				c <- UserFetchResult{User: data.(*UserColumns)}
			}
		},
	})
	return c
}

func (l *DBLoader) FetchPosts(from types.ID, count types.Int, fieldNames []string) <-chan PostsFetchResult {
	c := make(chan PostsFetchResult, 1)
	k := postsByPageLoadGroup{from, count}
	var itemKey interface{}
	if !from.Nil() {
		itemKey = from.String()
	}
	l.pending[k] = append(l.pending[k], &loadItem{
		itemKey,
		fieldNames,
		func(data interface{}, err error) {
			if err != nil || data == nil {
				c <- PostsFetchResult{Error: err}
			} else {
				c <- PostsFetchResult{Posts: data.([]*PostColumns)}
			}
		},
	})
	return c
}

func (l *DBLoader) FetchPostByID(postID types.ID, fieldNames []string) <-chan PostFetchResult {
	c := make(chan PostFetchResult, 1)

	k := postByIDLoadGroup{}
	l.pending[k] = append(l.pending[k], &loadItem{
		postID.String(),
		fieldNames,
		func(data interface{}, err error) {
			if err != nil || data == nil {
				c <- PostFetchResult{Error: err}
			} else {
				c <- PostFetchResult{Post: data.(*PostColumns)}
			}
		},
	})
	return c
}

func (l *DBLoader) FetchLikesByPostID(postID types.ID, fieldNames []string) <-chan LikesFetchResult {
	c := make(chan LikesFetchResult, 1)
	k := likesByPostIDLoadGroup{}
	l.pending[k] = append(l.pending[k], &loadItem{
		postID.String(),
		fieldNames,
		func(data interface{}, err error) {
			if err != nil || data == nil {
				c <- LikesFetchResult{Error: err}
			} else {
				c <- LikesFetchResult{Likes: data.([]*LikeColumns)}
			}
		},
	})
	return c
}

func (l *DBLoader) FetchLikesByUserID(userID types.ID, fieldNames []string) <-chan LikesFetchResult {
	c := make(chan LikesFetchResult, 1)
	k := likesByUserIDLoadGroup{}
	l.pending[k] = append(l.pending[k], &loadItem{
		userID.String(),
		fieldNames,
		func(data interface{}, err error) {
			if err != nil || data == nil {
				c <- LikesFetchResult{Error: err}
			} else {
				c <- LikesFetchResult{Likes: data.([]*LikeColumns)}
			}
		},
	})
	return c
}

func (l *DBLoader) FetchPostsByUserID(userID types.ID, fieldNames []string) <-chan PostsFetchResult {
	c := make(chan PostsFetchResult, 1)
	k := postsByUserIDLoadGroup{}
	l.pending[k] = append(l.pending[k], &loadItem{
		userID.String(),
		fieldNames,
		func(data interface{}, err error) {
			if err != nil || data == nil {
				c <- PostsFetchResult{Error: err}
			} else {
				c <- PostsFetchResult{Posts: data.([]*PostColumns)}
			}
		},
	})
	return c
}

func (l *DBLoader) ExecuteBatches() {
	pending := l.pending
	l.pending = make(map[loadGroup][]*loadItem)
	for lg, items := range pending {
		go func(lg loadGroup, items []*loadItem) {
			fieldToCol := lg.fieldToColumnMap()
			qs := make([]string, 0, len(items))
			args := make([]interface{}, 0, len(items))
			colNames := make(map[string]bool)
			colNames["id"] = true
			dedupKeys := make(map[interface{}]bool)
			for _, item := range items {
				for _, fn := range item.fieldNames {
					if colName, ok := fieldToCol[fn]; ok {
						colNames[colName] = true
					}
				}

				if item.key != nil && !dedupKeys[item.key] {
					dedupKeys[item.key] = true
					qs = append(qs, "?")
					args = append(args, item.key)
				}
			}

			if ec, ok := lg.(extraColumner); ok {
				for _, colName := range ec.extraColumns() {
					colNames[colName] = true
				}
			}

			cols := make([]string, 0, len(colNames))
			for k := range colNames {
				cols = append(cols, k)
			}

			var sql string
			if len(qs) > 0 {
				sql = fmt.Sprintf(lg.sqlTemplate(), strings.Join(cols, ","), strings.Join(qs, ","))
			} else {
				sql = fmt.Sprintf(lg.sqlTemplate(), strings.Join(cols, ","))
			}
			fmt.Println("Execute sql", sql, args)
			err := lg.query(l.db, sql, args, items)
			if err != nil {
				for _, item := range items {
					item.done(nil, err)
				}
			}
		}(lg, items)
	}
}

type loadGroup interface {
	sqlTemplate() string
	fieldToColumnMap() map[string]string
	query(db *sqlx.DB, sql string, args []interface{}, items []*loadItem) error
}

type extraColumner interface {
	extraColumns() []string
}

type userByPageLoadGroup struct {
	from  types.ID
	count types.Int
}

func (lg userByPageLoadGroup) sqlTemplate() string {
	var limit string
	if !lg.count.Nil() {
		limit = fmt.Sprintf(" LIMIT %v", lg.count.Int32())
	}
	if !lg.from.Nil() {
		return "SELECT %s FROM user WHERE id > %s ORDER BY id" + limit
	}
	return "SELECT %s FROM user ORDER BY id" + limit
}

func (userByPageLoadGroup) fieldToColumnMap() map[string]string {
	return userFieldToColumn
}

func (userByPageLoadGroup) query(db *sqlx.DB, sql string, args []interface{}, items []*loadItem) error {
	var result []*UserColumns
	err := db.Select(&result, sql, args...)
	if err != nil {
		return err
	}

	completeList(items, result)

	return nil
}

type userByIDLoadGroup struct{}

func (userByIDLoadGroup) sqlTemplate() string {
	return "SELECT %s FROM user WHERE id IN (%s)"
}

func (userByIDLoadGroup) fieldToColumnMap() map[string]string {
	return userFieldToColumn
}

func (userByIDLoadGroup) query(db *sqlx.DB, sql string, args []interface{}, items []*loadItem) error {
	var result []*UserColumns
	err := db.Select(&result, sql, args...)
	if err != nil {
		return err
	}

	completeByID(items, len(result), func(i int) (interface{}, interface{}) {
		return result[i].ID.String(), result[i]
	})

	return nil
}

type postsByPageLoadGroup struct {
	from  types.ID
	count types.Int
}

func (lg postsByPageLoadGroup) sqlTemplate() string {
	var limit string
	if !lg.count.Nil() {
		limit = fmt.Sprintf(" LIMIT %v", lg.count.Int32())
	}
	if !lg.from.Nil() {
		return "SELECT %s FROM post WHERE id > %s ORDER BY id" + limit
	}
	return "SELECT %s FROM post ORDER BY id" + limit
}

func (postsByPageLoadGroup) fieldToColumnMap() map[string]string {
	return postFieldToColumn
}

func (postsByPageLoadGroup) query(db *sqlx.DB, sql string, args []interface{}, items []*loadItem) error {
	var result []*PostColumns
	err := db.Select(&result, sql, args...)
	if err != nil {
		return err
	}

	completeList(items, result)

	return nil
}

type postByIDLoadGroup struct{}

func (postByIDLoadGroup) sqlTemplate() string {
	return "SELECT %s FROM post WHERE id IN (%s)"
}

func (postByIDLoadGroup) fieldToColumnMap() map[string]string {
	return postFieldToColumn
}

func (postByIDLoadGroup) query(db *sqlx.DB, sql string, args []interface{}, items []*loadItem) error {
	var result []*PostColumns
	err := db.Select(&result, sql, args...)
	if err != nil {
		return err
	}

	completeByID(items, len(result), func(i int) (interface{}, interface{}) {
		return result[i].ID.String(), result[i]
	})

	return nil
}

type likesByPostIDLoadGroup struct{}

func (likesByPostIDLoadGroup) sqlTemplate() string {
	return "SELECT %s FROM like WHERE post_id IN (%s)"
}

func (likesByPostIDLoadGroup) fieldToColumnMap() map[string]string {
	return likeFieldToColumn
}

func (likesByPostIDLoadGroup) query(db *sqlx.DB, sql string, args []interface{}, items []*loadItem) error {
	var result []*LikeColumns
	err := db.Select(&result, sql, args...)
	if err != nil {
		return err
	}

	completeByGroupID(items, len(result), func(i int) (interface{}, interface{}) {
		return result[i].PostID.String(), result[i]
	}, func(group interface{}, item interface{}) interface{} {
		if group == nil {
			group = make([]*LikeColumns, 0)
		}
		return append(group.([]*LikeColumns), item.(*LikeColumns))
	})

	return nil
}

func (likesByPostIDLoadGroup) extraColumns() []string {
	return []string{"post_id"}
}

type likesByUserIDLoadGroup struct{}

func (likesByUserIDLoadGroup) sqlTemplate() string {
	return "SELECT %s FROM like WHERE user_id IN (%s)"
}

func (likesByUserIDLoadGroup) fieldToColumnMap() map[string]string {
	return likeFieldToColumn
}

func (likesByUserIDLoadGroup) query(db *sqlx.DB, sql string, args []interface{}, items []*loadItem) error {
	var result []*LikeColumns
	err := db.Select(&result, sql, args...)
	if err != nil {
		return err
	}

	completeByGroupID(items, len(result), func(i int) (interface{}, interface{}) {
		return result[i].UserID.String(), result[i]
	}, func(group interface{}, item interface{}) interface{} {
		if group == nil {
			group = make([]*LikeColumns, 0)
		}
		return append(group.([]*LikeColumns), item.(*LikeColumns))
	})

	return nil
}

func (likesByUserIDLoadGroup) extraColumns() []string {
	return []string{"user_id"}
}

type postsByUserIDLoadGroup struct{}

func (postsByUserIDLoadGroup) sqlTemplate() string {
	return "SELECT %s FROM post WHERE author_id IN (%s)"
}

func (postsByUserIDLoadGroup) fieldToColumnMap() map[string]string {
	return postFieldToColumn
}

func (postsByUserIDLoadGroup) query(db *sqlx.DB, sql string, args []interface{}, items []*loadItem) error {
	var result []*PostColumns
	err := db.Select(&result, sql, args...)
	if err != nil {
		return err
	}

	completeByGroupID(items, len(result), func(i int) (interface{}, interface{}) {
		return result[i].AuthorID.String(), result[i]
	}, func(group interface{}, item interface{}) interface{} {
		if group == nil {
			group = make([]*PostColumns, 0)
		}
		return append(group.([]*PostColumns), item.(*PostColumns))
	})

	return nil
}

func (postsByUserIDLoadGroup) extraColumns() []string {
	return []string{"author_id"}
}

type loadItem struct {
	key        interface{}
	fieldNames []string
	done       func(interface{}, error)
}

func completeList(items []*loadItem, data interface{}) {
	for _, item := range items {
		item.done(data, nil)
	}
}

func completeByID(items []*loadItem, n int, get func(i int) (interface{}, interface{})) {
	byID := make(map[interface{}][]*loadItem)
	for _, item := range items {
		k := item.key
		byID[k] = append(byID[k], item)
	}

	for i := 0; i < n; i++ {
		k, v := get(i)
		group := byID[k]
		for _, item := range group {
			item.done(v, nil)
		}
		delete(byID, k)
	}

	for _, group := range byID {
		for _, item := range group {
			item.done(nil, nil)
		}
	}
}

func completeByGroupID(items []*loadItem, n int, get func(i int) (interface{}, interface{}), appendGroup func(interface{}, interface{}) interface{}) {
	itemsByID := make(map[interface{}][]*loadItem)
	for _, item := range items {
		k := item.key
		itemsByID[k] = append(itemsByID[k], item)
	}

	resultsByID := make(map[interface{}]interface{})

	for i := 0; i < n; i++ {
		k, v := get(i)
		resultsByID[k] = appendGroup(resultsByID[k], v)
	}

	for id, results := range resultsByID {
		for _, item := range itemsByID[id] {
			item.done(results, nil)
		}
		delete(itemsByID, id)
	}

	for _, group := range itemsByID {
		for _, item := range group {
			item.done(nil, nil)
		}
	}
}
