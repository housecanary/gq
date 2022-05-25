package ts_test

import (
	"context"
	"errors"

	"github.com/housecanary/gq/schema/ts"
	"github.com/housecanary/gq/schema/ts/result"
)

type provider struct{}

var otMod = ts.Module[provider]()

type testObject struct{}

var testObjectType = ts.Object(
	otMod, "TestObject",

	ts.Field("a", func(t *testObject) ts.Result[string] {
		if false {
			return result.Error[string](errors.New("test"))
		}

		if false {
			return result.Async(func(context.Context) (string, error) {
				return "", nil
			})
		}

		if false {
			var ch chan testStringResult
			return result.Chan(ch)
		}

		return result.Of("")
	}),

	ts.FieldP("b", otMod, func(p provider, t *testObject) ts.Result[string] {
		return result.Of("")
	}),
)

type testStringResult struct {
	Value string
	Error error
}

type bArgs struct {
	Value string
	Count int
}
