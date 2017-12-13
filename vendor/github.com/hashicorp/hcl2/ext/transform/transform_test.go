package transform

import (
	"testing"

	"reflect"

	"github.com/hashicorp/hcl2/hcltest"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
)

// Assert that deepWrapper implements Body
var deepWrapperIsBody hcl.Body = deepWrapper{}

func TestDeep(t *testing.T) {

	testTransform := TransformerFunc(func(body hcl.Body) hcl.Body {
		_, remain, diags := body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type: "remove",
				},
			},
		})

		return BodyWithDiagnostics(remain, diags)
	})

	src := hcltest.MockBody(&hcl.BodyContent{
		Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
			"true": hcltest.MockExprLiteral(cty.True),
		}),
		Blocks: []*hcl.Block{
			{
				Type: "remove",
				Body: hcl.EmptyBody(),
			},
			{
				Type: "child",
				Body: hcltest.MockBody(&hcl.BodyContent{
					Blocks: []*hcl.Block{
						{
							Type: "remove",
						},
					},
				}),
			},
		},
	})

	wrapped := Deep(src, testTransform)

	rootContent, diags := wrapped.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: "true",
			},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "child",
			},
		},
	})
	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics for root content")
		for _, diag := range diags {
			t.Logf("- %s", diag)
		}
	}

	wantAttrs := hcltest.MockAttrs(map[string]hcl.Expression{
		"true": hcltest.MockExprLiteral(cty.True),
	})
	if !reflect.DeepEqual(rootContent.Attributes, wantAttrs) {
		t.Errorf("wrong root attributes\ngot:  %#v\nwant: %#v", rootContent.Attributes, wantAttrs)
	}

	if got, want := len(rootContent.Blocks), 1; got != want {
		t.Fatalf("wrong number of root blocks %d; want %d", got, want)
	}
	if got, want := rootContent.Blocks[0].Type, "child"; got != want {
		t.Errorf("wrong block type %s; want %s", got, want)
	}

	childBlock := rootContent.Blocks[0]
	childContent, diags := childBlock.Body.Content(&hcl.BodySchema{})
	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics for child content")
		for _, diag := range diags {
			t.Logf("- %s", diag)
		}
	}

	if len(childContent.Attributes) != 0 {
		t.Errorf("unexpected attributes in child content; want empty content")
	}
	if len(childContent.Blocks) != 0 {
		t.Errorf("unexpected blocks in child content; want empty content")
	}
}
