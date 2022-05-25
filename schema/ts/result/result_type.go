package result

type Result[T any] interface {
	UnpackResult() (interface{}, error)
}
