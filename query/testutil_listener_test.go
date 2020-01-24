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

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

type dummyExecutionListener struct {
	notifyResolve func(queryField *ast.Field, schemaField *schema.FieldDescriptor) (ResolveCompleteCallback, error)
	notifyIdle    func()
	notifyError   func(err error)
}

func (l dummyExecutionListener) NotifyResolve(queryField *ast.Field, schemaField *schema.FieldDescriptor) (ResolveCompleteCallback, error) {
	if l.notifyResolve != nil {
		return l.notifyResolve(queryField, schemaField)
	}
	return nil, nil
}

func (l dummyExecutionListener) NotifyIdle() {
	if l.notifyIdle != nil {
		l.notifyIdle()
	}
}

func (l dummyExecutionListener) NotifyError(err error) {
	if l.notifyError != nil {
		l.notifyError(err)
	}
}

type executionListenerAssertion interface {
	NotifyResolve(l *assertExecutionListener, queryField *ast.Field, schemaField *schema.FieldDescriptor) (ResolveCompleteCallback, error)
	NotifyIdle(l *assertExecutionListener)
	NotifyError(l *assertExecutionListener, err error)
}
type baseExecutionListenerAssertion struct{}

func (baseExecutionListenerAssertion) NotifyResolve(l *assertExecutionListener, queryField *ast.Field, schemaField *schema.FieldDescriptor) (ResolveCompleteCallback, error) {
	panic(fmt.Errorf("Unexpected call to NotifyResolve: %v, %v", *queryField, *schemaField))
}

func (baseExecutionListenerAssertion) NotifyIdle(l *assertExecutionListener) {
	panic(fmt.Errorf("Unexpected call to NotifyIdle"))
}

func (baseExecutionListenerAssertion) NotifyError(l *assertExecutionListener, err error) {
	panic(fmt.Errorf("Unexpected call to NotifyError: %v", err))
}

type resolveAssertion struct {
	baseExecutionListenerAssertion
	queryField  *ast.Field
	schemaField *schema.FieldDescriptor
}

func (a resolveAssertion) NotifyResolve(l *assertExecutionListener, queryField *ast.Field, schemaField *schema.FieldDescriptor) (ResolveCompleteCallback, error) {
	if a.queryField != queryField {
		panic(fmt.Errorf("Invalid call to NotifyResolve: %v != %v", *a.queryField, *queryField))
	}

	if a.schemaField != schemaField {
		panic(fmt.Errorf("Invalid call to NotifyResolve: %v != %v", *a.schemaField, *schemaField))
	}

	if l.pendingCallbacks == nil {
		l.pendingCallbacks = make(map[*resolveAssertion]bool)
	}
	l.pendingCallbacks[&a] = true
	return func(i interface{}, e error) error {
		delete(l.pendingCallbacks, &a)
		return e
	}, nil
}

type idleAssertion struct {
	baseExecutionListenerAssertion
}

func (idleAssertion) NotifyIdle(l *assertExecutionListener) {
}

type errorAssertion struct {
	baseExecutionListenerAssertion
	err error
}

func (a errorAssertion) NotifyError(l *assertExecutionListener, err error) {
	if a.err != err {
		panic(fmt.Errorf("Invalid call to NotifyError: %v != %v", a.err, err))
	}
}

type assertExecutionListener struct {
	assertions       []executionListenerAssertion
	pendingCallbacks map[*resolveAssertion]bool
}

func (l *assertExecutionListener) nextAssertion() executionListenerAssertion {
	if len(l.assertions) == 0 {
		return baseExecutionListenerAssertion{}
	}

	r := l.assertions[0]
	l.assertions = l.assertions[1:]
	return r
}

func (l *assertExecutionListener) assertDone() {
	if len(l.assertions) != 0 {
		panic(fmt.Errorf("Unconsumed assertions: %v", spew.Sdump(l.assertions)))
	}

	if len(l.pendingCallbacks) != 0 {
		panic(fmt.Errorf("Unconsumed callbacks: %v", spew.Sdump(l.pendingCallbacks)))
	}
}
func (l *assertExecutionListener) NotifyResolve(queryField *ast.Field, schemaField *schema.FieldDescriptor) (ResolveCompleteCallback, error) {
	return l.nextAssertion().NotifyResolve(l, queryField, schemaField)
}

func (l *assertExecutionListener) NotifyIdle() {
	l.nextAssertion().NotifyIdle(l)
}

func (l *assertExecutionListener) NotifyError(err error) {
	l.nextAssertion().NotifyError(l, err)
}

type idleCountExecutionListener struct {
	BaseExecutionListener
	idleCount int
}

func (c *idleCountExecutionListener) NotifyIdle() {
	c.idleCount++
}
