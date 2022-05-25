package starwars

import (
	"github.com/housecanary/gq/schema/ts"
)

type Episode string

var episodeType = ts.Enum[Episode](modType, `"All of the episodes that count"`)

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
