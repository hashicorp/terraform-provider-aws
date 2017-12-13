package hclsyntax

import (
	"fmt"
	"testing"

	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
)

func TestVariables(t *testing.T) {
	tests := []struct {
		Expr Expression
		Want []hcl.Traversal
	}{
		{
			&LiteralValueExpr{
				Val: cty.True,
			},
			nil,
		},
		{
			&ScopeTraversalExpr{
				Traversal: hcl.Traversal{
					hcl.TraverseRoot{
						Name: "foo",
					},
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "foo",
					},
				},
			},
		},
		{
			&BinaryOpExpr{
				LHS: &ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "foo",
						},
					},
				},
				Op: OpAdd,
				RHS: &ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "bar",
						},
					},
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "foo",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "bar",
					},
				},
			},
		},
		{
			&UnaryOpExpr{
				Val: &ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "foo",
						},
					},
				},
				Op: OpNegate,
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "foo",
					},
				},
			},
		},
		{
			&ConditionalExpr{
				Condition: &ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "foo",
						},
					},
				},
				TrueResult: &ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "bar",
						},
					},
				},
				FalseResult: &ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "baz",
						},
					},
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "foo",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "bar",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "baz",
					},
				},
			},
		},
		{
			&ForExpr{
				KeyVar: "k",
				ValVar: "v",

				CollExpr: &ScopeTraversalExpr{
					Traversal: hcl.Traversal{
						hcl.TraverseRoot{
							Name: "foo",
						},
					},
				},
				KeyExpr: &BinaryOpExpr{
					LHS: &ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{
								Name: "k",
							},
						},
					},
					Op: OpAdd,
					RHS: &ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{
								Name: "bar",
							},
						},
					},
				},
				ValExpr: &BinaryOpExpr{
					LHS: &ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{
								Name: "v",
							},
						},
					},
					Op: OpAdd,
					RHS: &ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{
								Name: "baz",
							},
						},
					},
				},
				CondExpr: &BinaryOpExpr{
					LHS: &ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{
								Name: "k",
							},
						},
					},
					Op: OpLessThan,
					RHS: &ScopeTraversalExpr{
						Traversal: hcl.Traversal{
							hcl.TraverseRoot{
								Name: "limit",
							},
						},
					},
				},
			},
			[]hcl.Traversal{
				{
					hcl.TraverseRoot{
						Name: "foo",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "bar",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "baz",
					},
				},
				{
					hcl.TraverseRoot{
						Name: "limit",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Expr), func(t *testing.T) {
			got := Variables(test.Expr)

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\ngot:  %s\nwant: %s", spew.Sdump(got), spew.Sdump(test.Want))
			}
		})
	}
}
