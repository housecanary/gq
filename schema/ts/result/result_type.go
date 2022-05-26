package result

import "context"

type Result[T any] interface {
	UnpackResult() (T, func(context.Context) (T, error), error)
}

func empty[T any]() T {
	var x T
	return x
}
