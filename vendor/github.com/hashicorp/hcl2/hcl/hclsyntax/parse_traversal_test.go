package hclsyntax

import (
	"testing"

	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
)

func TestParseTraversalAbs(t *testing.T) {
	tests := []struct {
		src       string
		want      hcl.Traversal
		diagCount int
	}{
		{
			"",
			nil,
			1, // variable name required
		},
		{
			"foo",
			hcl.Traversal{
				hcl.TraverseRoot{
					Name: "foo",
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			0,
		},
		{
			"foo.bar.baz",
			hcl.Traversal{
				hcl.TraverseRoot{
					Name: "foo",
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
				hcl.TraverseAttr{
					Name: "bar",
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 4, Byte: 3},
						End:   hcl.Pos{Line: 1, Column: 8, Byte: 7},
					},
				},
				hcl.TraverseAttr{
					Name: "baz",
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 8, Byte: 7},
						End:   hcl.Pos{Line: 1, Column: 12, Byte: 11},
					},
				},
			},
			0,
		},
		{
			"foo[1]",
			hcl.Traversal{
				hcl.TraverseRoot{
					Name: "foo",
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
				hcl.TraverseIndex{
					Key: cty.NumberIntVal(1),
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 4, Byte: 3},
						End:   hcl.Pos{Line: 1, Column: 7, Byte: 6},
					},
				},
			},
			0,
		},
		{
			"foo[1][2]",
			hcl.Traversal{
				hcl.TraverseRoot{
					Name: "foo",
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
				hcl.TraverseIndex{
					Key: cty.NumberIntVal(1),
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 4, Byte: 3},
						End:   hcl.Pos{Line: 1, Column: 7, Byte: 6},
					},
				},
				hcl.TraverseIndex{
					Key: cty.NumberIntVal(2),
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
						End:   hcl.Pos{Line: 1, Column: 10, Byte: 9},
					},
				},
			},
			0,
		},
		{
			"foo[1].bar",
			hcl.Traversal{
				hcl.TraverseRoot{
					Name: "foo",
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
				hcl.TraverseIndex{
					Key: cty.NumberIntVal(1),
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 4, Byte: 3},
						End:   hcl.Pos{Line: 1, Column: 7, Byte: 6},
					},
				},
				hcl.TraverseAttr{
					Name: "bar",
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 7, Byte: 6},
						End:   hcl.Pos{Line: 1, Column: 11, Byte: 10},
					},
				},
			},
			0,
		},
		{
			"foo.",
			hcl.Traversal{
				hcl.TraverseRoot{
					Name: "foo",
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			1, // attribute name required
		},
		{
			"foo[",
			hcl.Traversal{
				hcl.TraverseRoot{
					Name: "foo",
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			1, // index required
		},
		{
			"foo[index]",
			hcl.Traversal{
				hcl.TraverseRoot{
					Name: "foo",
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			1, // index must be literal
		},
		{
			"foo[0",
			hcl.Traversal{
				hcl.TraverseRoot{
					Name: "foo",
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
				hcl.TraverseIndex{
					Key: cty.NumberIntVal(0),
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 4, Byte: 3},
						End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
					},
				},
			},
			1, // missing close bracket
		},
		{
			"foo 0",
			hcl.Traversal{
				hcl.TraverseRoot{
					Name: "foo",
					SrcRange: hcl.Range{
						Start: hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:   hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			1, // extra junk after traversal
		},
	}

	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			got, diags := ParseTraversalAbs([]byte(test.src), "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != test.diagCount {
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("wrong result\nsrc:  %s\ngot:  %s\nwant: %s", test.src, spew.Sdump(got), spew.Sdump(test.want))
			}
		})
	}
}
