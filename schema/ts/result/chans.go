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

// Chans constructs an asynchronous result by reading a value or error from a
// pair of supplied chans
func Chans[T any](ch <-chan T, errCh <-chan error) Result[T] {
	return resultChans[T]{ch, errCh}
}

type resultChans[T any] struct {
	ch    <-chan T
	errCh <-chan error
}

func (r resultChans[T]) UnpackResult() (T, func(context.Context) (T, error), error) {
	return empty[T](), func(ctx context.Context) (T, error) {
		select {
		case <-ctx.Done():
			return empty[T](), ctx.Err()
		case in := <-r.ch:
			return in, nil
		case err := <-r.errCh:
			return empty[T](), err
		}
	}, nil
}
