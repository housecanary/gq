package result

import (
	"context"

	"github.com/housecanary/gq/schema"
)

// Chans constructs an asynchronous result by reading a value or error from a
// pair of supplied chans
func Chans[T any](ch <-chan T, errCh <-chan error) Result[T] {
	return resultChans[T]{ch, errCh}
}

type resultChans[T any] struct {
	ch    <-chan T
	errCh <-chan error
}

func (r resultChans[T]) UnpackResult() (interface{}, error) {
	return schema.AsyncValueFunc(func(ctx context.Context) (interface{}, error) {
		var empty T
		select {
		case <-ctx.Done():
			return empty, ctx.Err()
		case in := <-r.ch:
			return in, nil
		case err := <-r.errCh:
			return empty, err
		}
	}), nil
}
