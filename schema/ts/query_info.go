package ts

import (
	"context"

	"github.com/housecanary/gq/schema"
)

// A QueryInfo is used in a resolver function to get access to information about
// the query being executed.
type QueryInfo interface {
	ArgumentValue(name string) (interface{}, error)
	QueryContext() context.Context
	ChildFieldsIterator() schema.FieldSelectionIterator
}
