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
	"fmt"
	"testing"
)

var err = fmt.Errorf("Test error")

func reportObject(c objectCollector, depth int) {
	if depth == 0 {
		return
	}
	depth--

	c.Field("boolField").Bool(true)
	c.Field("stringField").String("value")
	c.Field("intField").Int(123)
	c.Field("floatField").Float(123.0)
	sub := c.Field("oField").Object(7)
	reportObject(sub, depth)
	a := c.Field("aField").Array(100)
	for i := 0; i < depth; i++ {
		reportObject(a.Item().Object(7), 1)
	}
	c.Field("eField").Error(err, 0, 0)
}

func BenchmarkCollector(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cc := acquireJSONCollectorContext()
		root := &oJSONCollector{cc: cc}
		reportObject(root, 10)
		stream := streamPool.BorrowStream(nil)
		root.serializeJSON(stream, 0)
		streamPool.ReturnStream(stream)
		root.release()
		cc.release()
	}
}
