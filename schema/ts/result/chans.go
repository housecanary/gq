package result

import (
	"context"
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

func (r resultChans[T]) UnpackResult() (T, func(context.Context) (T, error), error) {
	return empty[T](), func(ctx context.Context) (T, error) {
		select {
		case <-ctx.Done():
			return empty[T](), ctx.Err()
		case in := <-r.ch:
			return in, nil
		case err := <-r.errCh:
			return empty[T](), err
		}
	}, nil
}
