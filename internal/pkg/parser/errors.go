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

package parser

import (
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
)

type ParseError interface {
	error
	GetLine() int
	GetColumn() int
}

type parseError struct {
	msg    string
	line   int
	column int
}

func (e parseError) Error() string {
	return e.msg
}

func (e parseError) GetLine() int {
	return e.line
}

func (e parseError) GetColumn() int {
	return e.column
}

type panicListener struct {
	antlr.ErrorListener
}

func (panicListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	panic(parseError{msg, line, column})
}
