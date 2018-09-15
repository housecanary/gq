// Copyright 2018 HouseCanary, Inc.
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

import (
	"testing"

	"github.com/housecanary/gq/schema"
)

func TestParseObjectVariables(t *testing.T) {
	vars, err := NewVariablesFromJSON([]byte(`{
		"s": "string",
		"f": 123.0,
		"b": true,
		"n": null,
		"o": {"a": "b"},
		"a": ["c"]
	}`))
	if err != nil {
		t.Error(err)
	}

	if vars["s"] != schema.LiteralString("string") {
		t.Errorf("Expected \"string\" got %v", vars["s"])
	}

	if vars["f"] != schema.LiteralNumber(123) {
		t.Errorf("Expected 123 got %v", vars["f"])
	}

	if vars["b"] != schema.LiteralBool(true) {
		t.Errorf("Expected true got %v", vars["b"])
	}

	if vars["n"] != nil {
		t.Errorf("Expected nil got %v", vars["n"])
	}

	if lo, ok := vars["o"].(schema.LiteralObject); !ok {
		t.Errorf("Expected {\"a\": \"b\"} got %v", vars["o"])
	} else if lo["a"] != schema.LiteralString("b") {
		t.Errorf("Expected {\"a\": \"b\"} got %v", vars["o"])
	}

	if lo, ok := vars["a"].(schema.LiteralArray); !ok {
		t.Errorf("Expected [\"c\"] got %v", vars["a"])
	} else if lo[0] != schema.LiteralString("c") {
		t.Errorf("Expected [\"c\"] got %v", vars["c"])
	}
}
