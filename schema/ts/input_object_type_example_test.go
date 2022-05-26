package ts_test

import (
	"github.com/housecanary/gq/schema/ts"
	"github.com/housecanary/gq/schema/ts/result"
	"github.com/housecanary/gq/types"
)

var inputObjectModType = ts.Module()

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
var AddressInputType = ts.InputObject[AddressInput](inputObjectModType, `"The components of a US street address"`)

// Here's a simple object that makes use of the AddressInput input object in its arguments
type addressQuery struct{}

var addressQueryType = ts.Object(
	inputObjectModType, `"Queries on addresses"`,
	ts.FieldA(
		`
		"Checks if an address is valid"
		valid
		`,
		func(q *addressQuery, args *struct {
			Address *AddressInput `gg:";The address to validate"`
		}) ts.Result[types.Boolean] {
			// Silly example: return true if the house number is "1"
			if args.Address.HouseNumber.String() == "1" {
				return result.Of(types.NewBoolean(true))
			}
			return result.Of(types.NewBoolean(false))
		},
	),
)

func ExampleInputObject() {
	// Now the input objects can be used as arguments to a resolver function
	// (see Object function and the Field methods for more on resolver functions and arguments)
	//
	// See the example above.
}
