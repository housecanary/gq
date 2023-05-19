package ts_test

import (
	"github.com/housecanary/gq/schema/ts"
)

var utMod = ts.NewModule()

type testUnion ts.Union

type foo struct{}

var fooType = ts.NewObjectType[foo](utMod, "")

type bar struct{}

var barType = ts.NewObjectType[bar](utMod, "")

var unionType = ts.NewUnionType[testUnion](utMod, "")

var TestUnionFromFoo = ts.UnionMember(unionType, fooType)
var TestUnionFromBar = ts.UnionMember(unionType, barType)
