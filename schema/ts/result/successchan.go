package result

import (
	"context"
)

// SuccessChan constructs an asynchronous result by reading a value
// mfrom a supplied chan
func SuccessChan[T any](ch <-chan T) Result[T] {
	return resultSuccessChan[T]{ch}
}

type resultSuccessChan[T any] struct {
	ch <-chan T
}

func (r resultSuccessChan[T]) UnpackResult() (T, func(context.Context) (T, error), error) {
	return empty[T](), func(ctx context.Context) (T, error) {
		var empty T
		select {
		case <-ctx.Done():
			return empty, ctx.Err()
		case in := <-r.ch:
			return in, nil
		}
	}, nil
}
