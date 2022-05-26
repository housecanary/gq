package ts_test

import (
	"github.com/housecanary/gq/schema/ts"
)

var utMod = ts.Module()

type testUnion ts.UnionBox

type foo struct{}

var fooType = ts.Object[foo](utMod, "")

type bar struct{}

var barType = ts.Object[bar](utMod, "")

var unionType = ts.Union[testUnion](utMod, "")

var TestUnionFromFoo = ts.UnionMember(unionType, fooType)
var TestUnionFromBar = ts.UnionMember(unionType, barType)
