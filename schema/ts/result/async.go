package result

import (
	"context"
)

// Async constructs an asynchronous result with the result value supplied
// by invoking a function
func Async[T any](f func(context.Context) (T, error)) Result[T] {
	return resultAsync[T]{f}
}

type resultAsync[T any] struct {
	value func(context.Context) (T, error)
}

func (r resultAsync[T]) UnpackResult() (T, func(context.Context) (T, error), error) {
	return empty[T](), r.value, nil
}
