package ts_test

import (
	"github.com/housecanary/gq/schema/ts"
	"github.com/housecanary/gq/schema/ts/result"
	"github.com/housecanary/gq/types"
)

var interfaceModType = ts.Module()

// To create an interface type, first declare a type derived from InterfaceBox
type Vehicle ts.InterfaceBox

// Next, construct the GQL type using the ts.Interface function
var VehicleType = ts.Interface[Vehicle](interfaceModType, `
	"The commmon fields of vehicles"
	{
		sound: String
		topSpeed: Int
	}
`)

// Now, you can make objects that implement the interface
type car struct {
	Sound    types.String
	TopSpeed types.Int

	Passengers types.Int
}

var carType = ts.Object[car](interfaceModType, `"Enclosed vehicle with 4 wheels"`)
var vehicleFromCar = ts.Implements(carType, VehicleType)

type motorcycle struct {
	Sound    types.String
	TopSpeed types.Int

	HasSidecar types.Boolean
}

var motorcycleType = ts.Object[motorcycle](interfaceModType, `"Open vehicle with 2 wheels"`)
var vehicleFromMotorcycle = ts.Implements(motorcycleType, VehicleType)

func ExampleInterface() {

	// Now you can create functions that return the interface and wrap one of the implementations
	x := 1
	var _ = func() ts.Result[Vehicle] {
		switch x {
		case 1:
			return result.Of(vehicleFromCar(&car{
				Sound:      types.NewString("Vroom"),
				TopSpeed:   types.NewInt(120),
				Passengers: types.NewInt(4),
			}))
		case 2:
			return result.Of(vehicleFromMotorcycle(&motorcycle{
				Sound:      types.NewString("Vroom"),
				TopSpeed:   types.NewInt(180),
				HasSidecar: types.NewBoolean(false),
			}))
		default:
			return result.Of(VehicleType.Nil())
		}
	}
}
