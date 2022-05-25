package result

// Wrap constructs a synchronous success or error result from a value, err pair
func Wrap[T any](value T, err error) Result[T] {
	return resultWrap[T]{value, err}
}

type resultWrap[T any] struct {
	value T
	err   error
}

func (r resultWrap[T]) UnpackResult() (interface{}, error) {
	return r.value, r.err
}
