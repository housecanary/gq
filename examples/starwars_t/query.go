package starwars

import (
	"fmt"

	"github.com/housecanary/gq/schema/ts"
	"github.com/housecanary/gq/schema/ts/result"
	"github.com/housecanary/gq/types"
)

type Query struct {
}

func (q *Query) resolveHuman(li *lookupInput) (*human, error) {
	id := li.ID
	name := li.Name
	if !id.Nil() {
		return humans[id.String()], nil
	}

	if !name.Nil() {
		for _, v := range humans {
			if v.Name == name {
				return v, nil
			}
		}
		return nil, nil
	}

	return nil, fmt.Errorf("Must supply either id or name to lookup")
}

var queryType = ts.Object(
	modType,
	`"Information about Star Wars"`,

	ts.Field(
		`
		"A random human or droid.  Selected by a fair roll of a die."
		random
		`,
		func(q *Query) ts.Result[humanOrDroid] {
			return result.Of(humanOrDroidFromHuman(&human{
				ID:   types.NewID("random one"),
				Name: types.NewString("John Q Random"),
			}))
		},
	),

	ts.FieldA(
		`
		"A human by their unique ID or name"
		human
		`,
		func(q *Query, args *struct {
			Lookup *lookupInput
		}) ts.Result[*human] {
			return result.Wrap(q.resolveHuman(args.Lookup))
		},
	),

	ts.FieldA(
		`
		"Humans by their unique ID or name"
		humans
		`,
		func(q *Query, args *struct {
			Lookups []*lookupInput
		}) ts.Result[[]*human] {
			var humans []*human
			for _, li := range args.Lookups {
				h, err := q.resolveHuman(li)
				if err != nil {
					return result.Error[[]*human](err)
				}
				humans = append(humans, h)
			}
			return result.Of(humans)
		},
	),

	ts.FieldA(
		`
		"A droid by their unique ID or name"
		droid
		`,
		func(q *Query, args *struct {
			Lookup *lookupInput
		}) ts.Result[*droid] {
			lookup := args.Lookup

			if !lookup.ID.Nil() {
				return result.Of(droids[lookup.ID.String()])
			}

			if !lookup.Name.Nil() {
				for _, v := range droids {
					if v.Name == lookup.Name {
						return result.Of(v)
					}
				}
				return result.Of((*droid)(nil))
			}

			return result.Error[*droid](fmt.Errorf("Must supply either id or name to lookup"))
		},
	),

	ts.FieldA(
		`
		"The hero of an episode"
		hero
		`,
		func(q *Query, args *struct {
			Episode Episode
		}) ts.Result[character] {
			c := make(chan character, 1)
			e := make(chan error)
			switch args.Episode {
			case "NEWHOPE":
				c <- characterFromHuman(humans["h1"])
			default:
				c <- characterType.Nil()
			}

			return result.Chans(c, e)
		},
	),
)
