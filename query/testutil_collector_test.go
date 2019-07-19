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

type testCollector struct {
	v interface{}
}

func (c *testCollector) Int(v int64) {
	c.v = v
}

func (c *testCollector) Float(v float64) {
	c.v = v
}

func (c *testCollector) Bool(v bool) {
	c.v = v
}

func (c *testCollector) String(v string) {
	c.v = v
}

func (c *testCollector) Error(err error, row, col int) {
	c.v = edat{err, row, col}
}

func (c *testCollector) Required(row, col int) {
}

func (c *testCollector) Object(sizeHint int) objectCollector {
	r := &testObjectCollector{make(map[string]*testCollector)}
	c.v = r
	return r
}

func (c *testCollector) Array(sizeHint int) arrayCollector {
	r := &testArrayCollector{}
	c.v = r
	return r
}

type testObjectCollector struct {
	v map[string]*testCollector
}

func (c *testObjectCollector) Field(name string) collector {
	r := &testCollector{}
	c.v[name] = r
	return r
}

type testArrayCollector struct {
	v []*testCollector
}

func (c *testArrayCollector) Item() collector {
	r := &testCollector{}
	c.v = append(c.v, r)
	return r
}
