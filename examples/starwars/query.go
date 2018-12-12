package starwars

import (
	"fmt"

	ss "github.com/housecanary/gq/schema/structschema"
	"github.com/housecanary/gq/types"
)

type Query struct {
	ss.Meta `"Information about Star Wars" {
		"A random human or droid.  Selected by a fair roll of a die."
		random : HumanOrDroid

		"A human by their unique ID or name"
		human(lookup: LookupInput!) : Human

		"A droid by their unique ID or name"
		droid(lookup: LookupInput!) : Droid

		"The hero of an episode"
		hero(episode: Episode) : Character
	}`
}

func (*Query) ResolveRandom() humanOrDroid {
	return humanOrDroid{&human{
		ID:   types.NewID("random one"),
		Name: types.NewString("John Q Random"),
	}}
}

func (*Query) ResolveHuman(lookup *lookupInput) (*human, error) {
	if !lookup.ID.Nil() {
		return humans[lookup.ID.String()], nil
	}

	if !lookup.Name.Nil() {
		for _, v := range humans {
			if v.Name == lookup.Name {
				return v, nil
			}
		}
		return nil, nil
	}

	return nil, fmt.Errorf("Must supply either id or name to lookup")
}

func (*Query) ResolveDroid(lookup *lookupInput) (*droid, error) {
	if !lookup.ID.Nil() {
		return droids[lookup.ID.String()], nil
	}

	if !lookup.Name.Nil() {
		for _, v := range droids {
			if v.Name == lookup.Name {
				return v, nil
			}
		}
		return nil, nil
	}

	return nil, fmt.Errorf("Must supply either id or name to lookup")
}

func (*Query) ResolveHero(episode Episode) (<-chan character, <-chan error) {
	c := make(chan character, 1)
	e := make(chan error)
	switch episode.String() {
	case "NEWHOPE":
		c <- character{humans["h1"]}
	case "DEFAULT":
		c <- character{nil}
	}

	return c, e
}
