package hclsyntax

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/kylelemons/godebug/pretty"
	"github.com/zclconf/go-cty/cty"
)

func TestBodyContent(t *testing.T) {
	tests := []struct {
		body      *Body
		schema    *hcl.BodySchema
		partial   bool
		want      *hcl.BodyContent
		diagCount int
	}{
		{
			&Body{},
			&hcl.BodySchema{},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
			},
			0,
		},

		// Attributes
		{
			&Body{
				Attributes: Attributes{
					"foo": &Attribute{
						Name: "foo",
					},
				},
			},
			&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name: "foo",
					},
				},
			},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{
					"foo": &hcl.Attribute{
						Name: "foo",
					},
				},
			},
			0,
		},
		{
			&Body{
				Attributes: Attributes{
					"foo": &Attribute{
						Name: "foo",
					},
				},
			},
			&hcl.BodySchema{},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
			},
			1, // attribute "foo" is not expected
		},
		{
			&Body{
				Attributes: Attributes{
					"foo": &Attribute{
						Name: "foo",
					},
				},
			},
			&hcl.BodySchema{},
			true,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
			},
			0, // in partial mode, so extra "foo" is acceptable
		},
		{
			&Body{
				Attributes: Attributes{},
			},
			&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name: "foo",
					},
				},
			},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
			},
			0, // "foo" not required, so no error
		},
		{
			&Body{
				Attributes: Attributes{},
			},
			&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{
						Name:     "foo",
						Required: true,
					},
				},
			},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
			},
			1, // "foo" is required
		},
		{
			&Body{
				Attributes: Attributes{
					"foo": &Attribute{
						Name: "foo",
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: "foo",
					},
				},
			},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
			},
			1, // attribute "foo" not expected (it's defined as a block)
		},

		// Blocks
		{
			&Body{
				Blocks: Blocks{
					&Block{
						Type: "foo",
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: "foo",
					},
				},
			},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks: hcl.Blocks{
					{
						Type: "foo",
						Body: (*Body)(nil),
					},
				},
			},
			0,
		},
		{
			&Body{
				Blocks: Blocks{
					&Block{
						Type: "foo",
					},
					&Block{
						Type: "foo",
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: "foo",
					},
				},
			},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks: hcl.Blocks{
					{
						Type: "foo",
						Body: (*Body)(nil),
					},
					{
						Type: "foo",
						Body: (*Body)(nil),
					},
				},
			},
			0,
		},
		{
			&Body{
				Blocks: Blocks{
					&Block{
						Type: "foo",
					},
					&Block{
						Type: "bar",
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: "foo",
					},
				},
			},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks: hcl.Blocks{
					{
						Type: "foo",
						Body: (*Body)(nil),
					},
				},
			},
			1, // blocks of type "bar" not expected
		},
		{
			&Body{
				Blocks: Blocks{
					&Block{
						Type: "foo",
					},
					&Block{
						Type: "bar",
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: "foo",
					},
				},
			},
			true,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks: hcl.Blocks{
					{
						Type: "foo",
						Body: (*Body)(nil),
					},
				},
			},
			0, // extra "bar" allowed because we're in partial mode
		},
		{
			&Body{
				Blocks: Blocks{
					&Block{
						Type:   "foo",
						Labels: []string{"bar"},
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type:       "foo",
						LabelNames: []string{"name"},
					},
				},
			},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
				Blocks: hcl.Blocks{
					{
						Type:   "foo",
						Labels: []string{"bar"},
						Body:   (*Body)(nil),
					},
				},
			},
			0,
		},
		{
			&Body{
				Blocks: Blocks{
					&Block{
						Type: "foo",
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type:       "foo",
						LabelNames: []string{"name"},
					},
				},
			},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
			},
			1, // missing label "name"
		},
		{
			&Body{
				Blocks: Blocks{
					&Block{
						Type:   "foo",
						Labels: []string{"bar"},

						LabelRanges: []hcl.Range{{}},
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: "foo",
					},
				},
			},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
			},
			1, // no labels expected
		},
		{
			&Body{
				Blocks: Blocks{
					&Block{
						Type:   "foo",
						Labels: []string{"bar", "baz"},

						LabelRanges: []hcl.Range{{}, {}},
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type:       "foo",
						LabelNames: []string{"name"},
					},
				},
			},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
			},
			1, // too many labels
		},
		{
			&Body{
				Attributes: Attributes{
					"foo": &Attribute{
						Name: "foo",
					},
				},
			},
			&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type: "foo",
					},
				},
			},
			false,
			&hcl.BodyContent{
				Attributes: hcl.Attributes{},
			},
			1, // should've been a block, not an attribute
		},
	}

	prettyConfig := &pretty.Config{
		Diffable:          true,
		IncludeUnexported: true,
		PrintStringers:    true,
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			var got *hcl.BodyContent
			var diags hcl.Diagnostics
			if test.partial {
				got, _, diags = test.body.PartialContent(test.schema)
			} else {
				got, diags = test.body.Content(test.schema)
			}

			if len(diags) != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf(
					"wrong result\ndiff: %s",
					prettyConfig.Compare(test.want, got),
				)
			}
		})
	}
}

func TestBodyJustAttributes(t *testing.T) {
	tests := []struct {
		body      *Body
		want      hcl.Attributes
		diagCount int
	}{
		{
			&Body{},
			hcl.Attributes{},
			0,
		},
		{
			&Body{
				Attributes: Attributes{},
			},
			hcl.Attributes{},
			0,
		},
		{
			&Body{
				Attributes: Attributes{
					"foo": &Attribute{
						Name: "foo",
						Expr: &LiteralValueExpr{
							Val: cty.StringVal("bar"),
						},
					},
				},
			},
			hcl.Attributes{
				"foo": &hcl.Attribute{
					Name: "foo",
					Expr: &LiteralValueExpr{
						Val: cty.StringVal("bar"),
					},
				},
			},
			0,
		},
		{
			&Body{
				Attributes: Attributes{
					"foo": &Attribute{
						Name: "foo",
						Expr: &LiteralValueExpr{
							Val: cty.StringVal("bar"),
						},
					},
				},
				Blocks: Blocks{
					{
						Type: "foo",
					},
				},
			},
			hcl.Attributes{
				"foo": &hcl.Attribute{
					Name: "foo",
					Expr: &LiteralValueExpr{
						Val: cty.StringVal("bar"),
					},
				},
			},
			1, // blocks are not allowed here
		},
		{
			&Body{
				Attributes: Attributes{
					"foo": &Attribute{
						Name: "foo",
						Expr: &LiteralValueExpr{
							Val: cty.StringVal("bar"),
						},
					},
				},
				hiddenAttrs: map[string]struct{}{
					"foo": struct{}{},
				},
			},
			hcl.Attributes{},
			0,
		},
	}

	prettyConfig := &pretty.Config{
		Diffable:          true,
		IncludeUnexported: true,
		PrintStringers:    true,
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			got, diags := test.body.JustAttributes()

			if len(diags) != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf(
					"wrong result\nbody: %s\ndiff: %s",
					prettyConfig.Sprint(test.body),
					prettyConfig.Compare(test.want, got),
				)
			}
		})
	}
}
