package result

import (
	"context"

	"github.com/housecanary/gq/schema"
)

// Async constructs an asynchronous result with the result value supplied
// by invoking a function
func Async[T any](f func(context.Context) (T, error)) Result[T] {
	return resultAsync[T]{f}
}

type resultAsync[T any] struct {
	value func(context.Context) (T, error)
}

func (r resultAsync[T]) UnpackResult() (interface{}, error) {
	return schema.AsyncValueFunc(func(ctx context.Context) (interface{}, error) {
		return r.value(ctx)
	}), nil
}
