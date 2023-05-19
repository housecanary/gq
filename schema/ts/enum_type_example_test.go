package ts_test

import (
	"fmt"

	"github.com/housecanary/gq/schema/ts"
)

var enumModType = ts.NewModule()

// To create an enum type, first declare a typedef from string containing the name
// of your enum
type Episode string

// Next, construct the GQL type using the ts.NewEnumType function
var episodeType = ts.NewEnumType[Episode](enumModType, `"All of the episodes that count"`)

// Next, use the Value function of the type to create the values of the enum. The
// Value function returns a value of type E that corresponds to the GQL value it
// is registered with.
var (
	EpisodeNewHope = episodeType.Value(`
		"Released in 1977."
		NEWHOPE
	`)

	EpisodeEmpire = episodeType.Value(`
		"Released in 1980."
		EMPIRE
	`)

	EpisodeJedi = episodeType.Value(`
		"Released in 1983."
		JEDI
	`)
)

func ExampleNewEnumType() {
	// Now the enum values are usable in normal code:
	var selectedEpisode Episode
	selectedEpisode = EpisodeNewHope
	switch selectedEpisode {
	case EpisodeNewHope:
		fmt.Println("A New Hope")
	case EpisodeEmpire:
		fmt.Println("The Empire Strikes Back")
	case EpisodeJedi:
		fmt.Println("Return Of The Jedi")
	}

	// Output:
	// A New Hope
}
