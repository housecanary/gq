package starwars

import (
	"github.com/housecanary/gq/schema/ts"
	"github.com/housecanary/gq/types"
)

var starwarsModule = ts.NewModule()

type character ts.Interface[any]

var characterType = ts.NewInterfaceType[character](starwarsModule, `"A character.  Could be human, alien, droid, etc." {
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

var humanType = ts.NewObjectType[human](starwarsModule, `"A humanoid character" Human`)

var characterFromHuman = ts.Implements(humanType, characterType)

type droid struct {
	ID              types.ID     `gq:"id:ID!;The ID of the droid"`
	Name            types.String `gq:";The name of the droid as assigned by the manufacturer"`
	SecretBackstory types.String `gq:";Production date and factory"`
	PrimaryFunction types.String
}

var droidType = ts.NewObjectType[droid](starwarsModule, `"A mechanical character"`)

var characterFromDroid = ts.Implements(droidType, characterType)

type humanOrDroid ts.Union

var humanOrDroidType = ts.NewUnionType[humanOrDroid](starwarsModule, `
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

var lookupInputType = ts.NewInputObjectType[lookupInput](starwarsModule, `"Either a name or a id"`)
