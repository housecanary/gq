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
	"fmt"

	"github.com/housecanary/gq/schema/ts"
	"github.com/housecanary/gq/schema/ts/result"
	"github.com/housecanary/gq/types"
)

var objectModType = ts.NewModule()

// To create an object type, declare a struct that represents the
// shape of the GraphQL object.
//
// Public fields of the struct will be turned into GraphQL fields, unless excluded
// by a struct tag of the form `gq:"-"`
//
// Additional metadata about each field can be supplied using the gq struct tag.
// The gq struct tag consists of two parts, separated by a semicolon.
//
// The first part contains a GraphQL schema language definition of the field.
// All parts of the definition may be omitted - any omitted parts will be
// inferred from the type and name of the field.
//
// The second part contains a description of the field.
type Catalog struct {
	ID              types.ID
	IssueDate       types.String `gq:";The issue date of the catalog expressed as a ISO date"` // Add a description
	ValidTill       types.String `gq:"validityEndDate"`                                        // Rename the field
	Pages           types.Int    `gq:":Int!"`                                                  // Change the return type of the field to not nil
	CoverPictureURL types.String `gq:"@relativeUrl"`                                           // Attach a directive

	// All together now
	Thumb types.String `gq:"thumbnailUrl: String! @relativeUrl;The URL of a thumbnail image of this catalog"`

	// Unexported fields will not be included in the schema
	replacementID string
}

// Next, construct the GQL type using the ts.Object function
//
// When constructing the GQL type, additional fields may be registered. These fields
// have their values constructed via functions known as resolvers.
//
// Resolver functions take several forms depending on the input they require. The
// ts.Field* functions account for these different signatures. Each resolver function
// takes at a minimum the address of the struct to which it is attached. In addition,
// the following components can be supplied as input arguments:
//
// QueryInfo - Information about the currently executing query
// Arguments - Definition of function arguments
//
// Since Go does not support method overloading, these features are encoded into
// the method name by using an abbreviation for each argument that is added:
//
// Q when the QueryInfo argument is present
// A when the Arguments argument is present
//
// The Arguments arg is special: it consists of a pointer to a struct. Each public
// field of the struct describes a named argument that can be supplied to the function.
// The fields of the struct can be annotated with gq struct tags to control argument
// types and naming.

var catalogType = ts.NewObjectType[Catalog](objectModType, `"A product catalog"`)

// A computed field
var catalogPageImageUrlsField = ts.AddField(
	catalogType,
	`
	"Images of all pages in the catalog"
	pageImageUrls
	`,
	func(c *Catalog) ts.Result[[]types.String] {
		pageImagesUrlResult := loadPageImages(c.ID.String())

		return result.MapChan(pageImagesUrlResult, func(in []string) ([]types.String, error) {
			out := make([]types.String, len(in))
			for i, url := range in {
				out[i] = types.NewString(url)
			}
			return out, nil
		})
	},
)

type pageImageUrlArgs struct {
	PageNumber types.Int `gq:":Int!"`
}

// A field with arguments
var catalogPageImageUrlField = ts.AddFieldWithArgs(
	catalogType,
	`
	"An image of a specific page"
	pageImageUrl
	`,
	func(c *Catalog, args *pageImageUrlArgs) ts.Result[types.String] {
		pageImageUrlResult, errors := loadPageImage(c.ID.String(), int(args.PageNumber.Int32()))
		return result.Chans(pageImageUrlResult, errors)
	},
)

type replacementArgs struct {
	QueryInfo ts.QueryInfo `ts:"inject"`
}

// A field with context info
var catalogReplacementField = ts.AddFieldWithArgs(
	catalogType,
	`
	"The next catalog that replaces this one"
	replacement
	`,
	func(c *Catalog, a *replacementArgs) ts.Result[*Catalog] {
		// Optimization: only load related catalog details if a field other than it's
		// ID is selected
		needLoadDetails := false
		fi := a.QueryInfo.ChildFieldsIterator()
		for fi.Next() {
			if fi.Selection().Name != "id" {
				needLoadDetails = true
				break
			}
		}

		if !needLoadDetails {
			return result.Of(&Catalog{
				ID: types.NewID(c.replacementID),
			})
		}

		return result.Chan(loadCatalog(c.replacementID))
	},
)

func loadPageImages(id string) chan []string {
	ch := make(chan []string, 1)
	ch <- []string{
		"http://foo/1",
		"http://foo/2",
	}
	return ch
}

func loadPageImage(id string, page int) (chan types.String, chan error) {
	ch := make(chan types.String, 1)
	ch <- types.NewString(fmt.Sprintf("http://foo/%v", page))
	return ch, nil
}

func loadCatalog(id string) chan struct {
	Value *Catalog
	Error error
} {
	return nil
}

func ExampleNewObjectType() {
	// Now the object can be used from GraphQL
	//
	// See the example above.

	// for unit tests, etc it can be convenient to access the resolvers
	// supplied to the field definitions
	var testCatalog Catalog
	catalogPageImageUrlsField.ResolverFunction(&testCatalog)
}
