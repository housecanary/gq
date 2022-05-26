package starwars

import (
	"github.com/housecanary/gq/schema/ts"
	"github.com/housecanary/gq/types"
)

var modType = ts.Module()

type character ts.InterfaceBox

var characterType = ts.Interface[character](modType, `"A character.  Could be human, alien, droid, etc." {
	"The id of the character."
	id: ID!

	"The name of the character."
	name: String

	"All secrets about their past."
	secretBackstory: String
}`)

type human struct {
	ID              types.ID     `gq:"id:ID!;The ID of the human"`
	Name            types.String `gq:";The name of the human"`
	SecretBackstory types.String
	HomePlanet      types.String
}

var humanType = ts.Object[human](modType, `"A humanoid character" Human`)

var characterFromHuman = ts.Implements(humanType, characterType)

type droid struct {
	ID              types.ID     `gq:"id:ID!;The ID of the droid"`
	Name            types.String `gq:";The name of the droid as assigned by the manufacturer"`
	SecretBackstory types.String `gq:";Production date and factory"`
	PrimaryFunction types.String
}

var droidType = ts.Object[droid](modType, `"A mechanical character"`)

var characterFromDroid = ts.Implements(droidType, characterType)

type humanOrDroid ts.UnionBox

var humanOrDroidType = ts.Union[humanOrDroid](modType, `
	"Either a human or a droid"
`)

var (
	humanOrDroidFromHuman = ts.UnionMember(humanOrDroidType, humanType)
	humanOrDroidFromDroid = ts.UnionMember(humanOrDroidType, droidType)
)

type lookupInput struct {
	ID   types.ID `gq:"id"`
	Name types.String
}

var lookupInputType = ts.InputObject[lookupInput](modType, `"Either a name or a id"`)
