package starwars

import "github.com/housecanary/gq/types"

var humans = map[string]*human{
	"h1": &human{
		ID:              types.NewID("h1"),
		Name:            types.NewString("Luke Skywalker"),
		SecretBackstory: types.NewString("A long time ago"),
		HomePlanet:      types.NewString("Tatooine"),
	},
}

var droids = map[string]*droid{
	"d1": &droid{
		ID:              types.NewID("d1"),
		Name:            types.NewString("C3PO"),
		SecretBackstory: types.NewString("A man and a vat of gold...accidents happen"),
		PrimaryFunction: types.NewString("Protocol Droid"),
	},
}
