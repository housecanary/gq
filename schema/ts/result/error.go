package result

import "context"

// Error creates a synchronous error result
func Error[T any](err error) Result[T] {
	return resultError[T]{err}
}

type resultError[T any] struct {
	value error
}

func (r resultError[T]) UnpackResult() (T, func(context.Context) (T, error), error) {
	return empty[T](), nil, r.value
}
