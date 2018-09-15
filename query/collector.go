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
	"sync"

	jsonstream "github.com/json-iterator/go"
)

// MaxPooledArrayCollectorSize sets the threshold for reusing an array collector
var MaxPooledArrayCollectorSize = 1000

// MaxPooledObjectCollectorSize sets the threshold for reusing an object collector
var MaxPooledObjectCollectorSize = 100

// A collector is responsible for collecting the results of a field in the selection tree
// It can have sub collectors, or errors reported to it
//
// The only collector implementation currently is a collector that serializes to
// JSON.  Since there's no client support for other serializations of the GraphQL
// model this is sufficient, but the collector interface allows us to add further
// serializations in the future if needed.
//
// In addition, the collector model is not tied to a one time serialization, and could
// be used to implement subscriptions or streaming results in the future.
type collector interface {
	Int(v int64)
	Float(v float64)
	Bool(v bool)
	String(v string)
	Error(err error, row, col int)
	Required(row, col int)
	Object(sizeHint int) objectCollector
	Array(sizeHint int) arrayCollector
}

// An objectCollector is a collector for object fields
type objectCollector interface {
	Field(name string) collector
}

// An objectCollector is a collector for array elements
type arrayCollector interface {
	Item() collector
}

var errUnexpectedNil = fmt.Errorf("Not null field was null")

// JSON collector implementation
var streamPool = jsonstream.NewStream(
	jsonstream.Config{
		EscapeHTML: true,
	}.Froze(), nil, 0,
).Pool()

type gqlError struct {
	error
	path []interface{}
	row  int
	col  int
}

// constructs a gqlError with an empty path that will be filled
// as the error bubbles up the call stack
func pe(err error, depth int, row, col int) gqlError {
	return gqlError{err, make([]interface{}, depth), row, col}
}

type jsonCollector interface {
	serializeJSON(stream *jsonstream.Stream, depth int) ([]gqlError, bool)
	release()
}

var _ collector = &vJSONCollector{}

// Pool so we can reuse array collectors. Avoids allocating the items arrays over and over.
var apool = &sync.Pool{
	New: func() interface{} {
		return &aJSONCollector{}
	},
}

// Pool so we can reuse object collectors. Avoids allocating the fields arrays over and over.
var opool = &sync.Pool{
	New: func() interface{} {
		return &oJSONCollector{}
	},
}

type vKind int8

const (
	vKindNil vKind = iota
	vKindSerialized
	vKindError
	vKindSub
)

type jsonCollectorContext struct {
	stream *jsonstream.Stream
}

func acquireJSONCollectorContext() jsonCollectorContext {
	return jsonCollectorContext{streamPool.BorrowStream(nil)}
}

func (c jsonCollectorContext) release() {
	streamPool.ReturnStream(c.stream)
}

type vJSONCollector struct {
	cc       jsonCollectorContext
	required bool

	// kind/dat store the value of this node. dat is interpreted according to kind.
	kind vKind
	dat  interface{}
}

// edat stores an error and location information
type edat struct {
	err error
	row int
	col int
}

func (c *vJSONCollector) serializeJSON(stream *jsonstream.Stream, depth int) ([]gqlError, bool) {
	if c.kind == vKindError {
		e := c.dat.(edat)
		errs := []gqlError{pe(e.err, depth, e.row, e.col)}
		if c.required {
			return errs, false
		}
		stream.WriteNil()
		return errs, true
	}

	if c.kind == vKindNil && c.required {
		e := c.dat.(edat)
		return []gqlError{pe(errUnexpectedNil, depth, e.row, e.col)}, false
	}

	switch c.kind {
	case vKindNil:
		stream.WriteNil()
	case vKindSerialized:
		stream.Write(c.dat.([]byte))
	case vKindSub:
		sub := c.dat.(jsonCollector)
		allErrors, ok := sub.serializeJSON(stream, depth)
		if !ok {
			// Error bubbled up from child.  Discard any content
			// written by the child, and either bubble the error
			// or write nil as our value depending if we are required
			if c.required {
				return allErrors, false
			}
			stream.WriteNil()
			return allErrors, true
		}
		return allErrors, true
	}

	return nil, true
}

func (c *vJSONCollector) Int(v int64) {
	if c.kind != vKindNil {
		panic("Set value on already set collector")
	}
	pre := len(c.cc.stream.Buffer())
	c.cc.stream.WriteInt64(v)
	c.kind = vKindSerialized
	c.dat = c.cc.stream.Buffer()[pre:]
}

func (c *vJSONCollector) Float(v float64) {
	if c.kind != vKindNil {
		panic("Set value on already set collector")
	}
	pre := len(c.cc.stream.Buffer())
	c.cc.stream.WriteFloat64(v)
	c.kind = vKindSerialized
	c.dat = c.cc.stream.Buffer()[pre:]
}
func (c *vJSONCollector) Bool(v bool) {
	if c.kind != vKindNil {
		panic("Set value on already set collector")
	}
	pre := len(c.cc.stream.Buffer())
	c.cc.stream.WriteBool(v)
	c.kind = vKindSerialized
	c.dat = c.cc.stream.Buffer()[pre:]
}
func (c *vJSONCollector) String(v string) {
	if c.kind != vKindNil {
		panic("Set value on already set collector")
	}
	pre := len(c.cc.stream.Buffer())
	c.cc.stream.WriteString(v)
	c.kind = vKindSerialized
	c.dat = c.cc.stream.Buffer()[pre:]
}

func (c *vJSONCollector) Object(sizeHint int) objectCollector {
	if c.kind != vKindNil {
		panic("Set value on already set collector")
	}
	if sizeHint < 0 {
		sizeHint = 0
	}
	o := opool.Get().(*oJSONCollector)
	o.cc = c.cc
	if cap(o.fields) < sizeHint {
		o.fields = make([]oJSONField, 0, sizeHint)
	}
	c.dat = o
	c.kind = vKindSub
	return o
}

func (c *vJSONCollector) Array(sizeHint int) arrayCollector {
	if c.kind != vKindNil {
		panic("Set value on already set collector")
	}
	if sizeHint < 0 {
		sizeHint = 0
	}
	a := apool.Get().(*aJSONCollector)
	a.cc = c.cc
	if cap(a.values) < sizeHint {
		a.values = make([]vJSONCollector, 0, sizeHint)
	}
	c.dat = a
	c.kind = vKindSub
	return a
}

func (c *vJSONCollector) Required(row, col int) {
	if c.kind == vKindNil {
		c.dat = edat{nil, row, col}
	}
	c.required = true
}

func (c *vJSONCollector) Error(err error, row, col int) {
	if c.kind == vKindSub { // We allow setting an error even after we set a value, but need to make sure to release if we're clobbering an existing sub-collector
		c.dat.(jsonCollector).release()
	}
	c.kind = vKindError
	c.dat = edat{err, row, col}
}

func (c *vJSONCollector) release() {
	if c.kind == vKindSub {
		sub := c.dat.(jsonCollector)
		sub.release()
	}
	c.kind = vKindNil
	c.dat = nil
}

var _ objectCollector = &oJSONCollector{}

type oJSONField struct {
	name string
	c    vJSONCollector
}

type oJSONCollector struct {
	cc     jsonCollectorContext
	fields []oJSONField
}

func (c *oJSONCollector) Field(name string) collector {
	sub := vJSONCollector{cc: c.cc}
	c.fields = append(c.fields, oJSONField{name, sub})
	return &c.fields[len(c.fields)-1].c
}

func (c *oJSONCollector) serializeJSON(stream *jsonstream.Stream, depth int) ([]gqlError, bool) {
	mark := stream.Buffer()
	var allErrs []gqlError
	allOk := true

	stream.WriteObjectStart()
	for i, f := range c.fields {
		if i != 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField(f.name)
		fieldErrors, ok := f.c.serializeJSON(stream, depth+1)
		for _, fe := range fieldErrors {
			fe.path[depth] = f.name
		}
		allErrs = append(allErrs, fieldErrors...)
		allOk = allOk && ok
	}
	stream.WriteObjectEnd()
	if !allOk {
		stream.SetBuffer(mark)
		return allErrs, false
	}
	return allErrs, true
}

func (c *oJSONCollector) release() {
	lenf := len(c.fields)
	for i := 0; i < lenf; i++ {
		(&c.fields[i].c).release()
	}
	if lenf < MaxPooledObjectCollectorSize {
		c.fields = c.fields[:0]
		opool.Put(c)
	}
}

type aJSONCollector struct {
	cc     jsonCollectorContext
	values []vJSONCollector
}

func (c *aJSONCollector) Item() collector {
	sub := vJSONCollector{cc: c.cc}
	c.values = append(c.values, sub)
	return &c.values[len(c.values)-1]
}

func (c *aJSONCollector) serializeJSON(stream *jsonstream.Stream, depth int) ([]gqlError, bool) {
	mark := stream.Buffer()
	var allErrs []gqlError
	allOk := true

	stream.WriteArrayStart()
	for i, v := range c.values {
		if i != 0 {
			stream.WriteMore()
		}
		fieldErrors, ok := v.serializeJSON(stream, depth+1)
		for _, fe := range fieldErrors {
			fe.path[depth] = i
		}
		allErrs = append(allErrs, fieldErrors...)
		allOk = allOk && ok
	}
	stream.WriteArrayEnd()
	if !allOk {
		stream.SetBuffer(mark)
		return allErrs, false
	}
	return allErrs, true
}

func (c *aJSONCollector) release() {
	lenv := len(c.values)
	for i := 0; i < lenv; i++ {
		(&c.values[i]).release()
	}
	if lenv < MaxPooledArrayCollectorSize {
		c.values = c.values[:0]
		apool.Put(c)
	}
}
