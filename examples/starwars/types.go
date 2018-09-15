package starwars

import (
	ss "github.com/housecanary/gq/schema/structschema"
	"github.com/housecanary/gq/types"
)

type character struct {
	Interface interface {
		isCharacter()
	} `"A character.  Could be human, alien, droid, etc." {
		"The id of the character."
		id: ID!

		"The name of the character."
		name: String

		"All secrets about their past."
		secretBackstory: String
	}`
}

type human struct {
	ss.Meta         `"A humanoid character" Human`
	ID              types.ID     `gq:"id:ID!;The ID of the human"`
	Name            types.String `gq:";The name of the human"`
	SecretBackstory types.String
	HomePlanet      types.String
}

func (human) isCharacter() {}

type droid struct {
	ss.Meta         `"A mechanical character"`
	ID              types.ID     `gq:"id:ID!;The ID of the droid"`
	Name            types.String `gq:";The name of the droid as assigned by the manufacturer"`
	SecretBackstory types.String `gq:";Production date and factory"`
	PrimaryFunction types.String
}

func (droid) isCharacter() {}

type humanOrDroid struct {
	Union interface {
		isHumanOrDroid()
	} `"Either a human or a droid"`
}

func (human) isHumanOrDroid() {}
func (droid) isHumanOrDroid() {}

type lookupInput struct {
	ss.InputObject `"Either a name or a id"`
	ID             types.ID `gq:"id"`
	Name           types.String
}
