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
	"github.com/housecanary/gq/types"
)

var inputObjectModType = ts.NewModule()

// To create an input object type, first declare a struct that represents the
// shape of the GraphQL input object
type AddressInput struct {
	HouseNumber types.String
	Street      types.String
	City        types.String
	State       types.String
	PostalCode  types.String
}

// Next, construct the GQL type using the ts.InputObject function
var AddressInputType = ts.NewInputObjectType[AddressInput](inputObjectModType, `"The components of a US street address"`)

// Here's a simple object that makes use of the AddressInput input object in its arguments
type addressQuery struct{}

type validArgs struct {
	Address *AddressInput `gg:";The address to validate"`
}

var addressQueryType = ts.NewObjectType[addressQuery](inputObjectModType, `"Queries on addresses"`)

var addressQueryValidField = ts.AddFieldWithArgs(
	addressQueryType,
	`
	"Checks if an address is valid"
	valid
	`,
	func(q *addressQuery, args *validArgs) ts.Result[types.Boolean] {
		// Silly example: return true if the house number is "1"
		if args.Address.HouseNumber.String() == "1" {
			return result.Of(types.NewBoolean(true))
		}
		return result.Of(types.NewBoolean(false))
	},
)

func ExampleNewInputObjectType() {
	// Now the input objects can be used as arguments to a resolver function
	// (see Object function and the Field methods for more on resolver functions and arguments)
	//
	// See the example above.
}
