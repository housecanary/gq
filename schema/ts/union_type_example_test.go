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

package ts_test

import (
	"github.com/housecanary/gq/schema/ts"
	"github.com/housecanary/gq/schema/ts/result"
)

var unionModType = ts.NewModule()

// To create a union type, first declare a type derived from Union
type Meal ts.Union

// Next, construct the GQL type using the ts.NewUnionType function
var MealType = ts.NewUnionType[Meal](unionModType, `"Different meals"`)

// Now, you can make objects that are members of the union
type hamburger struct {
}

var hamburgerType = ts.NewObjectType[hamburger](unionModType, ``)
var mealFromHamburger = ts.UnionMember(MealType, hamburgerType)

type hotdog struct {
}

var hotdogType = ts.NewObjectType[hotdog](unionModType, ``)
var mealFromHotdog = ts.UnionMember(MealType, hotdogType)

func ExampleNewUnionType() {

	// Now you can create functions that return the union and wrap one of the implementations
	x := 1
	var _ = func() ts.Result[Meal] {
		switch x {
		case 1:
			return result.Of(MealType.Nil())
		case 2:
			return result.Of(mealFromHotdog(&hotdog{}))
		default:
			return result.Of(mealFromHamburger(&hamburger{}))
		}
	}
}
