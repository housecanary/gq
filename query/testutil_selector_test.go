// Copyright 2019 HouseCanary, Inc.
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

package query

type dummySelector struct {
	Apply func(ctx *exeContext, value interface{}, collector collector) contFunc
}

func (s dummySelector) required() bool {
	return false
}

func (s dummySelector) apply(ctx *exeContext, value interface{}, collector collector) contFunc {
	if s.Apply != nil {
		return s.Apply(ctx, value, collector)
	}
	return nil
}
