// Copyright 2023 HouseCanary, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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

// MapChan constructs an asynchronous result that reads a value from the supplied
// chan and uses the supplied function to transform it to an output value.
func Map[S, T any](r Result[S], f func(S) T) Result[T] {
	return resultMap[S, T]{r, f}
}

type resultMap[S, T any] struct {
	s Result[S]
	f func(S) T
}

func (r resultMap[S, T]) UnpackResult() (T, func(context.Context) (T, error), error) {
	s, asyncS, err := r.s.UnpackResult()
	if err != nil {
		return empty[T](), nil, err
	}
	if asyncS != nil {
		return empty[T](), func(ctx context.Context) (T, error) {
			s, err := asyncS(ctx)
			if err != nil {
				return empty[T](), err
			}
			return r.f(s), nil
		}, nil
	}
	return r.f(s), nil, nil
}
