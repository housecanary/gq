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

package structschema_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/housecanary/gq/query"
	"github.com/housecanary/gq/schema/structschema"
	"github.com/housecanary/gq/types"
)

// Root is the root query type of the schema
type Root struct {
	structschema.Meta `"Root Query Object" {
		"Says hello to the named person"
		hello(name: String): String

		"Says hello politely"
		politeHello(title: TitleInput!): String

		"Says hello politely to multiple people"
		politeHellos(titles: [TitleInput!]): [String!]

		"Returns a canned greeting"
		cannedHello(type: CannedHelloType = BRIEF): String

		"Returns an object that can say hello (interface)"
		greeter: Greeter

		"Returns an object that can say hello (union)"
		greeterByName(name: String!): HumanOrRobot
	}`
	RandomNumber types.Int `gq:";A random number. Selected by a fair roll of a die."`
}

func (r *Root) ResolveHello(name types.String) types.String {
	// if the input is nil, nothing to do
	if name.Nil() {
		return types.NilString()
	}

	return types.NewString(fmt.Sprintf("Hello %v", name.String()))
}

func (r *Root) ResolvePoliteHello(title *TitleInput) types.String {
	greeting := fmt.Sprintf("Hello %s", title.Title.String())

	for _, e1 := range title.AdditionalTitles {
		for _, e2 := range e1 {
			greeting += " " + e2.String()
		}
	}

	if title.LastName.Nil() {
		greeting += " " + title.FirstName.String()
	} else {
		greeting += " " + title.LastName.String()

		if !title.FirstName.Nil() {
			greeting += ".  May I call you " + title.FirstName.String()
		}
	}

	return types.NewString(greeting)
}

func (r *Root) ResolvePoliteHellos(titles []*TitleInput) []types.String {
	var result []types.String
	for _, ti := range titles {
		result = append(result, r.ResolvePoliteHello(ti))
	}
	return result
}

func (r *Root) ResolveCannedHello(typ CannedHelloType) (types.String, error) {
	switch typ.String() {
	case "BRIEF":
		return types.NewString("Hi!"), nil
	case "LONG":
		return types.NewString("Good day"), nil
	case "ELABORATE":
		return types.NilString(), fmt.Errorf("Elaborate hellos are not yet implemented")
	}
	panic("Unreachable")
}

// Example of a resolver with injected arguments, returning an interface
func (r *Root) ResolveGreeter(randomizer *randomizer) Greeter {
	switch randomizer.random() {
	case 0:
		return Greeter{
			&Human{
				Greeting: types.NewString("Salutations"),
				Name:     types.NewString("Bob Smith"),
			},
		}
	case 1:
		return Greeter{
			&Robot{
				Greeting:    types.NewString("Beep Boop Bop"),
				ModelNumber: types.NewString("RX-123"),
			},
		}
	default:
		return Greeter{
			&Tree{
				Greeting: types.NewString("..."),
				Height:   types.NewInt(6),
			},
		}
	}
}

func (r *Root) ResolveGreeterByName(name types.String) (HumanOrRobot, error) {
	if name.String() == "Bob Smith" {
		return HumanOrRobot{
			&Human{
				Greeting: types.NewString("Salutations"),
				Name:     types.NewString("Bob Smith"),
			},
		}, nil
	} else if name.String() == "RX-123" {
		return HumanOrRobot{
			&Robot{
				Greeting:    types.NewString("Beep Boop Bop"),
				ModelNumber: types.NewString("RX-123"),
			},
		}, nil
	}

	return HumanOrRobot{}, fmt.Errorf("Thing with name %s not found", name.String())
}

// TitleInput is an example of an input object
type TitleInput struct {
	structschema.InputObject `"A name and an optional title"`
	FirstName                types.String
	LastName                 types.String
	Title                    types.String `gq:":String!;Title of the person (i.e. Dr/Mr/Mrs/Ms etc)"`
	AdditionalTitles         [][]types.String
}

// If an input object has a Validate method matching this signature, it will
// be invoked on input.
func (t *TitleInput) Validate() error {
	if t.FirstName.Nil() && t.LastName.Nil() {
		return fmt.Errorf("Either first name or last name (or both) must be provided")
	}
	return nil
}

// CannedHelloType is an example of an enum
type CannedHelloType struct {
	structschema.Enum `"The type of canned greeting to return" {
		"A brief greeting"
		BRIEF

		"A longer greeting"
		LONG

		"An elaborate greeting"
		ELABORATE
	}`
}

// Greeter is an example of an interface
type Greeter struct {
	Interface interface {
		isGreeter()
	} `"A greeter is any object that knows how to provide a greeting" {
		greeting: String!
	}`
}

// HumanOrRobot is an example of a union
type HumanOrRobot struct {
	Union interface {
		isHumanOrRobot()
	}
}

// A Human is an object type
type Human struct {
	structschema.Meta `"A human represents a person" {
		mood: String

		"Returns what this person is currently working on"
		currentActivity: String
	}`
	Greeting types.String `gq:":String!"`
	Name     types.String
}

func (Human) isGreeter()      {}
func (Human) isHumanOrRobot() {}

// Resolves the mood of a human using a function style async return
func (h *Human) ResolveMood() func() (types.String, error) {
	type moodResponse struct {
		err  error
		mood types.String
	}
	c := make(chan moodResponse)
	go func() {
		// Here we would talk to the human an ask what mood they're in
		c <- moodResponse{
			mood: types.NewString("Good"),
		}
	}()

	return func() (types.String, error) {
		result := <-c
		return result.mood, result.err
	}
}

// Resolves the current activity of a human using a channel style async return
func (h *Human) ResolveCurrentActivity() (<-chan types.String, <-chan error) {
	c := make(chan types.String)
	e := make(chan error)
	go func() {
		// Here we would talk to the human an ask what they are working on
		e <- fmt.Errorf("Could not contact human")
	}()

	return c, e
}

// A Robot is an object type
type Robot struct {
	Greeting    types.String `gq:":String!"`
	ModelNumber types.String
}

func (Robot) isGreeter()      {}
func (Robot) isHumanOrRobot() {}

// A Tree is an object type
type Tree struct {
	Greeting types.String `gq:":String!"`
	Height   types.Int
}

func (Tree) isGreeter() {}

type randomizer struct {
	randomValue int
}

func (r *randomizer) random() int {
	return r.randomValue
}

func Example() {
	builder := structschema.Builder{
		Types: []interface{}{
			&Root{},
			&Human{},
			&Robot{},
			&Tree{},
		},
	}

	// Register a value to be injected to resolvers.  In this case we use a deterministic random value
	// so our output is consistent.
	rando := &randomizer{randomValue: 2}
	builder.RegisterArgProvider("*structschema_test.randomizer", func(ctx context.Context) interface{} {
		return rando
	})

	schema := builder.MustBuild("Root")

	io.WriteString(os.Stdout, "---- Generated schema ----\n")
	schema.WriteDefinition(os.Stdout)
	io.WriteString(os.Stdout, "\n---- End generated schema ----\n")

	q, err := query.PrepareQuery(`{
		randomNumber
		hello(name: "Bob")
		politeHello(title: {
			title: "Mr"
			lastName: "Random"
		})
		politeHellos(titles: [{
			title: "Mr"
			lastName: "Random"
		}, {
			title: "Mrs"
			lastName: "Random"
		}])
		politeHelloLong: politeHello(title: {
			title: "Professor"
			additionalTitles: [["Doctor"]]
			lastName: "Random"
		})
		politeHelloError: politeHello(title: {
			title: "Mr"
		})
		cannedHello
		cannedHelloLong: cannedHello(type: LONG)
		cannedHelloElaborate: cannedHello(type: ELABORATE)
		greeter {
			greeting
		}

		greeterByNameHuman: greeterByName(name: "Bob Smith") {
			... on Human {
				name
				mood
				currentActivity
			}
		}

		greeterByNameDroid: greeterByName(name: "RX-123") {
			... on Robot {
				modelNumber
			}
		}
	}`, "", schema)

	if err != nil {
		panic(err)
	}

	io.WriteString(os.Stdout, "---- Query ----\n")
	data := q.Execute(context.Background(), &Root{RandomNumber: types.NewInt(7)}, nil, nil)
	buf := &bytes.Buffer{}
	_ = json.Indent(buf, data, "", "  ")
	os.Stdout.Write(buf.Bytes())
	io.WriteString(os.Stdout, "\n---- End query ----\n")
	// Output:
	// ---- Generated schema ----
	// schema {
	//   query: Root
	//
	//   "The type of canned greeting to return"
	//   enum CannedHelloType {
	//     "A brief greeting"
	//     BRIEF
	//
	//     "An elaborate greeting"
	//     ELABORATE
	//
	//     "A longer greeting"
	//     LONG
	//   }
	//
	//   "A greeter is any object that knows how to provide a greeting"
	//   interface Greeter {
	//     greeting: String!
	//   }
	//
	//   "A human represents a person"
	//   object Human implements & Greeter {
	//     "Returns what this person is currently working on"
	//     currentActivity: String
	//
	//     greeting: String!
	//
	//     mood: String
	//
	//     name: String
	//   }
	//
	//   union HumanOrRobot = | Human | Robot
	//
	//   object Robot implements & Greeter {
	//     greeting: String!
	//
	//     modelNumber: String
	//   }
	//
	//   "Root Query Object"
	//   object Root {
	//     "Returns a canned greeting"
	//     cannedHello (
	//       type: CannedHelloType = BRIEF
	//     ): String
	//
	//     "Returns an object that can say hello (interface)"
	//     greeter: Greeter
	//
	//     "Returns an object that can say hello (union)"
	//     greeterByName (
	//       name: String!
	//     ): HumanOrRobot
	//
	//     "Says hello to the named person"
	//     hello (
	//       name: String
	//     ): String
	//
	//     "Says hello politely"
	//     politeHello (
	//       title: TitleInput!
	//     ): String
	//
	//     "Says hello politely to multiple people"
	//     politeHellos (
	//       titles: [TitleInput!]
	//     ): [String!]
	//
	//     "A random number. Selected by a fair roll of a die."
	//     randomNumber: Int
	//   }
	//
	//   "A name and an optional title"
	//   input TitleInput {
	//     additionalTitles: [[String]]
	//
	//     firstName: String
	//
	//     lastName: String
	//
	//     "Title of the person (i.e. Dr/Mr/Mrs/Ms etc)"
	//     title: String!
	//   }
	//
	//   object Tree implements & Greeter {
	//     greeting: String!
	//
	//     height: Int
	//   }
	// }
	// ---- End generated schema ----
	// ---- Query ----
	// {
	//   "data": {
	//     "randomNumber": 7,
	//     "hello": "Hello Bob",
	//     "politeHello": "Hello Mr Random",
	//     "politeHellos": [
	//       "Hello Mr Random",
	//       "Hello Mrs Random"
	//     ],
	//     "politeHelloLong": "Hello Professor Doctor Random",
	//     "politeHelloError": null,
	//     "cannedHello": "Hi!",
	//     "cannedHelloLong": "Good day",
	//     "cannedHelloElaborate": null,
	//     "greeter": {
	//       "greeting": "..."
	//     },
	//     "greeterByNameHuman": {
	//       "name": "Bob Smith",
	//       "mood": "Good",
	//       "currentActivity": null
	//     },
	//     "greeterByNameDroid": {
	//       "modelNumber": "RX-123"
	//     }
	//   },
	//   "errors": [
	//     {
	//       "message": "Error resolving argument title: Error in argument title: Either first name or last name (or both) must be provided",
	//       "path": [
	//         "politeHelloError"
	//       ],
	//       "locations": [
	//         {
	//           "line": 21,
	//           "column": 3
	//         }
	//       ]
	//     },
	//     {
	//       "message": "Elaborate hellos are not yet implemented",
	//       "path": [
	//         "cannedHelloElaborate"
	//       ],
	//       "locations": [
	//         {
	//           "line": 26,
	//           "column": 3
	//         }
	//       ]
	//     },
	//     {
	//       "message": "Could not contact human",
	//       "path": [
	//         "greeterByNameHuman",
	//         "currentActivity"
	//       ],
	//       "locations": [
	//         {
	//           "line": 35,
	//           "column": 5
	//         }
	//       ]
	//     }
	//   ]
	// }
	// ---- End query ----
	//
}
