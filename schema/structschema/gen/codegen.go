// Copyright 2018 HouseCanary, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gen

import (
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/codemodus/kace"
	j "github.com/dave/jennifer/jen"

	"github.com/housecanary/gq/ast"
)

type outputCtx struct {
	*genCtx
	meta              []gqlMeta
	packagePrefixes   map[string]string
	listWrappers      map[string]typeAndName
	injectors         map[string]typeAndName
	inputListCreators map[string]*inputListCreatorSpec
}

type typeAndName struct {
	typ  types.Type
	name string
}

type sortTypeAndName []typeAndName

func (s sortTypeAndName) Len() int {
	return len(s)
}

func (s sortTypeAndName) Less(i, j int) bool {
	return s[i].name < s[j].name
}

func (s sortTypeAndName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type inputListCreatorSpec struct {
	typ      types.Type
	name     string
	maxDepth int
}

type sortInputListCreatorSpec []*inputListCreatorSpec

func (s sortInputListCreatorSpec) Len() int {
	return len(s)
}

func (s sortInputListCreatorSpec) Less(i, j int) bool {
	return s[i].name < s[j].name
}

func (s sortInputListCreatorSpec) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func assertToPtr(g *j.Group, from, to string, castTo *j.Statement) {
	g.Var().Id(to).Op("*").Add(castTo.Clone())
	g.If(
		j.List(j.Id("t"), j.Id("ok")).Op(":=").Id(from).Assert(j.Op("*").Add(castTo.Clone())),
		j.Id("ok"),
	).Block(
		j.Id(to).Op("=").Id("t"),
	).Else().Block(
		j.Id("t").Op(":=").Id(from).Assert(castTo.Clone()),
		j.Id(to).Op("=").Op("&").Id("t"),
	)
}

func assertToTyp(g *j.Group, from, to string, castTo *j.Statement) {
	g.Var().Id(to).Add(castTo.Clone())
	g.If(
		j.List(j.Id("t"), j.Id("ok")).Op(":=").Id(from).Assert(castTo.Clone()),
		j.Id("ok"),
	).Block(
		j.Id(to).Op("=").Id("t"),
	).Else().Block(
		j.Id("t").Op(":=").Id(from).Assert(j.Op("*").Add(castTo.Clone())),
		j.Id(to).Op("=").Op("*").Id("t"),
	)
}

func (c *outputCtx) addEnumTypeRegistration(body *j.Group, meta *enumMeta) {
	qualTyp := c.packageType(meta.NamedType())

	encode := j.Func().Params(
		j.Id("ctx").Qual("context", "Context"),
		j.Id("v").Interface(),
	).Params(
		j.Qual("github.com/housecanary/gq/schema", "LiteralValue"),
		j.Error(),
	).BlockFunc(func(g *j.Group) {
		g.If(j.Id("v").Op("==").Nil()).Block(
			j.Return(j.Nil(), j.Nil()),
		)
		g.Id("enm").Op(":=").Id("v").Assert(j.Qual("github.com/housecanary/gq/schema/structschema", "EnumValue"))
		g.If(j.Id("enm").Dot("Nil").Call()).Block(
			j.Return(j.Nil(), j.Nil()),
		)
		g.Return(
			j.Qual("github.com/housecanary/gq/schema", "LiteralString").Call(
				j.Id("enm").Dot("String").Call(),
			),
			j.Nil(),
		)
	})

	decode := j.Func().Params(
		j.Id("ctx").Qual("context", "Context"),
		j.Id("v").Qual("github.com/housecanary/gq/schema", "LiteralValue"),
	).Params(
		j.Interface(),
		j.Error(),
	).Block(
		j.If(j.Id("v").Op("==").Nil()).Block(
			j.Return(qualTyp.Clone().Values(), j.Nil()),
		),
		j.List(j.Id("s"), j.Id("ok")).Op(":=").Id("v").Assert(j.Qual("github.com/housecanary/gq/schema", "LiteralString")),
		j.If(j.Op("!").Id("ok")).Block(
			j.Return(j.Nil(), j.Qual("fmt", "Errorf").Call(j.Lit("Invalid enum input: %v is not an string"), j.Id("v"))),
		),
		j.Return(qualTyp.Clone().Values(j.Dict{
			j.Id("Enum"): j.Qual("github.com/housecanary/gq/schema/structschema", "Enum").Call(j.Id("s")),
		}), j.Nil()),
	)

	inputListCreator := j.Id(c.inputListCreatorFor(meta.Name(), meta.NamedType())).Values()

	body.BlockFunc(func(reg *j.Group) {
		reg.Id("etb").Op(":=").Id("sb").Dot("AddEnumType").Call(
			j.Lit(meta.GQL.Name),
			encode,
			decode,
			inputListCreator,
		)

		c.addSetElementProps(reg, "etb", meta.GQL.Description, meta.GQL.Directives)
		for _, e := range meta.GQL.EnumValueDefinitions {
			reg.BlockFunc(func(g *j.Group) {
				g.Id("evb").Op(":=").Id("etb").Dot("AddValue").Call(j.Lit(e.Value))
				c.addSetElementProps(g, "evb", e.Description, e.Directives)
			})
		}
	})
}

func (c *outputCtx) addInterfaceTypeRegistration(body *j.Group, meta *interfaceMeta) {
	qualTyp := c.packageType(meta.NamedType())
	memberTypeSwitch := []*j.Statement{}
	for _, candidate := range c.meta {
		if ometa, ok := candidate.(*objMeta); ok {
			if types.AssignableTo(ometa.NamedType(), meta.InterfaceType) {
				memberTypeSwitch = append(memberTypeSwitch, j.Case(
					j.Op("*").Add(c.toTypeSignature(ometa.NamedType())),
				).Block(
					j.Return(j.Id("v"), j.Lit(ometa.GQL.Name)),
				))
			}
		}
	}
	body.BlockFunc(func(reg *j.Group) {
		reg.Id("itb").Op(":=").Id("sb").Dot("AddInterfaceType").Call(
			j.Lit(meta.Name()),
			j.Func().Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("val").Interface(),
			).Params(
				j.Interface(),
				j.String(),
			).BlockFunc(func(g *j.Group) {
				assertToTyp(g, "val", "intf", qualTyp)
				g.Id("ifaceVal").Op(":=").Id("intf").Dot("Interface")
				g.If(j.Id("ifaceVal").Op("==").Nil()).Block(
					j.Return(j.Nil(), j.Lit("")),
				)
				g.Switch().Id("v").Op(":=").Id("ifaceVal").Assert(j.Type()).BlockFunc(func(g *j.Group) {
					for _, e := range memberTypeSwitch {
						g.Add(e)
					}
				})
				g.Return(j.Nil(), j.Lit(""))
			}),
		)
		c.addSetElementProps(reg, "itb", meta.GQL.Description, meta.GQL.Directives)
		for _, f := range meta.GQL.FieldsDefinition {
			reg.BlockFunc(func(g *j.Group) {
				g.Id("fb").Op(":=").Id("itb").Dot("AddField").Call(j.Lit(f.Name), c.createAstType(f.Type))
				c.addSetElementProps(g, "fb", f.Description, f.Directives)
				for _, a := range f.ArgumentsDefinition {
					g.BlockFunc(func(g *j.Group) {
						g.Id("ab").Op(":=").Id("fb").Dot("AddArgument").Call(j.Lit(a.Name), c.createAstType(a.Type), c.createAstValue(a.DefaultValue))
						c.addSetElementProps(g, "ab", a.Description, a.Directives)
					})
				}
			})
		}
	})
}

func (c *outputCtx) addObjectTypeRegistration(body *j.Group, meta *objMeta) {
	qualTyp := c.packageType(meta.NamedType())
	body.BlockFunc(func(reg *j.Group) {
		reg.Id("otb").Op(":=").Id("sb").Dot("AddObjectType").Call(j.Lit(meta.Name()))
		c.addSetElementProps(reg, "otb", meta.GQL.Description, meta.GQL.Directives)
		sort.Stable(sortFieldMetasByName(meta.Fields))
		for _, f := range meta.Fields {
			c.addObjectField(reg, qualTyp, f)
		}

		for _, candidate := range c.meta {
			if intfMeta, ok := candidate.(*interfaceMeta); ok {
				if types.AssignableTo(meta.NamedType(), intfMeta.InterfaceType) {
					reg.Id("otb").Dot("Implements").Call(j.Lit(intfMeta.Name()))
				}
			}
		}
	})
}

func (c *outputCtx) addScalarTypeRegistration(body *j.Group, meta *scalarMeta) {
	qualTyp := c.packageType(meta.NamedType())

	encode := j.Qual("github.com/housecanary/gq/schema", "EncodeScalarMarshaler")

	decode := j.Func().Params(
		j.Id("ctx").Qual("context", "Context"),
		j.Id("v").Qual("github.com/housecanary/gq/schema", "LiteralValue"),
	).Params(
		j.Interface(),
		j.Error(),
	).Block(
		j.Id("val").Op(":=").Add(qualTyp).Values(),
		j.Id("err").Op(":=").Id("val").Dot("FromLiteralValue").Call(j.Id("v")),
		j.If(j.Id("err").Op("!=").Nil()).Block(
			j.Return(j.Nil(), j.Id("err")),
		),
		j.Return(j.Id("val"), j.Nil()),
	)

	inputListCreator := j.Id(c.inputListCreatorFor(meta.Name(), meta.NamedType())).Values()

	body.BlockFunc(func(reg *j.Group) {
		reg.Id("scb").Op(":=").Id("sb").Dot("AddScalarType").Call(
			j.Lit(meta.Name()),
			encode,
			decode,
			inputListCreator,
		)

		c.addSetElementProps(reg, "scb", meta.GQL.Description, meta.GQL.Directives)
	})
}

func (c *outputCtx) addUnionTypeRegistration(body *j.Group, meta *unionMeta) {
	qualTyp := c.packageType(meta.NamedType())
	memberTypeSwitch := []*j.Statement{}
	members := j.Index().String().ValuesFunc(func(g *j.Group) {
		for _, candidate := range c.meta {
			if objMeta, ok := candidate.(*objMeta); ok {
				if types.AssignableTo(objMeta.NamedType(), meta.UnionType) {
					g.Lit(objMeta.GQL.Name)
					memberTypeSwitch = append(memberTypeSwitch, j.Case(
						j.Op("*").Add(c.toTypeSignature(objMeta.NamedType())),
					).Block(
						j.Return(j.Id("v"), j.Lit(objMeta.GQL.Name)),
					))
				}
			}
		}
	})

	body.BlockFunc(func(reg *j.Group) {
		reg.Id("utb").Op(":=").Id("sb").Dot("AddUnionType").Call(
			j.Lit(meta.Name()),
			members,
			j.Func().Params(
				j.Id("ctx").Qual("context", "Context"),
				j.Id("val").Interface(),
			).Params(
				j.Interface(),
				j.String(),
			).BlockFunc(func(g *j.Group) {
				assertToTyp(g, "val", "u", qualTyp)
				g.Id("uVal").Op(":=").Id("u").Dot("Union")
				g.If(j.Id("uVal").Op("==").Nil()).Block(
					j.Return(j.Nil(), j.Lit("")),
				)
				g.Switch().Id("v").Op(":=").Id("uVal").Assert(j.Type()).BlockFunc(func(g *j.Group) {
					for _, e := range memberTypeSwitch {
						g.Add(e)
					}
				})
				g.Return(j.Nil(), j.Lit(""))
			}),
		)
		c.addSetElementProps(reg, "utb", meta.GQL.Description, meta.GQL.Directives)
	})
}

func (c *outputCtx) addInputObjectTypeRegistration(body *j.Group, meta *inputObjMeta) {
	qualTyp := c.packageType(meta.NamedType())

	decode := j.Func().Params(
		j.Id("ctx").Qual("github.com/housecanary/gq/schema", "InputObjectDecodeContext"),
	).Params(
		j.Interface(),
		j.Error(),
	).BlockFunc(func(g *j.Group) {
		g.If(j.Id("ctx").Dot("IsNil").Call()).Block(
			j.Return(j.Nil(), j.Nil()),
		)

		g.Id("val").Op(":=").Op("&").Add(qualTyp.Clone().Values())

		sort.Stable(sortInputFieldMetasByName(meta.Fields))
		for _, f := range meta.Fields {
			c.ensureInputListDepth(f.GQL.Type, 0)
			g.Block(
				j.List(j.Id("fieldVal"), j.Id("err")).Op(":=").Id("ctx").Dot("GetFieldValue").Call(j.Lit(f.GQL.Name)),
				j.If(j.Id("err").Op("!=").Nil()).Block(
					j.Return(j.Nil(), j.Id("err")),
				),
				j.If(j.Id("fieldVal").Op("!=").Nil()).Block(
					j.Id("val").Dot(f.StructName).Op("=").Id("fieldVal").Assert(c.toTypeSignature(f.Type)),
				),
			)
		}

		if types.AssignableTo(meta.NamedType(), c.ssPkg.Scope().Lookup("validatable").Type()) {
			g.If(j.Id("err").Op(":=").Id("val").Dot("Validate").Call(), j.Id("err").Op("!=").Nil()).Block(
				j.Return(j.Nil(), j.Id("err")),
			)
		}

		g.Return(j.Id("val"), j.Nil())
	})

	inputListCreator := j.Id(c.inputListCreatorFor(meta.Name(), types.NewPointer(meta.NamedType()))).Values()

	body.BlockFunc(func(reg *j.Group) {
		reg.Id("iob").Op(":=").Id("sb").Dot("AddInputObjectType").Call(j.Lit(meta.Name()), decode, inputListCreator)
		c.addSetElementProps(reg, "iob", meta.GQL.Description, meta.GQL.Directives)
		for _, f := range meta.Fields {
			reg.BlockFunc(func(g *j.Group) {
				g.Id("fb").Op(":=").Id("iob").Dot("AddField").Call(j.Lit(f.GQL.Name), c.createAstType(f.GQL.Type), c.createAstValue(f.GQL.DefaultValue))
				c.addSetElementProps(g, "fb", f.GQL.Description, f.GQL.Directives)
			})
		}
	})
}

func (c *outputCtx) createResolver(objTyp *j.Statement, meta *fieldMeta) *j.Statement {
	if meta.Field != nil {
		return c.createFieldResolver(objTyp, meta)
	} else if meta.Method != nil {
		return c.createMethodResolver(objTyp, meta)
	} else {
		panic(fmt.Errorf("Type %s, field %s, resolver is not a field or method", meta.Obj.Name(), meta.Name))
	}
}

func (c *outputCtx) createFieldResolver(objTyp *j.Statement, meta *fieldMeta) *j.Statement {
	fv := j.Id("v").Assert(j.Op("*").Add(objTyp)).Dot(meta.Field.Name())
	return j.Qual("github.com/housecanary/gq/schema", "MarkSafe").Call(
		j.Qual("github.com/housecanary/gq/schema", "SimpleResolver").Call(
			j.Func().Params(j.Id("v").Interface()).Params(j.Interface(), j.Id("error")).BlockFunc(func(g *j.Group) {
				g.If(j.Id("v").Op("==").Nil().Op("||").Id("v").Assert(j.Op("*").Add(objTyp)).Op("==").Nil()).Block(
					j.Return(j.Nil(), j.Nil()),
				)

				if _, ok := meta.Field.Type().(*types.Pointer); ok {
					g.If(fv).Op("==").Nil().Block(
						j.Return(j.Nil(), j.Nil()),
					)
				}
				g.ReturnFunc(func(g *j.Group) {
					switch t := meta.Field.Type().(type) {
					case *types.Slice:
						g.Add(c.prepReturnValue(fv, t))
					case *types.Pointer:
						g.Add(fv)
					default:
						g.Op("&").Add(fv)
					}
					g.Nil()
				})
			}),
		),
	)
}

func (c *outputCtx) createMethodResolver(objTyp *j.Statement, meta *fieldMeta) *j.Statement {
	resolverLevel := 0

	extracters := []*j.Statement{}

	sig := meta.Method.Type().(*types.Signature)
	parms := sig.Params()
	defIndex := 0
	for i := 0; i < parms.Len(); i++ {
		argName := fmt.Sprintf("arg%d", i)
		parm := parms.At(i)
		if c.isContextType(parm.Type()) {
			if resolverLevel == 0 {
				resolverLevel = 1
			}
			extracters = append(extracters, j.Id(argName).Op(":=").Id("ctx"))
		} else if c.isResolverContextType(parm.Type()) {
			resolverLevel = 2
			extracters = append(extracters, j.Id(argName).Op(":=").Id("ctx"))
		} else if c.isInjectedArg(parm) {
			if resolverLevel == 0 {
				resolverLevel = 1
			}

			extracters = append(extracters, j.Id(argName).Op(":=").Id(c.injectorName(parm)).Call(j.Id("ctx")))
		} else {
			resolverLevel = 2
			def := meta.GQL.ArgumentsDefinition[defIndex]
			defIndex++
			extracters = append(extracters, j.Var().Id(argName).Add(c.toTypeSignature(parm.Type())))
			extracters = append(extracters, j.BlockFunc(func(g *j.Group) {
				g.List(j.Id("raw"), j.Id("err")).Op(":=").Id("ctx").Dot("GetArgumentValue").Call(j.Lit(def.Name))
				g.If(j.Id("err").Op("!=").Nil()).Block(
					j.Return(j.Nil(), j.Id("err")),
				)
				g.If(j.Id("raw").Op("!=").Nil()).Block(
					j.Id(argName).Op("=").Id("raw").Assert(c.toTypeSignature(parm.Type())),
				)
			}))
		}
	}

	resolverFunc := j.Func().ParamsFunc(func(g *j.Group) {
		switch resolverLevel {
		case 1:
			g.Id("ctx").Qual("context", "Context")
		case 2:
			g.Id("ctx").Qual("github.com/housecanary/gq/schema", "ResolverContext")
		}

		g.Id("v").Interface()
	}).Params(j.Interface(), j.Id("error")).BlockFunc(func(g *j.Group) {
		g.If(j.Id("v").Op("==").Nil()).Block(
			j.Return(j.Nil(), j.Nil()),
		)
		for _, e := range extracters {
			g.Add(e)
		}

		var resultAssign *j.Statement
		if sig.Results().Len() == 1 {
			resultAssign = j.Id("result")
		} else {
			resultAssign = j.List(j.Id("result"), j.Id("errResult"))
		}
		g.Add(resultAssign).Op(":=").Id("v").Assert(j.Op("*").Add(objTyp)).Dot(meta.Method.Obj().Name()).CallFunc(func(g *j.Group) {
			for i := 0; i < parms.Len(); i++ {
				argName := fmt.Sprintf("arg%d", i)
				g.Id(argName)
			}
		})
		c.addResultHandler(g, sig)
	})

	switch resolverLevel {
	case 0:
		return j.Qual("github.com/housecanary/gq/schema", "SimpleResolver").Call(resolverFunc)
	case 1:
		return j.Qual("github.com/housecanary/gq/schema", "ContextResolver").Call(resolverFunc)
	case 2:
		return j.Qual("github.com/housecanary/gq/schema", "FullResolver").Call(resolverFunc)
	default:
		panic("Invalid resolver level")
	}
}

func (c *outputCtx) prepReturnValue(s *j.Statement, typ types.Type) *j.Statement {
	if slc, ok := typ.Underlying().(*types.Slice); ok {
		return j.Id(c.listWrapperFor(slc.Elem())).Call(s)
	}
	return s
}

func asyncValueFunc(cb func(g *j.Group)) *j.Statement {
	return j.Qual("github.com/housecanary/gq/schema", "AsyncValueFunc").Call(
		j.Func().Params(
			j.Id("ctx").Qual("context", "Context"),
		).Params(
			j.Interface(),
			j.Error(),
		).BlockFunc(cb),
	)
}

func (c *outputCtx) addResultHandler(g *j.Group, sig *types.Signature) {
	results := sig.Results()
	if results.Len() == 1 {
		if fun, ok := results.At(0).Type().Underlying().(*types.Signature); ok {
			g.Return(asyncValueFunc(func(g *j.Group) {
				var resultAssign *j.Statement
				if fun.Results().Len() == 1 {
					resultAssign = j.Id("result")
				} else {
					resultAssign = j.List(j.Id("result"), j.Id("errResult"))
				}
				g.Add(resultAssign).Op(":=").Id("result").Call()
				c.addResultHandler(g, fun)
			}), j.Nil())
		} else if ch, ok := results.At(0).Type().Underlying().(*types.Chan); ok {
			g.Return(asyncValueFunc(func(g *j.Group) {
				g.List(j.Id("result"), j.Id("ok")).Op(":=").Op("<-").Id("result")
				g.If(j.Op("!").Id("ok")).Block(
					j.Return(
						j.Nil(), j.Qual("fmt", "Errorf").Call(j.Lit("Channel receive failed, closed prematurely")),
					),
				)
				g.If(j.Id("result").Op("==").Nil()).Block(
					j.Return(j.Nil(), j.Nil()),
				)
				g.Return(c.prepReturnValue(j.Id("result"), ch.Elem()), j.Nil())
			}), j.Nil())
		} else {
			if _, ok := results.At(0).Type().(*types.Pointer); ok {
				g.If(j.Id("result").Op("==").Nil()).Block(
					j.Return(j.Nil(), j.Nil()),
				)
			}
			g.Return(c.prepReturnValue(j.Id("result"), results.At(0).Type()), j.Nil())
		}
	} else {
		if ch, ok := results.At(0).Type().Underlying().(*types.Chan); ok {
			g.Return(asyncValueFunc(func(g *j.Group) {
				g.Var().Id("rc").Op("<-").Chan().Add(c.toTypeSignature(ch.Elem())).Op("=").Id("result")
				g.Var().Id("ec").Op("<-").Chan().Error().Op("=").Id("errResult")
				g.Var().Id("r").Add(c.toTypeSignature(ch.Elem()))
				g.Var().Id("e").Error()

				g.Id("loop").Op(":").For().BlockFunc(func(g *j.Group) {
					g.Var().Id("ok").Bool()
					g.Select().Block(
						j.Case(j.List(j.Id("r"), j.Id("ok")).Op("=").Op("<-").Id("rc")).BlockFunc(func(g *j.Group) {
							g.If(j.Op("!").Id("ok")).Block(
								j.Id("rc").Op("=").Nil(),
								j.If(j.Id("ec").Op("==").Nil()).Block(
									j.Return(j.Nil(), j.Qual("fmt", "Errorf").Call(j.Lit("Channel receive failed: error closed and result closed"))),
								),
							).Else().Block(
								j.Break().Id("loop"),
							)
						}),
						j.Case(j.List(j.Id("e"), j.Id("ok")).Op("=").Op("<-").Id("ec")).BlockFunc(func(g *j.Group) {
							g.If(j.Op("!").Id("ok")).Block(
								j.Id("ec").Op("=").Nil(),
								j.If(j.Id("rc").Op("==").Nil()).Block(
									j.Return(j.Nil(), j.Qual("fmt", "Errorf").Call(j.Lit("Channel receive failed: result closed and error closed"))),
								),
							).Else().Block(
								j.Break().Id("loop"),
							)
						}),
					)
				})

				g.If(j.Id("e").Op("!=").Nil()).Block(
					j.Return(j.Nil(), j.Id("e")),
				)

				g.Return(c.prepReturnValue(j.Id("r"), ch.Elem()), j.Nil())
			}), j.Nil())
		} else {
			if _, ok := results.At(0).Type().(*types.Pointer); ok {
				g.If(j.Id("result").Op("==").Nil()).Block(
					j.Return(c.prepReturnValue(j.Nil(), results.At(0).Type()), j.Id("errResult")),
				)
			}
			g.Return(c.prepReturnValue(j.Id("result"), results.At(0).Type()), j.Id("errResult"))
		}
	}
}

func (c *outputCtx) addObjectField(reg *j.Group, objTyp *j.Statement, meta *fieldMeta) {
	reg.BlockFunc(func(g *j.Group) {
		g.Id("fb").Op(":=").Id("otb").Dot("AddField").Call(j.Lit(meta.GQL.Name), c.createAstType(meta.GQL.Type), c.createResolver(objTyp, meta))
		c.addSetElementProps(g, "fb", meta.GQL.Description, meta.GQL.Directives)
		for _, a := range meta.GQL.ArgumentsDefinition {
			c.ensureInputListDepth(a.Type, 0)
			g.BlockFunc(func(g *j.Group) {
				g.Id("ab").Op(":=").Id("fb").Dot("AddArgument").Call(j.Lit(a.Name), c.createAstType(a.Type), c.createAstValue(a.DefaultValue))
				c.addSetElementProps(g, "ab", a.Description, a.Directives)
			})
		}
	})
}

func (c *outputCtx) createAstType(typ ast.Type) *j.Statement {
	switch t := typ.(type) {
	case *ast.SimpleType:
		return j.Op("&").Qual("github.com/housecanary/gq/ast", "SimpleType").Values(j.Dict{
			j.Id("Name"): j.Lit(t.Name),
		})
	case *ast.ListType:
		return j.Op("&").Qual("github.com/housecanary/gq/ast", "ListType").Values(j.Dict{
			j.Id("Of"): c.createAstType(t.Of),
		})
	case *ast.NotNilType:
		return j.Op("&").Qual("github.com/housecanary/gq/ast", "NotNilType").Values(j.Dict{
			j.Id("Of"): c.createAstType(t.Of),
		})
	}
	panic("Unknown type")
}

func (c *outputCtx) createAstValue(val ast.Value) *j.Statement {
	if val == nil {
		return j.Nil()
	}
	switch v := val.(type) {
	case ast.StringValue:
		return j.Qual("github.com/housecanary/gq/ast", "StringValue").Values(j.Dict{
			j.Id("V"): j.Lit(v.V),
		})
	case ast.IntValue:
		return j.Qual("github.com/housecanary/gq/ast", "IntValue").Values(j.Dict{
			j.Id("V"): j.Lit(v.V),
		})
	case ast.FloatValue:
		return j.Qual("github.com/housecanary/gq/ast", "FloatValue").Values(j.Dict{
			j.Id("V"): j.Lit(v.V),
		})
	case ast.BooleanValue:
		return j.Qual("github.com/housecanary/gq/ast", "BooleanValue").Values(j.Dict{
			j.Id("V"): j.Lit(v.V),
		})
	case ast.NilValue:
		return j.Qual("github.com/housecanary/gq/ast", "NilValue").Values()
	case ast.EnumValue:
		return j.Qual("github.com/housecanary/gq/ast", "EnumValue").Values(j.Dict{
			j.Id("V"): j.Lit(v.V),
		})
	case ast.ArrayValue:
		return j.Qual("github.com/housecanary/gq/ast", "ArrayValue").ValuesFunc(func(g *j.Group) {
			for _, e := range v.V {
				g.Add(c.createAstValue(e))
			}
		})
	case ast.ObjectValue:
		args := j.Dict{}

		for k, v := range v.V {
			args[j.Id(k)] = c.createAstValue(v)
		}
		return j.Qual("github.com/housecanary/gq/ast", "ObjectValue").Values(args)
	case ast.ReferenceValue:
		return j.Qual("github.com/housecanary/gq/ast", "ReferenceValue").Values(j.Dict{
			j.Id("Name"): j.Lit(v.Name),
		})
	}
	panic(fmt.Errorf("Unknown ast value %v", val))
}

func (c *outputCtx) addSetElementProps(g *j.Group, name string, desc string, directives ast.Directives) {
	g.Id(name).Dot("SetDescription").Call(j.Lit(desc))
	for _, d := range directives {
		g.BlockFunc(func(g *j.Group) {
			if len(d.Arguments) > 0 {
				g.Id("db").Op(":=").Id(name).Dot("AddDirective").Call(j.Lit(d.Name))
			} else {
				g.Id(name).Dot("AddDirective").Call(j.Lit(d.Name))
			}
			for _, a := range d.Arguments {
				g.Id("db").Dot("AddArgument").Call(j.Lit(a.Name), c.createAstValue(a.Value))
			}
		})
	}
}

func (c *outputCtx) listWrapperFor(typ types.Type) string {
	sig := c.toMangledTypeName(typ)
	if existing, ok := c.listWrappers[sig]; ok {
		return existing.name
	}

	name := fmt.Sprintf("listWrapper%s", kace.Pascal(sig))
	c.listWrappers[sig] = typeAndName{typ, name}
	return name
}

func (c *outputCtx) inputListCreatorFor(schemaName string, typ types.Type) string {
	if existing, ok := c.inputListCreators[schemaName]; ok {
		existing.typ = typ
		return existing.name
	}

	name := fmt.Sprintf("inputListCreator%s", kace.Pascal(schemaName))
	c.inputListCreators[schemaName] = &inputListCreatorSpec{name: name, typ: typ, maxDepth: 0}
	return name
}

func (c *outputCtx) ensureInputListDepth(typ ast.Type, depth int) {
	switch v := typ.(type) {
	case *ast.SimpleType:
		if existing, ok := c.inputListCreators[v.Name]; ok {
			if depth > existing.maxDepth {
				existing.maxDepth = depth
			}
		} else {
			name := fmt.Sprintf("inputListCreator%s", kace.Pascal(v.Name))
			c.inputListCreators[v.Name] = &inputListCreatorSpec{name: name, maxDepth: depth}
		}
	case *ast.ListType:
		c.ensureInputListDepth(v.Of, depth+1)
	case *ast.NotNilType:
		c.ensureInputListDepth(v.Of, depth+1)
	}
}

func (c *outputCtx) addListWrapperType(typ types.Type, name string, g *j.Group) {
	g.Type().Id(name).Index().Add(c.toTypeSignature(typ))
	g.Func().Params(j.Id("w").Id(name)).Id("Len").Params().Int().BlockFunc(func(g *j.Group) {
		g.Return(j.Len(j.Id("w")))
	})

	g.Func().Params(j.Id("w").Id(name)).Id("ForEachElement").Params(j.Id("cb").Qual("github.com/housecanary/gq/schema", "ListValueCallback")).BlockFunc(func(g *j.Group) {
		g.For(j.List(j.Id("_"), j.Id("e")).Op(":=").Range().Id("w")).BlockFunc(func(g *j.Group) {
			if _, ok := typ.(*types.Pointer); ok {
				g.If(j.Id("e").Op("==").Nil()).Block(
					j.Id("cb").Call(j.Nil()),
				).Else().Block(
					j.Id("cb").Call(j.Id("e")),
				)
			} else {
				g.Id("cb").Call(j.Id("e"))
			}
		})
	})
}

func (c *outputCtx) addInputListCreatorType(name string, typ types.Type, nextName string, g *j.Group) {
	g.Type().Id(name).Struct()
	g.Func().Params(j.Id(name)).Id("NewList").Params(
		j.Id("size").Int(),
		j.Id("get").Func().Params(
			j.Id("i").Int(),
		).Params(
			j.Interface(),
			j.Error(),
		),
	).Params(
		j.Interface(),
		j.Error(),
	).Block(
		j.Id("lst").Op(":=").Make(j.Index().Add(c.toTypeSignature(typ)), j.Id("size")),
		j.For(j.Id("i").Op(":=").Lit(0), j.Id("i").Op("<").Id("size"), j.Id("i").Op("++")).Block(
			j.List(j.Id("v"), j.Id("err")).Op(":=").Id("get").Call(j.Id("i")),
			j.If(j.Id("err").Op("!=").Nil()).Block(
				j.Return(j.List(j.Nil(), j.Id("err"))),
			),
			j.Id("lst").Index(j.Id("i")).Op("=").Id("v").Assert(c.toTypeSignature(typ)),
		),
		j.Return(j.Id("lst"), j.Nil()),
	)

	if nextName != "" {
		g.Func().Params(j.Id(name)).Id("Creator").Params().Params(
			j.Qual("github.com/housecanary/gq/schema", "InputListCreator"),
		).Block(
			j.Return(j.Id(nextName).Values()),
		)
	} else {
		g.Func().Params(j.Id(name)).Id("Creator").Params().Params(
			j.Qual("github.com/housecanary/gq/schema", "InputListCreator"),
		).Block(
			j.Panic(j.Lit("Unreachable code - static analysis of the schema indicates that this level of list wrapping cannot be reached")),
		)
	}
}

func (c *outputCtx) injectorName(v *types.Var) string {
	typ := v.Type()
	sig := c.toMangledTypeName(typ)
	if existing, ok := c.injectors[sig]; ok {
		return existing.name
	}

	name := fmt.Sprintf("create%s", kace.Pascal(sig))
	c.injectors[sig] = typeAndName{typ, name}
	return name
}

func (c *outputCtx) packageType(typ *types.Named) *j.Statement {
	if strings.HasSuffix(typ.Obj().Pkg().Path(), c.outputPath) {
		return j.Id(typ.Obj().Name())
	}
	return j.Qual(typ.Obj().Pkg().Path(), typ.Obj().Name())
}

func (c *outputCtx) toTypeSignature(typ types.Type) j.Code {
	var castType j.Code
	switch t := typ.(type) {
	case *types.Named:
		castType = c.packageType(t)
	case *types.Pointer:
		castType = j.Op("*").Add(c.toTypeSignature(t.Elem()))
	case *types.Basic:
		castType = j.Id(t.Name())
	case *types.Array:
		castType = j.Index(j.Lit(t.Len())).Add(c.toTypeSignature(t.Elem()))
	case *types.Slice:
		castType = j.Index().Add(c.toTypeSignature(t.Elem()))
	case *types.Struct:
		castType = j.StructFunc(func(g *j.Group) {
			for i := 0; i < t.NumFields(); i++ {
				f := t.Field(i)
				g.Id(f.Name()).Add(c.toTypeSignature(f.Type()))
			}
		})
	case *types.Interface:
		castType = j.Interface() // Anon interface with methods is not supported
	case *types.Map:
		castType = j.Map(c.toTypeSignature(t.Key())).Add(c.toTypeSignature(t.Elem()))
	case *types.Chan:
		castType = j.Chan().Add(c.toTypeSignature(t.Elem()))
	}
	return castType
}

func (c *outputCtx) toMangledTypeName(typ types.Type) string {
	var ret string
	switch t := typ.(type) {
	case *types.Named:
		ret = fmt.Sprintf("%s%s", c.prefixForPackage(t.Obj().Pkg().Path()), t.Obj().Name())
	case *types.Pointer:
		ret = "ptrTo" + kace.Pascal(c.toMangledTypeName(t.Elem()))
	case *types.Basic:
		ret = "_" + t.Name()
	case *types.Array:
		ret = fmt.Sprintf("array%vOf%s", t.Len(), kace.Pascal(c.toMangledTypeName(t.Elem())))
	case *types.Slice:
		ret = fmt.Sprintf("sliceOf%s", kace.Pascal(c.toMangledTypeName(t.Elem())))
	default:
		panic(fmt.Errorf("Unsupported type to mangle: %v", typ))
	}
	return ret
}

func (c *outputCtx) prefixForPackage(path string) string {
	if alias, ok := c.packagePrefixes[path]; ok {
		return alias
	}
	alias := path

	if strings.HasSuffix(alias, "/") {
		// training slashes are usually tolerated, so we can get rid of one if
		// it exists
		alias = alias[:len(alias)-1]
	}

	if strings.Contains(alias, "/") {
		// if the path contains a "/", use the last part
		alias = alias[strings.LastIndex(alias, "/")+1:]
	}

	// alias should be lower case
	alias = strings.ToLower(alias)

	// alias should now only contain alphanumerics
	importsRegex := regexp.MustCompile(`[^a-z0-9]`)
	alias = importsRegex.ReplaceAllString(alias, "")

outer:
	for i := 0; ; i++ {
		var uniqued string
		if i == 0 {
			uniqued = alias
		} else {
			uniqued = fmt.Sprintf("%s%v", alias, i)
		}
		for _, v := range c.packagePrefixes {
			if v == uniqued {
				continue outer
			}
		}
		c.packagePrefixes[path] = uniqued
		return uniqued

	}
}

func (c *genCtx) createOutput() error {

	allMetas := make([]gqlMeta, 0, len(c.meta))
	for _, v := range c.meta {
		allMetas = append(allMetas, v)
	}
	sort.Stable(sortMetaByName(allMetas))

	oc := &outputCtx{c, allMetas, make(map[string]string), make(map[string]typeAndName), make(map[string]typeAndName), make(map[string]*inputListCreatorSpec)}
	f := j.NewFile(c.outputPackageName)

	f.HeaderComment("// Code generated by GQ DO NOT EDIT.")
	f.HeaderComment("// +build use_generated")

	var builderParamsGroup *j.Group
	f.Func().Id("NewSchemaBuilder").ParamsFunc(func(g *j.Group) {
		builderParamsGroup = g
	}).Params(j.Op("*").Qual("github.com/housecanary/gq/schema", "Builder"), j.Error()).BlockFunc(func(g *j.Group) {
		g.Id("sb").Op(":=").Qual("github.com/housecanary/gq/schema", "NewBuilder").Call()

		for _, v := range allMetas {
			switch t := v.(type) {
			case *objMeta:
				oc.addObjectTypeRegistration(g, t)
			case *enumMeta:
				oc.addEnumTypeRegistration(g, t)
			case *interfaceMeta:
				oc.addInterfaceTypeRegistration(g, t)
			case *unionMeta:
				oc.addUnionTypeRegistration(g, t)
			case *scalarMeta:
				oc.addScalarTypeRegistration(g, t)
			case *inputObjMeta:
				oc.addInputObjectTypeRegistration(g, t)
			default:
				panic("Invalid meta type")
			}
		}

		g.Return(j.Id("sb"), j.Nil())
	})

	allListWrappers := make(sortTypeAndName, 0)
	for _, v := range oc.listWrappers {
		allListWrappers = append(allListWrappers, v)
	}
	sort.Stable(allListWrappers)
	for _, e := range allListWrappers {
		oc.addListWrapperType(e.typ, e.name, f.Group)
	}

	allInjectors := make(sortTypeAndName, 0)
	for _, v := range oc.injectors {
		allInjectors = append(allInjectors, v)
	}
	sort.Stable(allInjectors)
	for _, e := range allInjectors {
		builderParamsGroup.Add(
			j.Line().Id(e.name).Func().Params(j.Qual("context", "Context")).Add(oc.toTypeSignature(e.typ)),
		)
	}

	allInputListCreators := make(sortInputListCreatorSpec, 0)
	for _, v := range oc.inputListCreators {
		allInputListCreators = append(allInputListCreators, v)
	}
	sort.Stable(allInputListCreators)

	for _, e := range allInputListCreators {
		typ := e.typ
		for i := 0; i <= e.maxDepth; i++ {
			var name string
			var nextName string
			if i == 0 {
				name = e.name
			} else {
				name = fmt.Sprintf("%s_%v", e.name, i)
			}
			if i < e.maxDepth {
				nextName = fmt.Sprintf("%s_%v", e.name, i+1)
			}
			oc.addInputListCreatorType(name, typ, nextName, f.Group)
			typ = types.NewSlice(typ)
		}
	}

	err := os.MkdirAll(c.outputPath, 0777)
	if err != nil {
		return err
	}
	of, err := os.OpenFile(filepath.Join(c.outputPath, c.outputFileName), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0660)
	if err != nil {
		return err
	}
	defer func() { of.Close() }()
	return f.Render(of)
}
