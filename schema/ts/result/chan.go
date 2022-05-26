package result

import (
	"context"
)

// Chan constructs an asynchronous result by reading a struct with a Value and Error
// member from a supplied chan
func Chan[T any, S valErr[T]](ch <-chan S) Result[T] {
	return resultChan[T, S]{ch}
}

type valErr[T any] interface {
	~struct {
		Value T
		Error error
	}
}

type resultChan[T any, S valErr[T]] struct {
	ch <-chan S
}

func (r resultChan[T, S]) UnpackResult() (T, func(context.Context) (T, error), error) {
	return empty[T](), func(ctx context.Context) (T, error) {
		var empty T
		select {
		case <-ctx.Done():
			return empty, ctx.Err()
		case in := <-r.ch:
			result := (struct {
				Value T
				Error error
			})(in)
			return result.Value, result.Error
		}
	}, nil
}
