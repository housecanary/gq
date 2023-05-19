package gen

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/codemodus/kace"
	gqast "github.com/housecanary/gq/ast"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

func findNodeContaining[T ast.Node](pkg *packages.Package, pos token.Pos) T {
	matched := false
	var stack []T
	for _, file := range pkg.Syntax {
		if file.FileStart <= pos && pos <= file.FileEnd {
			astutil.Apply(file, func(c *astutil.Cursor) bool {
				n := c.Node()
				if n == nil {
					return true
				}

				if t, ok := n.(T); ok {
					stack = append(stack, t)
				}

				if n.Pos() == pos {
					matched = true
				}

				return true
			}, func(c *astutil.Cursor) bool {
				if matched {
					return false
				}
				if _, ok := c.Node().(T); ok {
					stack = stack[:len(stack)-1]
				}
				return true
			})
			break
		}
	}

	if matched {
		return stack[len(stack)-1]
	}

	var empty T
	return empty
}

func findMethodBody(pkg *packages.Package, method *types.Selection) *ast.BlockStmt {
	node := findNodeContaining[*ast.FuncDecl](pkg, method.Obj().Pos())
	return node.Body
}

func findReceiverName(pkg *packages.Package, method *types.Selection) string {
	node := findNodeContaining[*ast.FuncDecl](pkg, method.Obj().Pos())
	if len(node.Recv.List[0].Names) == 0 {
		return "_"
	}
	return node.Recv.List[0].Names[0].Name
}

func buildFieldResolver(c *genCtx, pkg *packages.Package, typeName string, fieldDef *fieldMeta) []ast.Node {
	var nodes []ast.Node

	sig := fieldDef.Method.Type().(*types.Signature)
	parms := sig.Params()
	needQueryInfo := false
	needsArgs := parms.Len() > 0

	var gql string
	if fieldDef.GQL.Description != "" {
		gql = gqast.StringValue{V: fieldDef.GQL.Description}.Representation() + "\n"
	}

	gql += fieldDef.Name

	if len(fieldDef.GQL.Directives) > 0 {
		var sb strings.Builder
		fieldDef.GQL.Directives.MarshalGraphQL(&sb)
		gql += " " + sb.String()
	}

	var argMap []argTranslate

	if needsArgs {
		var injectedVars []*types.Var
		var argVars []*types.Var
		for i := 0; i < parms.Len(); i++ {
			parm := parms.At(i)
			if c.isContextType(parm.Type()) {
				argMap = append(argMap, argTranslate{parm.Name(), &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "args",
							},
							Sel: &ast.Ident{
								Name: "QueryInfo",
							},
						},
						Sel: &ast.Ident{
							Name: "QueryContext",
						},
					},
				}})
				needQueryInfo = true
			} else if c.isResolverContextType(parm.Type()) {
				needQueryInfo = true
			} else if c.isInjectedArg(parm) {
				argMap = append(argMap, argTranslate{parm.Name(), &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "args",
					},
					Sel: &ast.Ident{
						Name: kace.Pascal(parm.Name()),
					},
				}})
				injectedVars = append(injectedVars, parm)
			} else {
				argMap = append(argMap, argTranslate{parm.Name(), &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "args",
					},
					Sel: &ast.Ident{
						Name: kace.Pascal(parm.Name()),
					},
				}})
				argVars = append(argVars, parm)
			}
		}

		var fl []*ast.Field
		if needQueryInfo {
			fl = append(fl, &ast.Field{
				Names: []*ast.Ident{
					{
						Name: "QueryInfo",
					},
				},
				Type: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "ts",
					},
					Sel: &ast.Ident{
						Name: "QueryInfo",
					},
				},
				Tag: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "`gq:\"@inject\"`",
				},
			})
		}

		for _, v := range injectedVars {
			n := findNodeContaining[*ast.Field](pkg, v.Pos())

			fl = append(fl, &ast.Field{
				Names: []*ast.Ident{
					{
						Name: kace.Pascal(v.Name()),
					},
				},
				Type: n.Type,
				Tag: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "`gq:\"@inject\"`",
				},
			})
		}

		for _, v := range argVars {
			n := findNodeContaining[*ast.Field](pkg, v.Pos())

			fl = append(fl, &ast.Field{
				Names: []*ast.Ident{
					{
						Name: kace.Pascal(v.Name()),
					},
				},
				Type: n.Type,
			})
		}

		argsStructName := typeName + kace.Pascal(fieldDef.Name) + "Args"

		nodes = append(nodes, &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{
						Name: argsStructName,
					},
					Type: &ast.StructType{
						Fields: &ast.FieldList{
							List: fl,
						},
					},
				},
			},
		})

		nodes = append(nodes, &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						{
							Name: typeName + kace.Pascal(fieldDef.Name) + "Field",
						},
					},
					Values: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "ts",
								},
								Sel: &ast.Ident{
									Name: "AddFieldWithArgs",
								},
							},
							Args: []ast.Expr{
								&ast.Ident{
									Name: typeName + "Type",
								},
								&ast.BasicLit{
									Kind:  token.STRING,
									Value: "`" + gql + "`",
								},
								&ast.FuncLit{
									Body: getTransformedBody(pkg, fieldDef, argMap),
									Type: &ast.FuncType{
										Params: &ast.FieldList{
											List: []*ast.Field{
												{
													Names: []*ast.Ident{
														{
															Name: findReceiverName(pkg, fieldDef.Method),
														},
													},
													Type: &ast.StarExpr{
														X: &ast.Ident{
															Name: typeName,
														},
													},
												},

												{
													Names: []*ast.Ident{
														{
															Name: "args",
														},
													},
													Type: &ast.StarExpr{
														X: &ast.Ident{
															Name: argsStructName,
														},
													},
												},
											},
										},
										Results: &ast.FieldList{
											List: []*ast.Field{
												{
													Type: &ast.IndexExpr{
														X: &ast.SelectorExpr{
															X: &ast.Ident{
																Name: "ts",
															},
															Sel: &ast.Ident{
																Name: "Result",
															},
														},
														Index: getFieldType(pkg, sig),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
	} else {
		nodes = append(nodes, &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						{
							Name: typeName + kace.Pascal(fieldDef.Name) + "Field",
						},
					},
					Values: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "ts",
								},
								Sel: &ast.Ident{
									Name: "AddField",
								},
							},
							Args: []ast.Expr{
								&ast.Ident{
									Name: typeName + "Type",
								},
								&ast.BasicLit{
									Kind:  token.STRING,
									Value: "`" + gql + "`",
								},
								&ast.FuncLit{
									Body: getTransformedBody(pkg, fieldDef, argMap),
									Type: &ast.FuncType{
										Params: &ast.FieldList{
											List: []*ast.Field{
												{
													Names: []*ast.Ident{
														{
															Name: findReceiverName(pkg, fieldDef.Method),
														},
													},
													Type: &ast.StarExpr{
														X: &ast.Ident{
															Name: typeName,
														},
													},
												},
											},
										},
										Results: &ast.FieldList{
											List: []*ast.Field{
												{
													Type: &ast.IndexExpr{
														X: &ast.SelectorExpr{
															X: &ast.Ident{
																Name: "ts",
															},
															Sel: &ast.Ident{
																Name: "Result",
															},
														},
														Index: getFieldType(pkg, sig),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		})
	}

	return nodes
}

func getTransformedBody(pkg *packages.Package, fieldDef *fieldMeta, argMap []argTranslate) *ast.BlockStmt {
	body := findMethodBody(pkg, fieldDef.Method)
	body = fixArgs(argMap, body)
	return fixReturns(pkg, fieldDef.Method.Type().(*types.Signature), body)
}

func fixReturns(pkg *packages.Package, sig *types.Signature, body *ast.BlockStmt) *ast.BlockStmt {
	var rule transformRule
	results := sig.Results()
	if results.Len() == 1 {
		if fun, ok := results.At(0).Type().Underlying().(*types.Signature); ok {
			rule = transformRule{
				matcher: matchAnyUntil(match[*ast.ReturnStmt]()),
				action: func(c *astutil.Cursor) {
					if fun.Params().Len() == 0 {
						if lit, ok := c.Node().(*ast.ReturnStmt).Results[0].(*ast.FuncLit); ok {
							lit.Type.Params = &ast.FieldList{
								List: []*ast.Field{
									{
										Type: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "context",
											},
											Sel: &ast.Ident{
												Name: "Context",
											},
										},
									},
								},
							}
						}
					}

					c.Replace(&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "result",
									},
									Sel: &ast.Ident{
										Name: "Async",
									},
								},
								Args: c.Node().(*ast.ReturnStmt).Results,
							},
						},
					})
				},
			}
		} else if _, ok := results.At(0).Type().Underlying().(*types.Chan); ok {
			rule = transformRule{
				matcher: matchAnyUntil(match[*ast.ReturnStmt]()),
				action: func(c *astutil.Cursor) {
					c.Replace(&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "result",
									},
									Sel: &ast.Ident{
										Name: "SuccessChan",
									},
								},
								Args: c.Node().(*ast.ReturnStmt).Results,
							},
						},
					})
				},
			}
		} else {
			rule = transformRule{
				matcher: matchAnyUntil(match[*ast.ReturnStmt]()),
				action: func(c *astutil.Cursor) {
					c.Replace(&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "result",
									},
									Sel: &ast.Ident{
										Name: "Of",
									},
								},
								Args: c.Node().(*ast.ReturnStmt).Results,
							},
						},
					})
				},
			}
		}
	} else {
		if _, ok := results.At(0).Type().Underlying().(*types.Chan); ok {
			rule = transformRule{
				matcher: matchAnyUntil(match[*ast.ReturnStmt]()),
				action: func(c *astutil.Cursor) {
					c.Replace(&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "result",
									},
									Sel: &ast.Ident{
										Name: "Chans",
									},
								},
								Args: c.Node().(*ast.ReturnStmt).Results,
							},
						},
					})
				},
			}
		} else {
			rule = transformRule{
				matcher: matchAnyUntil(match[*ast.ReturnStmt]()),
				action: func(c *astutil.Cursor) {
					c.Replace(&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "result",
									},
									Sel: &ast.Ident{
										Name: "Wrap",
									},
								},
								Args: c.Node().(*ast.ReturnStmt).Results,
							},
						},
					})
				},
			}
		}
	}

	transform(body, rule)
	return body
}

func fixArgs(argMap []argTranslate, body *ast.BlockStmt) *ast.BlockStmt {
	var assignStmts []ast.Stmt
	for _, item := range argMap {
		assignStmts = append(assignStmts, &ast.AssignStmt{
			Tok: token.DEFINE,
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: item.paramName,
				},
			},
			Rhs: []ast.Expr{
				item.expr,
			},
		})
	}
	body.List = append(assignStmts, body.List...)
	return body
}

func getFieldType(pkg *packages.Package, sig *types.Signature) ast.Expr {
	results := sig.Results()
	if results.Len() == 1 {
		if fun, ok := results.At(0).Type().Underlying().(*types.Signature); ok {
			return getFieldType(pkg, fun)
		} else if _, ok := results.At(0).Type().Underlying().(*types.Chan); ok {
			f := findNodeContaining[*ast.Field](pkg, sig.Results().At(0).Pos())
			return f.Type.(*ast.ChanType).Value
		} else {
			f := findNodeContaining[*ast.Field](pkg, sig.Results().At(0).Pos())
			return f.Type
		}
	} else {
		if _, ok := results.At(0).Type().Underlying().(*types.Chan); ok {
			f := findNodeContaining[*ast.Field](pkg, sig.Results().At(0).Pos())
			return f.Type.(*ast.ChanType).Value
		} else {
			f := findNodeContaining[*ast.Field](pkg, sig.Results().At(0).Pos())
			return f.Type
		}
	}
}

type argTranslate struct {
	paramName string
	expr      ast.Expr
}
