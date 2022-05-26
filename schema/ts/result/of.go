package result

import "context"

// Of constructs a synchronous success result
func Of[T any](value T) Result[T] {
	return resultOf[T]{value}
}

type resultOf[T any] struct {
	value T
}

func (r resultOf[T]) UnpackResult() (T, func(context.Context) (T, error), error) {
	return r.value, nil, nil
}
