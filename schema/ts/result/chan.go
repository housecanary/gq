// Copyright 2023 HouseCanary, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
