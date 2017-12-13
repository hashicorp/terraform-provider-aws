package parser

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hil/ast"
	"github.com/hashicorp/hil/scanner"
)

func TestParser(t *testing.T) {
	cases := []struct {
		Input  string
		Error  bool
		Result ast.Node
	}{
		{
			"",
			false,
			&ast.LiteralNode{
				Value: "",
				Typex: ast.TypeString,
				Posx:  ast.Pos{Column: 1, Line: 1},
			},
		},

		{
			"$",
			false,
			&ast.LiteralNode{
				Value: "$",
				Typex: ast.TypeString,
				Posx:  ast.Pos{Column: 1, Line: 1},
			},
		},

		{
			"foo",
			false,
			&ast.LiteralNode{
				Value: "foo",
				Typex: ast.TypeString,
				Posx:  ast.Pos{Column: 1, Line: 1},
			},
		},

		{
			"$${var.foo}",
			false,
			&ast.LiteralNode{
				Value: "${var.foo}",
				Typex: ast.TypeString,
				Posx:  ast.Pos{Column: 1, Line: 1},
			},
		},

		// Identifier starting with a number
		{
			`foo ${123abcd}`,
			true,
			nil,
		},

		// Identifier starting with a *
		{
			`foo ${*abcd}`,
			true,
			nil,
		},

		{
			"foo ${var.bar}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.VariableAccess{
						Name: "var.bar",
						Posx: ast.Pos{Column: 7, Line: 1},
					},
				},
			},
		},

		{
			"foo ${var.bar.*.baz}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.VariableAccess{
						Name: "var.bar.*.baz",
						Posx: ast.Pos{Column: 7, Line: 1},
					},
				},
			},
		},

		{
			"foo ${var.bar} baz",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.VariableAccess{
						Name: "var.bar",
						Posx: ast.Pos{Column: 7, Line: 1},
					},
					&ast.LiteralNode{
						Value: " baz",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 15, Line: 1},
					},
				},
			},
		},

		{
			"foo ${var.bar.0}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.VariableAccess{
						Name: "var.bar.0",
						Posx: ast.Pos{Column: 7, Line: 1},
					},
				},
			},
		},

		{
			"foo ${foo.foo-bar.baz.0.attr}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.VariableAccess{
						Name: "foo.foo-bar.baz.0.attr",
						Posx: ast.Pos{Column: 7, Line: 1},
					},
				},
			},
		},

		{
			`foo ${"bar"}`,
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.LiteralNode{
						Value: "bar",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 8, Line: 1},
					},
				},
			},
		},

		{
			`foo ${"bar\nbaz"}`,
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.LiteralNode{
						Value: "bar\nbaz",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 8, Line: 1},
					},
				},
			},
		},

		{
			`foo ${"bar \"baz\""}`,
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.LiteralNode{
						Value: `bar "baz"`,
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 8, Line: 1},
					},
				},
			},
		},

		{
			`foo ${func('baz')}`,
			true,
			nil,
		},

		{
			"foo ${42}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.LiteralNode{
						Value: 42,
						Typex: ast.TypeInt,
						Posx:  ast.Pos{Column: 7, Line: 1},
					},
				},
			},
		},

		{
			"foo ${3.14159}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.LiteralNode{
						Value: 3.14159,
						Typex: ast.TypeFloat,
						Posx:  ast.Pos{Column: 7, Line: 1},
					},
				},
			},
		},

		{
			"föo ${true}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "föo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.LiteralNode{
						Value: true,
						Typex: ast.TypeBool,
						Posx:  ast.Pos{Column: 7, Line: 1},
					},
				},
			},
		},

		{
			"foo ${42+1}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.Arithmetic{
						Op: ast.ArithmeticOpAdd,
						Exprs: []ast.Node{
							&ast.LiteralNode{
								Value: 42,
								Typex: ast.TypeInt,
								Posx:  ast.Pos{Column: 7, Line: 1},
							},
							&ast.LiteralNode{
								Value: 1,
								Typex: ast.TypeInt,
								Posx:  ast.Pos{Column: 10, Line: 1},
							},
						},
						Posx: ast.Pos{Column: 7, Line: 1},
					},
				},
			},
		},

		{
			"foo ${-1}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.Arithmetic{
						Op: ast.ArithmeticOpSub,
						Exprs: []ast.Node{
							&ast.LiteralNode{
								Value: 0,
								Typex: ast.TypeInt,
								Posx:  ast.Pos{Column: 7, Line: 1},
							},
							&ast.LiteralNode{
								Value: 1,
								Typex: ast.TypeInt,
								Posx:  ast.Pos{Column: 8, Line: 1},
							},
						},
						Posx: ast.Pos{Column: 7, Line: 1},
					},
				},
			},
		},

		{
			"foo ${var.bar*1} baz",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.Arithmetic{
						Op: ast.ArithmeticOpMul,
						Exprs: []ast.Node{
							&ast.VariableAccess{
								Name: "var.bar",
								Posx: ast.Pos{Column: 7, Line: 1},
							},
							&ast.LiteralNode{
								Value: 1,
								Typex: ast.TypeInt,
								Posx:  ast.Pos{Column: 15, Line: 1},
							},
						},
						Posx: ast.Pos{Column: 7, Line: 1},
					},
					&ast.LiteralNode{
						Value: " baz",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 17, Line: 1},
					},
				},
			},
		},

		{
			"${!a}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					// !a parses as (false == a)
					&ast.Arithmetic{
						Op: ast.ArithmeticOpEqual,
						Exprs: []ast.Node{
							&ast.LiteralNode{
								Value: false,
								Typex: ast.TypeBool,
								Posx:  ast.Pos{Column: 3, Line: 1},
							},
							&ast.VariableAccess{
								Name: "a",
								Posx: ast.Pos{Column: 4, Line: 1},
							},
						},
						Posx: ast.Pos{Column: 3, Line: 1},
					},
				},
			},
		},

		{
			"${a==b}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Arithmetic{
						Op: ast.ArithmeticOpEqual,
						Exprs: []ast.Node{
							&ast.VariableAccess{
								Name: "a",
								Posx: ast.Pos{Column: 3, Line: 1},
							},
							&ast.VariableAccess{
								Name: "b",
								Posx: ast.Pos{Column: 6, Line: 1},
							},
						},
						Posx: ast.Pos{Column: 3, Line: 1},
					},
				},
			},
		},

		{
			"${a!=b}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Arithmetic{
						Op: ast.ArithmeticOpNotEqual,
						Exprs: []ast.Node{
							&ast.VariableAccess{
								Name: "a",
								Posx: ast.Pos{Column: 3, Line: 1},
							},
							&ast.VariableAccess{
								Name: "b",
								Posx: ast.Pos{Column: 6, Line: 1},
							},
						},
						Posx: ast.Pos{Column: 3, Line: 1},
					},
				},
			},
		},

		{
			"${a < 5 ? a + 5 : a + 10}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Conditional{
						CondExpr: &ast.Arithmetic{
							Op: ast.ArithmeticOpLessThan,
							Exprs: []ast.Node{
								&ast.VariableAccess{
									Name: "a",
									Posx: ast.Pos{Column: 3, Line: 1},
								},
								&ast.LiteralNode{
									Value: 5,
									Typex: ast.TypeInt,
									Posx:  ast.Pos{Column: 7, Line: 1},
								},
							},
							Posx: ast.Pos{Column: 3, Line: 1},
						},
						TrueExpr: &ast.Arithmetic{
							Op: ast.ArithmeticOpAdd,
							Exprs: []ast.Node{
								&ast.VariableAccess{
									Name: "a",
									Posx: ast.Pos{Column: 11, Line: 1},
								},
								&ast.LiteralNode{
									Value: 5,
									Typex: ast.TypeInt,
									Posx:  ast.Pos{Column: 15, Line: 1},
								},
							},
							Posx: ast.Pos{Column: 11, Line: 1},
						},
						FalseExpr: &ast.Arithmetic{
							Op: ast.ArithmeticOpAdd,
							Exprs: []ast.Node{
								&ast.VariableAccess{
									Name: "a",
									Posx: ast.Pos{Column: 19, Line: 1},
								},
								&ast.LiteralNode{
									Value: 10,
									Typex: ast.TypeInt,
									Posx:  ast.Pos{Column: 23, Line: 1},
								},
							},
							Posx: ast.Pos{Column: 19, Line: 1},
						},
						Posx: ast.Pos{Column: 3, Line: 1},
					},
				},
			},
		},

		{
			"${true&&false}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Arithmetic{
						Op: ast.ArithmeticOpLogicalAnd,
						Exprs: []ast.Node{
							&ast.LiteralNode{
								Value: true,
								Typex: ast.TypeBool,
								Posx:  ast.Pos{Column: 3, Line: 1},
							},
							&ast.LiteralNode{
								Value: false,
								Typex: ast.TypeBool,
								Posx:  ast.Pos{Column: 9, Line: 1},
							},
						},
						Posx: ast.Pos{Column: 3, Line: 1},
					},
				},
			},
		},

		{
			"${true||false}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Arithmetic{
						Op: ast.ArithmeticOpLogicalOr,
						Exprs: []ast.Node{
							&ast.LiteralNode{
								Value: true,
								Typex: ast.TypeBool,
								Posx:  ast.Pos{Column: 3, Line: 1},
							},
							&ast.LiteralNode{
								Value: false,
								Typex: ast.TypeBool,
								Posx:  ast.Pos{Column: 9, Line: 1},
							},
						},
						Posx: ast.Pos{Column: 3, Line: 1},
					},
				},
			},
		},

		{
			"${a||b&&c}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Arithmetic{
						Op: ast.ArithmeticOpLogicalOr,
						Exprs: []ast.Node{
							&ast.VariableAccess{
								Name: "a",
								Posx: ast.Pos{Column: 3, Line: 1},
							},
							&ast.Arithmetic{
								Op: ast.ArithmeticOpLogicalAnd,
								Exprs: []ast.Node{
									&ast.VariableAccess{
										Name: "b",
										Posx: ast.Pos{Column: 6, Line: 1},
									},
									&ast.VariableAccess{
										Name: "c",
										Posx: ast.Pos{Column: 9, Line: 1},
									},
								},
								Posx: ast.Pos{Column: 6, Line: 1},
							},
						},
						Posx: ast.Pos{Column: 3, Line: 1},
					},
				},
			},
		},

		{
			"${a&&b||c}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Arithmetic{
						Op: ast.ArithmeticOpLogicalOr,
						Exprs: []ast.Node{
							&ast.Arithmetic{
								Op: ast.ArithmeticOpLogicalAnd,
								Exprs: []ast.Node{
									&ast.VariableAccess{
										Name: "a",
										Posx: ast.Pos{Column: 3, Line: 1},
									},
									&ast.VariableAccess{
										Name: "b",
										Posx: ast.Pos{Column: 6, Line: 1},
									},
								},
								Posx: ast.Pos{Column: 3, Line: 1},
							},
							&ast.VariableAccess{
								Name: "c",
								Posx: ast.Pos{Column: 9, Line: 1},
							},
						},
						Posx: ast.Pos{Column: 3, Line: 1},
					},
				},
			},
		},

		{
			"${a<5||b>2}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Arithmetic{
						Op: ast.ArithmeticOpLogicalOr,
						Exprs: []ast.Node{
							&ast.Arithmetic{
								Op: ast.ArithmeticOpLessThan,
								Exprs: []ast.Node{
									&ast.VariableAccess{
										Name: "a",
										Posx: ast.Pos{Column: 3, Line: 1},
									},
									&ast.LiteralNode{
										Value: 5,
										Typex: ast.TypeInt,
										Posx:  ast.Pos{Column: 5, Line: 1},
									},
								},
								Posx: ast.Pos{Column: 3, Line: 1},
							},
							&ast.Arithmetic{
								Op: ast.ArithmeticOpGreaterThan,
								Exprs: []ast.Node{
									&ast.VariableAccess{
										Name: "b",
										Posx: ast.Pos{Column: 8, Line: 1},
									},
									&ast.LiteralNode{
										Value: 2,
										Typex: ast.TypeInt,
										Posx:  ast.Pos{Column: 10, Line: 1},
									},
								},
								Posx: ast.Pos{Column: 8, Line: 1},
							},
						},
						Posx: ast.Pos{Column: 3, Line: 1},
					},
				},
			},
		},

		{
			"${a<5&&b>2}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Arithmetic{
						Op: ast.ArithmeticOpLogicalAnd,
						Exprs: []ast.Node{
							&ast.Arithmetic{
								Op: ast.ArithmeticOpLessThan,
								Exprs: []ast.Node{
									&ast.VariableAccess{
										Name: "a",
										Posx: ast.Pos{Column: 3, Line: 1},
									},
									&ast.LiteralNode{
										Value: 5,
										Typex: ast.TypeInt,
										Posx:  ast.Pos{Column: 5, Line: 1},
									},
								},
								Posx: ast.Pos{Column: 3, Line: 1},
							},
							&ast.Arithmetic{
								Op: ast.ArithmeticOpGreaterThan,
								Exprs: []ast.Node{
									&ast.VariableAccess{
										Name: "b",
										Posx: ast.Pos{Column: 8, Line: 1},
									},
									&ast.LiteralNode{
										Value: 2,
										Typex: ast.TypeInt,
										Posx:  ast.Pos{Column: 10, Line: 1},
									},
								},
								Posx: ast.Pos{Column: 8, Line: 1},
							},
						},
						Posx: ast.Pos{Column: 3, Line: 1},
					},
				},
			},
		},

		{
			"${föo()}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Call{
						Func: "föo",
						Args: nil,
						Posx: ast.Pos{Column: 3, Line: 1},
					},
				},
			},
		},

		{
			"${foo(bar)}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Call{
						Func: "foo",
						Posx: ast.Pos{Column: 3, Line: 1},
						Args: []ast.Node{
							&ast.VariableAccess{
								Name: "bar",
								Posx: ast.Pos{Column: 7, Line: 1},
							},
						},
					},
				},
			},
		},

		{
			"${foo(bar, baz)}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Call{
						Func: "foo",
						Posx: ast.Pos{Column: 3, Line: 1},
						Args: []ast.Node{
							&ast.VariableAccess{
								Name: "bar",
								Posx: ast.Pos{Column: 7, Line: 1},
							},
							&ast.VariableAccess{
								Name: "baz",
								Posx: ast.Pos{Column: 12, Line: 1},
							},
						},
					},
				},
			},
		},

		{
			"${foo(bar(baz))}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Call{
						Func: "foo",
						Posx: ast.Pos{Column: 3, Line: 1},
						Args: []ast.Node{
							&ast.Call{
								Func: "bar",
								Posx: ast.Pos{Column: 7, Line: 1},
								Args: []ast.Node{
									&ast.VariableAccess{
										Name: "baz",
										Posx: ast.Pos{Column: 11, Line: 1},
									},
								},
							},
						},
					},
				},
			},
		},

		{
			`foo ${"bar ${baz}"}`,
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.LiteralNode{
						Value: "foo ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 1, Line: 1},
					},
					&ast.Output{
						Posx: ast.Pos{Column: 7, Line: 1},
						Exprs: []ast.Node{
							&ast.LiteralNode{
								Value: "bar ",
								Typex: ast.TypeString,
								Posx:  ast.Pos{Column: 8, Line: 1},
							},
							&ast.VariableAccess{
								Name: "baz",
								Posx: ast.Pos{Column: 14, Line: 1},
							},
						},
					},
				},
			},
		},

		{
			"${foo[1]}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Index{
						Posx: ast.Pos{Column: 6, Line: 1},
						Target: &ast.VariableAccess{
							Name: "foo",
							Posx: ast.Pos{Column: 3, Line: 1},
						},
						Key: &ast.LiteralNode{
							Value: 1,
							Typex: ast.TypeInt,
							Posx:  ast.Pos{Column: 7, Line: 1},
						},
					},
				},
			},
		},

		{
			"${foo[1]} - ${bar[0]}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Index{
						Posx: ast.Pos{Column: 6, Line: 1},
						Target: &ast.VariableAccess{
							Name: "foo",
							Posx: ast.Pos{Column: 3, Line: 1},
						},
						Key: &ast.LiteralNode{
							Value: 1,
							Typex: ast.TypeInt,
							Posx:  ast.Pos{Column: 7, Line: 1},
						},
					},
					&ast.LiteralNode{
						Value: " - ",
						Typex: ast.TypeString,
						Posx:  ast.Pos{Column: 10, Line: 1},
					},
					&ast.Index{
						Posx: ast.Pos{Column: 18, Line: 1},
						Target: &ast.VariableAccess{
							Name: "bar",
							Posx: ast.Pos{Column: 15, Line: 1},
						},
						Key: &ast.LiteralNode{
							Value: 0,
							Typex: ast.TypeInt,
							Posx:  ast.Pos{Column: 19, Line: 1},
						},
					},
				},
			},
		},

		{
			// * has higher precedence than +
			"${42+2*2}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Arithmetic{
						Posx: ast.Pos{Column: 3, Line: 1},
						Op:   ast.ArithmeticOpAdd,
						Exprs: []ast.Node{
							&ast.LiteralNode{
								Posx:  ast.Pos{Column: 3, Line: 1},
								Value: 42,
								Typex: ast.TypeInt,
							},
							&ast.Arithmetic{
								Posx: ast.Pos{Column: 6, Line: 1},
								Op:   ast.ArithmeticOpMul,
								Exprs: []ast.Node{
									&ast.LiteralNode{
										Posx:  ast.Pos{Column: 6, Line: 1},
										Value: 2,
										Typex: ast.TypeInt,
									},
									&ast.LiteralNode{
										Posx:  ast.Pos{Column: 8, Line: 1},
										Value: 2,
										Typex: ast.TypeInt,
									},
								},
							},
						},
					},
				},
			},
		},

		{
			// parentheses override precedence rules
			"${(42+2)*2}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Arithmetic{
						Posx: ast.Pos{Column: 3, Line: 1},
						Op:   ast.ArithmeticOpMul,
						Exprs: []ast.Node{
							&ast.Arithmetic{
								Posx: ast.Pos{Column: 4, Line: 1},
								Op:   ast.ArithmeticOpAdd,
								Exprs: []ast.Node{
									&ast.LiteralNode{
										Posx:  ast.Pos{Column: 4, Line: 1},
										Value: 42,
										Typex: ast.TypeInt,
									},
									&ast.LiteralNode{
										Posx:  ast.Pos{Column: 7, Line: 1},
										Value: 2,
										Typex: ast.TypeInt,
									},
								},
							},
							&ast.LiteralNode{
								Posx:  ast.Pos{Column: 10, Line: 1},
								Value: 2,
								Typex: ast.TypeInt,
							},
						},
					},
				},
			},
		},

		{
			// Left-associative parsing of operators with equal precedence
			"${42+2+2}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Arithmetic{
						Posx: ast.Pos{Column: 3, Line: 1},
						Op:   ast.ArithmeticOpAdd,
						Exprs: []ast.Node{
							&ast.Arithmetic{
								Posx: ast.Pos{Column: 3, Line: 1},
								Op:   ast.ArithmeticOpAdd,
								Exprs: []ast.Node{
									&ast.LiteralNode{
										Posx:  ast.Pos{Column: 3, Line: 1},
										Value: 42,
										Typex: ast.TypeInt,
									},
									&ast.LiteralNode{
										Posx:  ast.Pos{Column: 6, Line: 1},
										Value: 2,
										Typex: ast.TypeInt,
									},
								},
							},
							&ast.LiteralNode{
								Posx:  ast.Pos{Column: 8, Line: 1},
								Value: 2,
								Typex: ast.TypeInt,
							},
						},
					},
				},
			},
		},

		{
			// Unary - has higher precedence than addition
			"${42+-2+2}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Arithmetic{
						Posx: ast.Pos{Column: 3, Line: 1},
						Op:   ast.ArithmeticOpAdd,
						Exprs: []ast.Node{
							&ast.Arithmetic{
								Posx: ast.Pos{Column: 3, Line: 1},
								Op:   ast.ArithmeticOpAdd,
								Exprs: []ast.Node{
									&ast.LiteralNode{
										Posx:  ast.Pos{Column: 3, Line: 1},
										Value: 42,
										Typex: ast.TypeInt,
									},
									&ast.Arithmetic{
										Posx: ast.Pos{Column: 6, Line: 1},
										Op:   ast.ArithmeticOpSub,
										Exprs: []ast.Node{
											&ast.LiteralNode{
												Posx:  ast.Pos{Column: 6, Line: 1},
												Value: 0,
												Typex: ast.TypeInt,
											},
											&ast.LiteralNode{
												Posx:  ast.Pos{Column: 7, Line: 1},
												Value: 2,
												Typex: ast.TypeInt,
											},
										},
									},
								},
							},
							&ast.LiteralNode{
								Posx:  ast.Pos{Column: 9, Line: 1},
								Value: 2,
								Typex: ast.TypeInt,
							},
						},
					},
				},
			},
		},

		{
			"${-46+5}",
			false,
			&ast.Output{
				Posx: ast.Pos{Column: 1, Line: 1},
				Exprs: []ast.Node{
					&ast.Arithmetic{
						Posx: ast.Pos{Column: 3, Line: 1},
						Op:   ast.ArithmeticOpAdd,
						Exprs: []ast.Node{
							&ast.Arithmetic{
								Posx: ast.Pos{Column: 3, Line: 1},
								Op:   ast.ArithmeticOpSub,
								Exprs: []ast.Node{
									&ast.LiteralNode{
										Posx:  ast.Pos{Column: 3, Line: 1},
										Value: 0,
										Typex: ast.TypeInt,
									},
									&ast.LiteralNode{
										Posx:  ast.Pos{Column: 4, Line: 1},
										Value: 46,
										Typex: ast.TypeInt,
									},
								},
							},
							&ast.LiteralNode{
								Posx:  ast.Pos{Column: 7, Line: 1},
								Value: 5,
								Typex: ast.TypeInt,
							},
						},
					},
				},
			},
		},

		{
			"${foo=baz}",
			true,
			nil,
		},

		{
			"${foo&baz}",
			true,
			nil,
		},

		{
			"${foo|baz}",
			true,
			nil,
		},

		{
			"${foo[1][2]}",
			true,
			nil,
		},

		{
			`foo ${bar ${baz}}`,
			true,
			nil,
		},

		{
			`foo ${${baz}}`,
			true,
			nil,
		},

		{
			"${var",
			true,
			nil,
		},

		{
			`${"unclosed`,
			true,
			nil,
		},

		{
			`${"bar\nbaz}`,
			true,
			nil,
		},

		{
			`${ö(o("")`,
			true,
			nil,
		},

		{
			`${"${"${"`,
			true,
			nil,
		},

		{
			`${("$"`,
			true,
			nil,
		},

		{
			`${("${("${"${"$"`,
			true,
			nil,
		},

		{
			`${(p["$"`,
			true,
			nil,
		},

		{
			`${e(e,e,`,
			true,
			nil,
		},

		{
			"${file(/tmp/somefile)}",
			true,
			nil,
		},
	}

	for _, tc := range cases {
		ch := scanner.Scan(tc.Input, ast.Pos{Line: 1, Column: 1})
		actual, err := Parse(ch)
		if err != nil != tc.Error {
			t.Errorf("\nError: %s\n\nInput: %s\n", err, tc.Input)
		}
		if !reflect.DeepEqual(actual, tc.Result) {
			t.Errorf("\nGot:  %#v\nWant: %#v\n\nInput: %s\n", actual, tc.Result, tc.Input)
		}
	}
}
