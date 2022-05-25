package result

// Error creates a synchronous error result
func Error[T any](err error) Result[T] {
	return resultError[T]{err}
}

type resultError[T any] struct {
	value error
}

func (r resultError[T]) UnpackResult() (interface{}, error) {
	var emptyT T
	return emptyT, r.value
}
