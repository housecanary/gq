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
