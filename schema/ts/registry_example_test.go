package ts_test

import (
	"fmt"
	"strings"

	"github.com/housecanary/gq/schema/ts"
)

var queryMod = ts.NewModule()

type query struct {
	Catalog      *Catalog
	Meal         Meal
	Vehicle      Vehicle
	AddressQuery *addressQuery
	Episode      Episode
	DateTime     DateTime
}

var queryType = ts.NewObjectType[query](queryMod, ``)

func ExampleNewTypeRegistry() {
	sb, err := ts.NewTypeRegistry(
		ts.WithModule(queryMod),
		ts.WithModule(enumModType),
		ts.WithModule(inputObjectModType),
		ts.WithModule(interfaceModType),
		ts.WithModule(objectModType),
		ts.WithModule(scalarModType),
		ts.WithModule(unionModType),
	)
	if err != nil {
		panic(err)
	}
	s := sb.MustBuildSchema("Query")

	buf := &strings.Builder{}
	s.WriteDefinition(buf)
	fmt.Println(buf)

	// Output:
	// schema {
	//   query: Query
	//
	//   "The components of a US street address"
	//   input AddressInput {
	//     city: String
	//
	//     houseNumber: String
	//
	//     postalCode: String
	//
	//     state: String
	//
	//     street: String
	//   }
	//
	//   "Queries on addresses"
	//   object AddressQuery {
	//     "Checks if an address is valid"
	//     valid (
	//       address: AddressInput
	//     ): Boolean
	//   }
	//
	//   "Enclosed vehicle with 4 wheels"
	//   object Car implements & Vehicle {
	//     passengers: Int
	//
	//     sound: String
	//
	//     topSpeed: Int
	//   }
	//
	//   "A product catalog"
	//   object Catalog {
	//     coverPictureURL: String @relativeUrl
	//
	//     id: ID
	//
	//     "The issue date of the catalog expressed as a ISO date"
	//     issueDate: String
	//
	//     "An image of a specific page"
	//     pageImageUrl (
	//       pageNumber: Int!
	//     ): String
	//
	//     "Images of all pages in the catalog"
	//     pageImageUrls: [String]
	//
	//     pages: Int!
	//
	//     "The next catalog that replaces this one"
	//     replacement: Catalog
	//
	//     "The URL of a thumbnail image of this catalog"
	//     thumbnailUrl: String! @relativeUrl
	//
	//     validityEndDate: String
	//   }
	//
	//   "An ISO format datetime."
	//   scalar DateTime
	//
	//   "All of the episodes that count"
	//   enum Episode {
	//     "Released in 1980."
	//     EMPIRE
	//
	//     "Released in 1983."
	//     JEDI
	//
	//     "Released in 1977."
	//     NEWHOPE
	//   }
	//
	//   object Hamburger
	//
	//   object Hotdog
	//
	//   "Different meals"
	//   union Meal = | Hamburger | Hotdog
	//
	//   "Open vehicle with 2 wheels"
	//   object Motorcycle implements & Vehicle {
	//     hasSidecar: Boolean
	//
	//     sound: String
	//
	//     topSpeed: Int
	//   }
	//
	//   object Query {
	//     addressQuery: AddressQuery
	//
	//     catalog: Catalog
	//
	//     dateTime: DateTime
	//
	//     episode: Episode
	//
	//     meal: Meal
	//
	//     vehicle: Vehicle
	//   }
	//
	//   "The commmon fields of vehicles"
	//   interface Vehicle {
	//     sound: String
	//
	//     topSpeed: Int
	//   }
	// }
}
