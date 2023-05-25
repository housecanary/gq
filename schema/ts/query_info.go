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

package ts

import (
	"context"
	"reflect"

	"github.com/housecanary/gq/schema"
)

// A QueryInfo is used in a resolver function to get access to information about
// the query being executed.
type QueryInfo interface {
	ArgumentValue(name string) (interface{}, error)
	QueryContext() context.Context
	ChildFieldsIterator() schema.FieldSelectionIterator
}

type internalQueryInfo interface {
	QueryInfo
	setArgumentValue(name string, dest reflect.Value, converter inputConverter) error
}
