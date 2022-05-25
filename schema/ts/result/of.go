package result

// Of constructs a synchronous success result
func Of[T any](value T) Result[T] {
	return resultOf[T]{value}
}

type resultOf[T any] struct {
	value T
}

func (r resultOf[T]) UnpackResult() (interface{}, error) {
	return r.value, nil
}
