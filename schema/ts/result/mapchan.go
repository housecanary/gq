package result

import (
	"context"

	"github.com/housecanary/gq/schema"
)

// MapChan constructs an asynchronous result that reads a value from the supplied
// chan and uses the supplied function to transform it to an output value.
func MapChan[S, T any](ch <-chan S, f func(S) (T, error)) Result[T] {
	return resultMapChan[S, T]{ch, f}
}

type resultMapChan[S, T any] struct {
	ch <-chan S
	f  func(S) (T, error)
}

func (r resultMapChan[S, T]) UnpackResult() (interface{}, error) {
	return schema.AsyncValueFunc(func(ctx context.Context) (interface{}, error) {
		var empty T
		select {
		case <-ctx.Done():
			return empty, ctx.Err()
		case in := <-r.ch:
			return r.f(in)
		}
	}), nil
}
