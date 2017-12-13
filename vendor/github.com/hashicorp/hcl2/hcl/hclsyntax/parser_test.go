package hclsyntax

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/kylelemons/godebug/pretty"
	"github.com/zclconf/go-cty/cty"
)

func TestParseConfig(t *testing.T) {
	tests := []struct {
		input     string
		diagCount int
		want      *Body
	}{
		{
			``,
			0,
			&Body{
				Attributes: Attributes{},
				Blocks:     Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 1, Column: 1, Byte: 0},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 1, Column: 1, Byte: 0},
				},
			},
		},

		{
			"block {}\n",
			0,
			&Body{
				Attributes: Attributes{},
				Blocks: Blocks{
					&Block{
						Type:   "block",
						Labels: nil,
						Body: &Body{
							Attributes: Attributes{},
							Blocks:     Blocks{},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 9, Byte: 8},
								End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: nil,
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
							End:   hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 9},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 9},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 9},
				},
			},
		},
		{
			"block {}block {}\n",
			1, // missing newline after block definition
			&Body{
				Attributes: Attributes{},
				Blocks: Blocks{
					&Block{
						Type:   "block",
						Labels: nil,
						Body: &Body{
							Attributes: Attributes{},
							Blocks:     Blocks{},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 9, Byte: 8},
								End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: nil,
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
							End:   hcl.Pos{Line: 1, Column: 8, Byte: 7},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 8, Byte: 7},
							End:   hcl.Pos{Line: 1, Column: 9, Byte: 8},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 17},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 17},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 17},
				},
			},
		},
		{
			"block \"foo\" {}\n",
			0,
			&Body{
				Attributes: Attributes{},
				Blocks: Blocks{
					&Block{
						Type:   "block",
						Labels: []string{"foo"},
						Body: &Body{
							Attributes: Attributes{},
							Blocks:     Blocks{},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
								End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 15, Byte: 14},
								End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: []hcl.Range{
							{
								Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
							},
						},
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
							End:   hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 14, Byte: 13},
							End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 15},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
			},
		},
		{
			`
block "invalid" 1.2 {}
block "valid" {}
`,
			1,
			&Body{
				Attributes: Attributes{},
				Blocks: Blocks{
					&Block{
						Type:   "block",
						Labels: []string{"invalid"},
						Body:   nil,

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 2, Column: 1, Byte: 1},
							End:   hcl.Pos{Line: 2, Column: 6, Byte: 6},
						},
						LabelRanges: []hcl.Range{
							{
								Start: hcl.Pos{Line: 2, Column: 7, Byte: 7},
								End:   hcl.Pos{Line: 2, Column: 16, Byte: 16},
							},
						},

						// Since we failed parsing before we got to the
						// braces, the type range is used as a placeholder
						// for these.
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 2, Column: 1, Byte: 1},
							End:   hcl.Pos{Line: 2, Column: 6, Byte: 6},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 2, Column: 1, Byte: 1},
							End:   hcl.Pos{Line: 2, Column: 6, Byte: 6},
						},
					},

					// Recovery behavior should allow us to still see this
					// second block, even though the first was invalid.
					&Block{
						Type:   "block",
						Labels: []string{"valid"},
						Body: &Body{
							Attributes: Attributes{},
							Blocks:     Blocks{},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 3, Column: 15, Byte: 38},
								End:   hcl.Pos{Line: 3, Column: 17, Byte: 40},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 3, Column: 17, Byte: 40},
								End:   hcl.Pos{Line: 3, Column: 17, Byte: 40},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 3, Column: 1, Byte: 24},
							End:   hcl.Pos{Line: 3, Column: 6, Byte: 29},
						},
						LabelRanges: []hcl.Range{
							{
								Start: hcl.Pos{Line: 3, Column: 7, Byte: 30},
								End:   hcl.Pos{Line: 3, Column: 14, Byte: 37},
							},
						},
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 3, Column: 15, Byte: 38},
							End:   hcl.Pos{Line: 3, Column: 16, Byte: 39},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 3, Column: 16, Byte: 39},
							End:   hcl.Pos{Line: 3, Column: 17, Byte: 40},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 4, Column: 1, Byte: 41},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 4, Column: 1, Byte: 41},
					End:   hcl.Pos{Line: 4, Column: 1, Byte: 41},
				},
			},
		},
		{
			`block "f\o" {}
`,
			1, // "\o" is not a valid escape sequence
			&Body{
				Attributes: Attributes{},
				Blocks: Blocks{
					&Block{
						Type:   "block",
						Labels: []string{"fo"},
						Body:   nil,

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: []hcl.Range{
							{
								Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
							},
						},
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 15},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
			},
		},
		{
			`block "f\n" {}
`,
			0,
			&Body{
				Attributes: Attributes{},
				Blocks: Blocks{
					&Block{
						Type:   "block",
						Labels: []string{"f\n"},
						Body: &Body{
							Attributes: Attributes{},
							Blocks:     Blocks{},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
								End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
							EndRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 15, Byte: 14},
								End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
							},
						},

						TypeRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						LabelRanges: []hcl.Range{
							{
								Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
								End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
							},
						},
						OpenBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 13, Byte: 12},
							End:   hcl.Pos{Line: 1, Column: 14, Byte: 13},
						},
						CloseBraceRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 14, Byte: 13},
							End:   hcl.Pos{Line: 1, Column: 15, Byte: 14},
						},
					},
				},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 15},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 15},
				},
			},
		},

		{
			"a = 1\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &LiteralValueExpr{
							Val: cty.NumberIntVal(1),

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 6},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 6},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 6},
				},
			},
		},
		{
			"a = \"hello ${true}\"\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &TemplateExpr{
							Parts: []Expression{
								&LiteralValueExpr{
									Val: cty.StringVal("hello "),

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 6, Byte: 5},
										End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
									},
								},
								&LiteralValueExpr{
									Val: cty.True,

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 14, Byte: 13},
										End:   hcl.Pos{Line: 1, Column: 18, Byte: 17},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 20, Byte: 19},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 20, Byte: 19},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 20},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 20},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 20},
				},
			},
		},
		{
			"a = foo.bar\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &ScopeTraversalExpr{
							Traversal: hcl.Traversal{
								hcl.TraverseRoot{
									Name: "foo",

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
										End:   hcl.Pos{Line: 1, Column: 8, Byte: 7},
									},
								},
								hcl.TraverseAttr{
									Name: "bar",

									SrcRange: hcl.Range{
										Start: hcl.Pos{Line: 1, Column: 8, Byte: 7},
										End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
									},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 12},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 12},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 12},
				},
			},
		},
		{
			"a = 1 # line comment\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &LiteralValueExpr{
							Val: cty.NumberIntVal(1),

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 21},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 21},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 21},
				},
			},
		},

		{
			"a = [for k, v in foo: v if true]\n",
			0,
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &ForExpr{
							KeyVar: "k",
							ValVar: "v",

							CollExpr: &ScopeTraversalExpr{
								Traversal: hcl.Traversal{
									hcl.TraverseRoot{
										Name: "foo",
										SrcRange: hcl.Range{
											Start: hcl.Pos{Line: 1, Column: 18, Byte: 17},
											End:   hcl.Pos{Line: 1, Column: 21, Byte: 20},
										},
									},
								},
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 18, Byte: 17},
									End:   hcl.Pos{Line: 1, Column: 21, Byte: 20},
								},
							},
							ValExpr: &ScopeTraversalExpr{
								Traversal: hcl.Traversal{
									hcl.TraverseRoot{
										Name: "v",
										SrcRange: hcl.Range{
											Start: hcl.Pos{Line: 1, Column: 23, Byte: 22},
											End:   hcl.Pos{Line: 1, Column: 24, Byte: 23},
										},
									},
								},
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 23, Byte: 22},
									End:   hcl.Pos{Line: 1, Column: 24, Byte: 23},
								},
							},
							CondExpr: &LiteralValueExpr{
								Val: cty.True,
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 28, Byte: 27},
									End:   hcl.Pos{Line: 1, Column: 32, Byte: 31},
								},
							},

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 33, Byte: 32},
							},
							OpenRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
							CloseRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 32, Byte: 31},
								End:   hcl.Pos{Line: 1, Column: 33, Byte: 32},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 33, Byte: 32},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 33},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 33},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 33},
				},
			},
		},
		{
			"a = [for k, v in foo: k => v... if true]\n",
			2, // can't use => or ... in a tuple for
			&Body{
				Attributes: Attributes{
					"a": {
						Name: "a",
						Expr: &ForExpr{
							KeyVar: "k",
							ValVar: "v",

							CollExpr: &ScopeTraversalExpr{
								Traversal: hcl.Traversal{
									hcl.TraverseRoot{
										Name: "foo",
										SrcRange: hcl.Range{
											Start: hcl.Pos{Line: 1, Column: 18, Byte: 17},
											End:   hcl.Pos{Line: 1, Column: 21, Byte: 20},
										},
									},
								},
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 18, Byte: 17},
									End:   hcl.Pos{Line: 1, Column: 21, Byte: 20},
								},
							},
							KeyExpr: &ScopeTraversalExpr{
								Traversal: hcl.Traversal{
									hcl.TraverseRoot{
										Name: "k",
										SrcRange: hcl.Range{
											Start: hcl.Pos{Line: 1, Column: 23, Byte: 22},
											End:   hcl.Pos{Line: 1, Column: 24, Byte: 23},
										},
									},
								},
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 23, Byte: 22},
									End:   hcl.Pos{Line: 1, Column: 24, Byte: 23},
								},
							},
							ValExpr: &ScopeTraversalExpr{
								Traversal: hcl.Traversal{
									hcl.TraverseRoot{
										Name: "v",
										SrcRange: hcl.Range{
											Start: hcl.Pos{Line: 1, Column: 28, Byte: 27},
											End:   hcl.Pos{Line: 1, Column: 29, Byte: 28},
										},
									},
								},
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 28, Byte: 27},
									End:   hcl.Pos{Line: 1, Column: 29, Byte: 28},
								},
							},
							CondExpr: &LiteralValueExpr{
								Val: cty.True,
								SrcRange: hcl.Range{
									Start: hcl.Pos{Line: 1, Column: 36, Byte: 35},
									End:   hcl.Pos{Line: 1, Column: 40, Byte: 39},
								},
							},
							Group: true,

							SrcRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 41, Byte: 40},
							},
							OpenRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
								End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
							},
							CloseRange: hcl.Range{
								Start: hcl.Pos{Line: 1, Column: 40, Byte: 39},
								End:   hcl.Pos{Line: 1, Column: 41, Byte: 40},
							},
						},

						SrcRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 41, Byte: 40},
						},
						NameRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
						},
						EqualsRange: hcl.Range{
							Start: hcl.Pos{Line: 1, Column: 3, Byte: 2},
							End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
						},
					},
				},
				Blocks: Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 41},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 2, Column: 1, Byte: 41},
					End:   hcl.Pos{Line: 2, Column: 1, Byte: 41},
				},
			},
		},

		{
			`	`,
			2, // tabs not allowed, and body item is required here
			&Body{
				Attributes: Attributes{},
				Blocks:     Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 2, Byte: 1},
					End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
		},
		{
			`\x81`,
			2, // invalid UTF-8, and body item is required here
			&Body{
				Attributes: Attributes{},
				Blocks:     Blocks{},
				SrcRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
					End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
				EndRange: hcl.Range{
					Start: hcl.Pos{Line: 1, Column: 2, Byte: 1},
					End:   hcl.Pos{Line: 1, Column: 2, Byte: 1},
				},
			},
		},
	}

	prettyConfig := &pretty.Config{
		Diffable:          true,
		IncludeUnexported: true,
		PrintStringers:    true,
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			file, diags := ParseConfig([]byte(test.input), "", hcl.Pos{Byte: 0, Line: 1, Column: 1})
			if len(diags) != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}

			got := file.Body

			if !reflect.DeepEqual(got, test.want) {
				diff := prettyConfig.Compare(test.want, got)
				if diff != "" {
					t.Errorf("wrong result\ninput: %s\ndiff:  %s", test.input, diff)
				} else {
					t.Errorf("wrong result\ninput: %s\ngot:   %s\nwant:  %s", test.input, spew.Sdump(got), spew.Sdump(test.want))
				}
			}
		})
	}
}
