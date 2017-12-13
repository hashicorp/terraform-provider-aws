package hcltest

import (
	"testing"

	"reflect"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
)

var mockBodyIsBody hcl.Body = mockBody{}
var mockExprLiteralIsExpr hcl.Expression = mockExprLiteral{}
var mockExprVariableIsExpr hcl.Expression = mockExprVariable("")

func TestMockBodyPartialContent(t *testing.T) {
	tests := map[string]struct {
		In        *hcl.BodyContent
		Schema    *hcl.BodySchema
		Want      *hcl.BodyContent
		Remain    *hcl.BodyContent
		DiagCount int
	}{
		"empty": {
			&hcl.BodyContent{},
			&hcl.BodySchema{},
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks:     hcl.Blocks{},
			},
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks:     hcl.Blocks{},
			},
			0,
		},
		"attribute requested": {
			&hcl.BodyContent{
				Attributes: MockAttrs(map[string]hcl.Expression{
					"name": MockExprLiteral(cty.StringVal("Ermintrude")),
				}),
			},
			&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name: "name",
					},
				},
			},
			&hcl.BodyContent{
				Attributes: MockAttrs(map[string]hcl.Expression{
					"name": MockExprLiteral(cty.StringVal("Ermintrude")),
				}),
				Blocks: hcl.Blocks{},
			},
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks:     hcl.Blocks{},
			},
			0,
		},
		"attribute remains": {
			&hcl.BodyContent{
				Attributes: MockAttrs(map[string]hcl.Expression{
					"name": MockExprLiteral(cty.StringVal("Ermintrude")),
				}),
			},
			&hcl.BodySchema{},
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks:     hcl.Blocks{},
			},
			&hcl.BodyContent{
				Attributes: MockAttrs(map[string]hcl.Expression{
					"name": MockExprLiteral(cty.StringVal("Ermintrude")),
				}),
				Blocks: hcl.Blocks{},
			},
			0,
		},
		"attribute missing": {
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
			},
			&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name:     "name",
						Required: true,
					},
				},
			},
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks:     hcl.Blocks{},
			},
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks:     hcl.Blocks{},
			},
			1, // missing attribute "name"
		},
		"block requested, no labels": {
			&hcl.BodyContent{
				Blocks: hcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: "baz",
					},
				},
			},
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks: hcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks:     hcl.Blocks{},
			},
			0,
		},
		"block requested, wrong labels": {
			&hcl.BodyContent{
				Blocks: hcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type:       "baz",
						LabelNames: []string{"foo"},
					},
				},
			},
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks: hcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks:     hcl.Blocks{},
			},
			1, // "baz" requires 1 label
		},
		"block remains": {
			&hcl.BodyContent{
				Blocks: hcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			&hcl.BodySchema{},
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks:     hcl.Blocks{},
			},
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks: hcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			0,
		},
		"various": {
			&hcl.BodyContent{
				Attributes: MockAttrs(map[string]hcl.Expression{
					"name": MockExprLiteral(cty.StringVal("Ermintrude")),
					"age":  MockExprLiteral(cty.NumberIntVal(32)),
				}),
				Blocks: hcl.Blocks{
					{
						Type: "baz",
					},
					{
						Type:   "bar",
						Labels: []string{"foo1"},
					},
					{
						Type:   "bar",
						Labels: []string{"foo2"},
					},
				},
			},
			&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name: "name",
					},
				},
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type:       "bar",
						LabelNames: []string{"name"},
					},
				},
			},
			&hcl.BodyContent{
				Attributes: MockAttrs(map[string]hcl.Expression{
					"name": MockExprLiteral(cty.StringVal("Ermintrude")),
				}),
				Blocks: hcl.Blocks{
					{
						Type:   "bar",
						Labels: []string{"foo1"},
					},
					{
						Type:   "bar",
						Labels: []string{"foo2"},
					},
				},
			},
			&hcl.BodyContent{
				Attributes: MockAttrs(map[string]hcl.Expression{
					"age": MockExprLiteral(cty.NumberIntVal(32)),
				}),
				Blocks: hcl.Blocks{
					{
						Type: "baz",
					},
				},
			},
			0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			inBody := MockBody(test.In)
			got, remainBody, diags := inBody.PartialContent(test.Schema)
			if len(diags) != test.DiagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.DiagCount)
				for _, diag := range diags {
					t.Logf("- %s", diag)
				}
			}

			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}

			gotRemain := remainBody.(mockBody).C
			if !reflect.DeepEqual(gotRemain, test.Remain) {
				t.Errorf("wrong remain\ngot:  %#v\nwant: %#v", gotRemain, test.Remain)
			}
		})
	}
}
