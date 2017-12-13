package include

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcltest"
	"github.com/zclconf/go-cty/cty"
)

func TestTransformer(t *testing.T) {
	caller := hcltest.MockBody(&hcl.BodyContent{
		Blocks: hcl.Blocks{
			{
				Type: "include",
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"path": hcltest.MockExprVariable("var_path"),
					}),
				}),
			},
			{
				Type: "include",
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"path": hcltest.MockExprLiteral(cty.StringVal("include2")),
					}),
				}),
			},
			{
				Type: "foo",
				Body: hcltest.MockBody(&hcl.BodyContent{
					Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
						"from": hcltest.MockExprLiteral(cty.StringVal("caller")),
					}),
				}),
			},
		},
	})

	resolver := MapResolver(map[string]hcl.Body{
		"include1": hcltest.MockBody(&hcl.BodyContent{
			Blocks: hcl.Blocks{
				{
					Type: "foo",
					Body: hcltest.MockBody(&hcl.BodyContent{
						Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
							"from": hcltest.MockExprLiteral(cty.StringVal("include1")),
						}),
					}),
				},
			},
		}),
		"include2": hcltest.MockBody(&hcl.BodyContent{
			Blocks: hcl.Blocks{
				{
					Type: "foo",
					Body: hcltest.MockBody(&hcl.BodyContent{
						Attributes: hcltest.MockAttrs(map[string]hcl.Expression{
							"from": hcltest.MockExprLiteral(cty.StringVal("include2")),
						}),
					}),
				},
			},
		}),
	})

	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var_path": cty.StringVal("include1"),
		},
	}

	transformer := Transformer("include", ctx, resolver)
	merged := transformer.TransformBody(caller)

	type foo struct {
		From string `hcl:"from,attr"`
	}
	type result struct {
		Foos []foo `hcl:"foo,block"`
	}
	var got result
	diags := gohcl.DecodeBody(merged, nil, &got)
	if len(diags) != 0 {
		t.Errorf("unexpected diags")
		for _, diag := range diags {
			t.Logf("- %s", diag)
		}
	}

	want := result{
		Foos: []foo{
			{
				From: "caller",
			},
			{
				From: "include1",
			},
			{
				From: "include2",
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("wrong result\ngot: %swant: %s", spew.Sdump(got), spew.Sdump(want))
	}
}
