package starwars

import (
	ss "github.com/housecanary/gq/schema/structschema"
)

type Episode struct {
	ss.Enum `"All of the episodes that count" {
		"Released in 1977."
		NEWHOPE
	  
		"Released in 1980."
		EMPIRE
	  
		"Released in 1983."
		JEDI
	}`
}
