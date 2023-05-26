package ts_test

import (
	"context"
	"embed"
	"testing"

	querypkg "github.com/housecanary/gq/query"
	"github.com/housecanary/gq/schema"
	"github.com/housecanary/gq/schema/ts"
	"github.com/housecanary/gq/schema/ts/result"
	"github.com/housecanary/gq/types"
	"github.com/nsf/jsondiff"
)

//go:embed testcases
var testcases embed.FS

type customScalar struct {
	v schema.LiteralValue
}

func (v customScalar) ToLiteralValue() (schema.LiteralValue, error) {
	return v.v, nil
}

func (v *customScalar) FromLiteralValue(l schema.LiteralValue) error {
	v.v = l
	return nil
}

func TestTSSchema(t *testing.T) {
	type testCase struct {
		name   string
		schema func(mod *ts.Module) any
	}

	testCases := []testCase{
		{
			name:   "alltypes",
			schema: buildDefaultSchema,
		},
		{
			name: "lists",
			schema: func(mod *ts.Module) any {
				type Query struct {
					FieldList []types.String
				}
				queryGQLType := ts.NewObjectType[Query](mod, ``)

				type args struct {
					List [][]types.String
				}
				ts.AddFieldWithArgs(queryGQLType, `lists`, func(q *Query, a *args) ts.Result[[][]types.String] {
					return result.Of(a.List)
				})

				return &Query{
					FieldList: []types.String{types.NewString("c")},
				}
			},
		},
		{
			name: "notnils",
			schema: func(mod *ts.Module) any {
				type Query struct {
					FieldList []types.String
				}
				queryGQLType := ts.NewObjectType[Query](mod, ``)

				type args struct {
					List [][]types.String `gq:": [[String!]!]!"`
				}
				ts.AddFieldWithArgs(queryGQLType, `lists: [[String!]!]!`, func(q *Query, a *args) ts.Result[[][]types.String] {
					return result.Of(a.List)
				})

				type args2 struct {
					String types.String `gq:":String!"`
				}
				ts.AddFieldWithArgs(queryGQLType, `requiredArg`, func(q *Query, a *args2) ts.Result[types.String] {
					return result.Of(a.String)
				})

				type inputObject struct {
					RequiredField            types.String `gq:": String!"`
					RequiredFieldWithDefault types.String `gq:": String! = \"test2\""`
				}
				ts.NewInputObjectType[inputObject](mod, ``)

				type args3 struct {
					Input *inputObject
				}
				ts.AddFieldWithArgs(queryGQLType, `requiredField`, func(q *Query, a *args3) ts.Result[types.String] {
					return result.Of(types.NewString(a.Input.RequiredField.String() + a.Input.RequiredFieldWithDefault.String()))
				})

				type args4 struct {
					String types.String `gq:":String! = \"hello\""`
				}
				ts.AddFieldWithArgs(queryGQLType, `requiredArgDefault`, func(q *Query, a *args4) ts.Result[types.String] {
					return result.Of(a.String)
				})

				return &Query{
					FieldList: []types.String{types.NewString("c")},
				}
			},
		},
	}

	for _, tc := range testCases {
		mod := ts.NewModule()
		root := tc.schema(mod)
		tr, err := ts.NewTypeRegistry(ts.WithModule(mod))
		if err != nil {
			t.Fatal(err)
		}

		schema := tr.MustBuildSchema("Query")
		pq, err := querypkg.PrepareQuery(loadString(t, tc.name+".gql"), "", schema)
		if err != nil {
			t.Fatal(err)
		}
		result := pq.Execute(context.Background(), root, nil, nil)
		diff, msg := jsondiff.Compare([]byte(loadString(t, tc.name+".json")), result, &jsondiff.Options{
			Added:            jsondiff.Tag{Begin: "+"},
			Removed:          jsondiff.Tag{Begin: "-"},
			Changed:          jsondiff.Tag{Begin: "~"},
			ChangedSeparator: " => ",
			Indent:           "    ",
		})
		if diff != jsondiff.FullMatch {
			t.Fatal(msg)
		}
	}
}

func loadString(t *testing.T, name string) string {
	content, err := testcases.ReadFile("testcases/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func buildDefaultSchema(mod *ts.Module) any {
	type Enum string
	enumGQLType := ts.NewEnumType[Enum](mod, ``)
	enumValue1 := enumGQLType.Value(`VALUE1`)
	enumValue2 := enumGQLType.Value(`VALUE2`)
	_ = enumValue2

	type InputObject struct {
		ID          types.ID
		String      types.String
		Int         types.Int
		Float       types.Float
		Bool        types.Boolean
		InputObject *InputObject
		Custom      customScalar
	}
	ts.NewInputObjectType[InputObject](mod, ``)

	type Interface ts.Interface[any]
	interfaceGQLType := ts.NewInterfaceType[Interface](mod, `{
		field1: String
	}`)

	type Object1 struct {
		ID     types.ID
		String types.String
		Int    types.Int
		Float  types.Float
		Bool   types.Boolean
		Custom customScalar
		Nested *Object1
	}
	object1GQLType := ts.NewObjectType[Object1](mod, ``)
	interfaceFromObject1 := ts.Implements(object1GQLType, interfaceGQLType)
	_ = interfaceFromObject1

	ts.AddField(object1GQLType, `field1`, func(q *Object1) ts.Result[types.String] {
		return result.Of(types.NewString("o1 resolver"))
	})

	type Object2 struct {
		Field1 types.String
	}
	object2GQLType := ts.NewObjectType[Object2](mod, ``)
	interfaceFromObject2 := ts.Implements(object2GQLType, interfaceGQLType)

	ts.NewScalarType[customScalar](mod, ``)

	type Union ts.Union
	unionGQLType := ts.NewUnionType[Union](mod, ``)
	unionFromObject1 := ts.UnionMember(unionGQLType, object1GQLType)
	unionFromObject2 := ts.UnionMember(unionGQLType, object2GQLType)
	_ = unionFromObject2

	type Query struct{}
	queryGQLType := ts.NewObjectType[Query](mod, ``)

	ts.AddField(queryGQLType, `enum`, func(q *Query) ts.Result[Enum] {
		return result.Of(enumValue1)
	})

	type object1Args struct {
		InputObject *InputObject
	}
	ts.AddFieldWithArgs(queryGQLType, `object1`, func(q *Query, a *object1Args) ts.Result[*Object1] {
		var mapInput func(in *InputObject) *Object1
		mapInput = func(in *InputObject) *Object1 {
			v := &Object1{
				ID:     in.ID,
				String: in.String,
				Int:    in.Int,
				Float:  in.Float,
				Bool:   in.Bool,
				Custom: in.Custom,
			}
			if in.InputObject != nil {
				v.Nested = mapInput(in.InputObject)
			}
			return v
		}

		return result.Of(mapInput(a.InputObject))
	})

	ts.AddField(queryGQLType, `interface`, func(q *Query) ts.Result[Interface] {
		return result.Of(interfaceFromObject2(&Object2{
			Field1: types.NewString("o2 field"),
		}))
	})

	ts.AddField(queryGQLType, `union`, func(q *Query) ts.Result[Union] {
		return result.Of(unionFromObject1(&Object1{
			String: types.NewString("o1 field"),
		}))
	})

	return &Query{}
}
