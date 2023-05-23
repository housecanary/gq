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
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"

	"github.com/codemodus/kace"
	gqast "github.com/housecanary/gq/ast"
)

// Generate runs code generation
func ConvertToTS(
	types []string,
	dir string,
	packageNames []string,
) error {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedFiles | packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Dir:  dir,
	}, append([]string{
		"github.com/housecanary/gq/schema",
		"github.com/housecanary/gq/schema/structschema",
	}, packageNames...)...)

	if err != nil {
		return err
	}
	schemaPkg, ssPkg := pkgs[0].Types, pkgs[1].Types

	pkgs = pkgs[2:]
	c := &genCtx{
		schemaPkg,
		ssPkg,
		pkgs,
		"",
		"",
		"",
		make(map[string]gqlMeta),
	}

	for _, name := range types {
		if err := c.addTypeByName(name); err != nil {
			return err
		}
	}

	return c.convertToTS()
}

func (c *genCtx) convertToTS() error {

	allMetas := make([]gqlMeta, 0, len(c.meta))
	for _, v := range c.meta {
		allMetas = append(allMetas, v)
	}
	sort.Stable(sortMetaByName(allMetas))

	for _, gq := range allMetas {
		switch gq.Kind() {
		case typeKindEnum:
			rewriteEnum(gq, c)
		case typeKindInputObject:
			rewriteInputObject(gq, c)
		case typeKindInterface:
			rewriteInterface(gq, c)
		case typeKindObject:
			rewriteObject(gq, c)
		case typeKindScalar:
			rewriteScalar(gq, c)
		case typeKindUnion:
			rewriteUnion(gq, c)
		}
	}

	for _, pkg := range c.pkgs {
		for _, file := range pkg.Syntax {
			for _, imp := range file.Imports {
				if imp.Path.Value == `"github.com/housecanary/gq/schema/structschema"` {
					if imp.Name != nil && imp.Name.Name != "" {
						astutil.DeleteNamedImport(pkg.Fset, file, imp.Name.Name, "github.com/housecanary/gq/schema/structschema")
					} else {
						astutil.DeleteImport(pkg.Fset, file, "github.com/housecanary/gq/schema/structschema")
					}
				}
			}
			f, err := os.Create(pkg.Fset.File(file.Pos()).Name())
			if err != nil {
				return err
			}
			defer f.Close()
			format.Node(f, pkg.Fset, file)
		}
	}

	return nil
}

type rewriteInfo struct {
	typeMatcher matcher
	typeName    string
	pkg         *packages.Package
	file        *ast.File
}

func rewriteType(gq gqlMeta, c *genCtx, apply func(ri rewriteInfo) []transformRule) {
	typ := gq.NamedType()
	pkgPath := typ.Obj().Pkg().Path()
	updatePos := typ.Obj().Pos()
	for _, pkg := range c.pkgs {
		if pkg.ID != pkgPath {
			continue
		}

		for _, file := range pkg.Syntax {
			if file.FileStart <= updatePos && updatePos <= file.FileEnd {
				astutil.AddImport(pkg.Fset, file, "github.com/housecanary/gq/schema/ts")
				rules := apply(rewriteInfo{
					typeMatcher: matchAnyUntil(matchPosition(updatePos)),
					typeName:    typ.Obj().Name(),
					pkg:         pkg,
					file:        file,
				})
				transform(file, rules...)
				for _, other := range pkg.Syntax {
					if other == file {
						continue
					}
					transform(other, rules...)
				}
				break
			}
		}
	}
}

func rewriteEnum(gq gqlMeta, c *genCtx) {
	rewriteType(gq, c, func(ri rewriteInfo) []transformRule {
		return []transformRule{
			{
				matcher: ri.typeMatcher,
				action: func(c *astutil.Cursor) {
					c.Replace(&ast.TypeSpec{
						Name: &ast.Ident{
							Name: ri.typeName,
						},
						Type: &ast.Ident{
							Name: "string",
						},
					})
				},
			},

			{
				matcher: matchAnyUntil(match(func(n *ast.GenDecl) bool {
					r := test(n, ri.typeMatcher)
					return r
				})),
				action: func(c *astutil.Cursor) {
					var appendNodes []ast.Node

					enumTypeName := ri.typeName + "GQLType"

					etd := gq.(*enumMeta).GQL

					var enumGQL []string
					if etd.Description != "" {
						enumGQL = append(enumGQL, gqast.StringValue{V: etd.Description}.Representation())
					}

					if etd.Name != ri.typeName {
						enumGQL = append(enumGQL, etd.Name)
					}

					if len(etd.Directives) > 0 {
						var sb strings.Builder
						etd.Directives.MarshalGraphQL(&sb)
						enumGQL = append(enumGQL, sb.String())
					}

					appendNodes = append(appendNodes, &ast.GenDecl{
						Tok: token.VAR,
						Specs: []ast.Spec{
							&ast.ValueSpec{
								Names: []*ast.Ident{
									{
										Name: enumTypeName,
									},
								},
								Values: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.IndexExpr{
											X: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: "ts",
												},
												Sel: &ast.Ident{
													Name: "NewEnumType",
												},
											},
											Index: &ast.Ident{
												Name: ri.typeName,
											},
										},
										Args: []ast.Expr{
											&ast.Ident{
												Name: "GQLModule",
											},
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: "`" + strings.Join(enumGQL, " ") + "`",
											},
										},
									},
								},
							},
						},
					})

					var valueSpecs []ast.Spec
					for _, ev := range etd.EnumValueDefinitions {
						var enumValueGQL []string
						if ev.Description != "" {
							enumValueGQL = append(enumValueGQL, gqast.StringValue{V: ev.Description}.Representation())
						}

						enumValueGQL = append(enumValueGQL, ev.Value)

						if len(ev.Directives) > 0 {
							var sb strings.Builder
							ev.Directives.MarshalGraphQL(&sb)
							enumValueGQL = append(enumValueGQL, sb.String())
						}

						valueSpecs = append(valueSpecs, &ast.ValueSpec{
							Names: []*ast.Ident{
								{
									Name: ri.typeName + kace.Pascal(ev.Value),
								},
							},
							Values: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: enumTypeName,
										},
										Sel: &ast.Ident{
											Name: "Value",
										},
									},
									Args: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.STRING,
											Value: "`" + strings.Join(enumValueGQL, " ") + "`",
										},
									},
								},
							},
						})
					}

					if len(valueSpecs) > 0 {
						appendNodes = append(appendNodes, &ast.GenDecl{
							Tok:   token.VAR,
							Specs: valueSpecs,
						})
					}

					for i, j := 0, len(appendNodes)-1; i < j; i, j = i+1, j-1 {
						appendNodes[i], appendNodes[j] = appendNodes[j], appendNodes[i]
					}

					for _, n := range appendNodes {
						c.InsertAfter(n)
					}

				},
			},
		}
	})
}

func rewriteInputObject(gq gqlMeta, c *genCtx) {
	rewriteType(gq, c, func(ri rewriteInfo) []transformRule {
		return []transformRule{
			{
				matcher: sequenceOf(
					ri.typeMatcher,
					match[*ast.StructType](),
					match[*ast.FieldList](),
					matchField(ri.pkg.TypesInfo, "github.com/housecanary/gq/schema/structschema", "InputObject", true),
				),
				action: func(c *astutil.Cursor) {
					c.Delete()
				},
			},

			{
				matcher: matchAnyUntil(match(func(n *ast.GenDecl) bool {
					return test(n, ri.typeMatcher)
				})),
				action: func(c *astutil.Cursor) {
					inputObjectTypeName := ri.typeName + "GQLType"

					iod := gq.(*inputObjMeta).GQL

					var gql []string
					if iod.Description != "" {
						gql = append(gql, gqast.StringValue{V: iod.Description}.Representation())
					}

					if iod.Name != ri.typeName {
						gql = append(gql, iod.Name)
					}

					if len(iod.Directives) > 0 {
						var sb strings.Builder
						iod.Directives.MarshalGraphQL(&sb)
						gql = append(gql, sb.String())
					}

					c.InsertAfter(&ast.GenDecl{
						Tok: token.VAR,
						Specs: []ast.Spec{
							&ast.ValueSpec{
								Names: []*ast.Ident{
									{
										Name: inputObjectTypeName,
									},
								},
								Values: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.IndexExpr{
											X: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: "ts",
												},
												Sel: &ast.Ident{
													Name: "NewInputObjectType",
												},
											},
											Index: &ast.Ident{
												Name: ri.typeName,
											},
										},
										Args: []ast.Expr{
											&ast.Ident{
												Name: "GQLModule",
											},
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: "`" + strings.Join(gql, " ") + "`",
											},
										},
									},
								},
							},
						},
					})
				},
			},
		}
	})
}

func rewriteInterface(gq gqlMeta, c *genCtx) {
	rewriteType(gq, c, func(ri rewriteInfo) []transformRule {
		return []transformRule{
			{
				matcher: ri.typeMatcher,
				action: func(c *astutil.Cursor) {
					c.Replace(&ast.TypeSpec{
						Name: &ast.Ident{
							Name: ri.typeName,
						},
						Type: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "ts",
							},
							Sel: &ast.Ident{
								Name: "Interface",
							},
						},
					})
				},
			},

			{
				matcher: matchAnyUntil(match(func(n *ast.GenDecl) bool {
					return test(n, ri.typeMatcher)
				})),
				action: func(c *astutil.Cursor) {
					c.InsertAfter(&ast.GenDecl{
						Tok: token.VAR,
						Specs: []ast.Spec{
							&ast.ValueSpec{
								Names: []*ast.Ident{
									{
										Name: ri.typeName + "GQLType",
									},
								},
								Values: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.IndexExpr{
											X: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: "ts",
												},
												Sel: &ast.Ident{
													Name: "NewInterfaceType",
												},
											},
											Index: &ast.Ident{
												Name: ri.typeName,
											},
										},
										Args: []ast.Expr{
											&ast.Ident{
												Name: "GQLModule",
											},
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: "`" + gq.(*interfaceMeta).OriginalTag + "`",
											},
										},
									},
								},
							},
						},
					})
				},
			},
		}
	})
}

func rewriteObject(gq gqlMeta, c *genCtx) {
	rewriteType(gq, c, func(ri rewriteInfo) []transformRule {
		ot := gq.(*objMeta)

		var addNodes []ast.Node

		objectTypeName := ri.typeName + "GQLType"

		od := gq.(*objMeta).GQL

		var gql []string
		if od.Description != "" {
			gql = append(gql, gqast.StringValue{V: od.Description}.Representation())
		}

		if od.Name != ri.typeName {
			gql = append(gql, od.Name)
		}

		if len(od.Directives) > 0 {
			var sb strings.Builder
			od.Directives.MarshalGraphQL(&sb)
			gql = append(gql, sb.String())
		}

		addNodes = append(addNodes, &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						{
							Name: objectTypeName,
						},
					},
					Values: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.IndexExpr{
								X: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "ts",
									},
									Sel: &ast.Ident{
										Name: "NewObjectType",
									},
								},
								Index: &ast.Ident{
									Name: ri.typeName,
								},
							},
							Args: []ast.Expr{
								&ast.Ident{
									Name: "GQLModule",
								},
								&ast.BasicLit{
									Kind:  token.STRING,
									Value: "`" + strings.Join(gql, " ") + "`",
								},
							},
						},
					},
				},
			},
		})

		rules := []transformRule{
			{
				matcher: sequenceOf(
					ri.typeMatcher,
					match[*ast.StructType](),
					match[*ast.FieldList](),
					matchField(ri.pkg.TypesInfo, "github.com/housecanary/gq/schema/structschema", "Meta", true),
				),
				action: func(c *astutil.Cursor) {
					c.Delete()
				},
			},

			{
				matcher: matchAnyUntil(match(func(n *ast.GenDecl) bool {
					return test(n, ri.typeMatcher)
				})),
				action: func(c *astutil.Cursor) {
					for i, j := 0, len(addNodes)-1; i < j; i, j = i+1, j-1 {
						addNodes[i], addNodes[j] = addNodes[j], addNodes[i]
					}

					for _, n := range addNodes {
						c.InsertAfter(n)
					}
				},
			},
		}

		for _, meta := range c.meta {
			if meta.Kind() == typeKindInterface {
				intfType := fieldByName(meta.NamedType(), "Interface").field.Type()
				if types.AssignableTo(gq.NamedType(), intfType) {
					it := intfType.(*types.Interface)
					for i := 0; i < it.NumMethods(); i++ {
						fn := it.Method(i)
						rules = append(rules, transformRule{
							matcher: matchAnyUntil(match(func(fd *ast.FuncDecl) bool {
								if fd.Recv == nil {
									return false
								}
								if fd.Name.Name != fn.Name() {
									return false
								}
								if id, ok := fd.Recv.List[0].Type.(*ast.Ident); ok {
									if id.Name == ri.typeName {
										return true
									}
								} else if star, ok := fd.Recv.List[0].Type.(*ast.StarExpr); ok {
									if id, ok := star.X.(*ast.Ident); ok {
										if id.Name == ri.typeName {
											return true
										}
									}
								}
								return false
							})),
							action: func(c *astutil.Cursor) {
								c.Delete()
							},
						})
					}
					addNodes = append(addNodes, &ast.GenDecl{
						Tok: token.VAR,
						Specs: []ast.Spec{
							&ast.ValueSpec{
								Names: []*ast.Ident{
									{
										Name: meta.NamedType().Obj().Name() + "From" + kace.Pascal(ri.typeName),
									},
								},
								Values: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "ts",
											},
											Sel: &ast.Ident{
												Name: "Implements",
											},
										},
										Args: []ast.Expr{
											&ast.Ident{
												Name: ri.typeName + "GQLType",
											},
											&ast.Ident{
												Name: meta.NamedType().Obj().Name() + "GQLType",
											},
										},
									},
								},
							},
						},
					})
				}
			} else if meta.Kind() == typeKindUnion {
				unionType := fieldByName(meta.NamedType(), "Union").field.Type()
				if types.AssignableTo(gq.NamedType(), unionType) {
					it := unionType.(*types.Interface)
					for i := 0; i < it.NumMethods(); i++ {
						fn := it.Method(i)
						rules = append(rules, transformRule{
							matcher: matchAnyUntil(match(func(fd *ast.FuncDecl) bool {
								if fd.Recv == nil {
									return false
								}
								if fd.Name.Name != fn.Name() {
									return false
								}
								if id, ok := fd.Recv.List[0].Type.(*ast.Ident); ok {
									if id.Name == ri.typeName {
										return true
									}
								} else if star, ok := fd.Recv.List[0].Type.(*ast.StarExpr); ok {
									if id, ok := star.X.(*ast.Ident); ok {
										if id.Name == ri.typeName {
											return true
										}
									}
								}
								return false
							})),
							action: func(c *astutil.Cursor) {
								c.Delete()
							},
						})
					}
				}
			}
		}

		for _, f := range ot.Fields {
			f := f
			genctx := c
			if f.Method != nil && !f.FromEmbedded {
				toDelete := findNodeContaining[*ast.FuncDecl](ri.pkg, f.Method.Obj().Pos())
				rules = append(rules, transformRule{
					matcher: sequenceOf(
						match[*ast.File](),
						match(func(n *ast.FuncDecl) bool {
							return n == toDelete
						}),
					),
					action: func(c *astutil.Cursor) {
						nodes := buildFieldResolver(genctx, ri.pkg, ri.typeName, f)
						for _, n := range nodes {
							c.InsertBefore(n)
						}
						c.Delete()
					},
				})
			}
		}

		return rules
	})
}

func rewriteScalar(gq gqlMeta, c *genCtx) {
	rewriteType(gq, c, func(ri rewriteInfo) []transformRule {
		return []transformRule{
			{
				matcher: sequenceOf(
					ri.typeMatcher,
					match[*ast.StructType](),
					match[*ast.FieldList](),
					matchField(ri.pkg.TypesInfo, "github.com/housecanary/gq/schema/structschema", "Meta", true),
				),
				action: func(c *astutil.Cursor) {
					c.Delete()
				},
			},

			{
				matcher: matchAnyUntil(match(func(n *ast.GenDecl) bool {
					return test(n, ri.typeMatcher)
				})),
				action: func(c *astutil.Cursor) {
					scalarTypeName := ri.typeName + "GQLType"

					sd := gq.(*scalarMeta).GQL

					var gql []string
					if sd.Description != "" {
						gql = append(gql, gqast.StringValue{V: sd.Description}.Representation())
					}

					if sd.Name != ri.typeName {
						gql = append(gql, sd.Name)
					}

					if len(sd.Directives) > 0 {
						var sb strings.Builder
						sd.Directives.MarshalGraphQL(&sb)
						gql = append(gql, sb.String())
					}

					c.InsertAfter(&ast.GenDecl{
						Tok: token.VAR,
						Specs: []ast.Spec{
							&ast.ValueSpec{
								Names: []*ast.Ident{
									{
										Name: scalarTypeName,
									},
								},
								Values: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.IndexExpr{
											X: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: "ts",
												},
												Sel: &ast.Ident{
													Name: "NewScalarType",
												},
											},
											Index: &ast.Ident{
												Name: ri.typeName,
											},
										},
										Args: []ast.Expr{
											&ast.Ident{
												Name: "GQLModule",
											},
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: "`" + strings.Join(gql, " ") + "`",
											},
										},
									},
								},
							},
						},
					})
				},
			},
		}
	})
}

func rewriteUnion(gq gqlMeta, c *genCtx) {
	rewriteType(gq, c, func(ri rewriteInfo) []transformRule {
		addNodes := []ast.Node{
			&ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{
							{
								Name: ri.typeName + "GQLType",
							},
						},
						Values: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.IndexExpr{
									X: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "ts",
										},
										Sel: &ast.Ident{
											Name: "NewUnionType",
										},
									},
									Index: &ast.Ident{
										Name: ri.typeName,
									},
								},
								Args: []ast.Expr{
									&ast.Ident{
										Name: "GQLModule",
									},
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: "`" + gq.(*unionMeta).OriginalTag + "`",
									},
								},
							},
						},
					},
				},
			},
		}

		for _, meta := range c.meta {
			if meta.Kind() == typeKindObject {
				unionType := fieldByName(gq.NamedType(), "Union").field.Type()
				if types.AssignableTo(meta.NamedType(), unionType) {
					addNodes = append(addNodes, &ast.GenDecl{
						Tok: token.VAR,
						Specs: []ast.Spec{
							&ast.ValueSpec{
								Names: []*ast.Ident{
									{
										Name: ri.typeName + "From" + kace.Pascal(meta.NamedType().Obj().Name()),
									},
								},
								Values: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "ts",
											},
											Sel: &ast.Ident{
												Name: "UnionMember",
											},
										},
										Args: []ast.Expr{
											&ast.Ident{
												Name: ri.typeName + "GQLType",
											},
											&ast.Ident{
												Name: meta.NamedType().Obj().Name() + "GQLType",
											},
										},
									},
								},
							},
						},
					})
				}
			}
		}
		return []transformRule{
			{
				matcher: ri.typeMatcher,
				action: func(c *astutil.Cursor) {
					c.Replace(&ast.TypeSpec{
						Name: &ast.Ident{
							Name: ri.typeName,
						},
						Type: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "ts",
							},
							Sel: &ast.Ident{
								Name: "Union",
							},
						},
					})
				},
			},

			{
				matcher: matchAnyUntil(match(func(n *ast.GenDecl) bool {
					return test(n, ri.typeMatcher)
				})),
				action: func(c *astutil.Cursor) {
					for i, j := 0, len(addNodes)-1; i < j; i, j = i+1, j-1 {
						addNodes[i], addNodes[j] = addNodes[j], addNodes[i]
					}

					for _, n := range addNodes {
						c.InsertAfter(n)
					}
				},
			},
		}
	})
}
