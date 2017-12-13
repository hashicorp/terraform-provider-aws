package userfunc

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
)

func TestDecodeUserFunctions(t *testing.T) {
	tests := []struct {
		src       string
		testExpr  string
		baseCtx   *hcl.EvalContext
		want      cty.Value
		diagCount int
	}{
		{
			`
function "greet" {
  params = ["name"]
  result = "Hello, ${name}."
}
`,
			`greet("Ermintrude")`,
			nil,
			cty.StringVal("Hello, Ermintrude."),
			0,
		},
		{
			`
function "greet" {
  params = ["name"]
  result = "Hello, ${name}."
}
`,
			`greet()`,
			nil,
			cty.DynamicVal,
			1, // missing value for "name"
		},
		{
			`
function "greet" {
  params = ["name"]
  result = "Hello, ${name}."
}
`,
			`greet("Ermintrude", "extra")`,
			nil,
			cty.DynamicVal,
			1, // too many arguments
		},
		{
			`
function "add" {
  params = ["a", "b"]
  result = a + b
}
`,
			`add(1, 5)`,
			nil,
			cty.NumberIntVal(6),
			0,
		},
		{
			`
function "argstuple" {
  params = []
  variadic_param = "args"
  result = args
}
`,
			`argstuple("a", true, 1)`,
			nil,
			cty.TupleVal([]cty.Value{cty.StringVal("a"), cty.True, cty.NumberIntVal(1)}),
			0,
		},
		{
			`
function "missing_var" {
  params = []
  result = nonexist
}
`,
			`missing_var()`,
			nil,
			cty.DynamicVal,
			1, // no variable named "nonexist"
		},
		{
			`
function "closure" {
  params = []
  result = upvalue
}
`,
			`closure()`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"upvalue": cty.True,
				},
			},
			cty.True,
			0,
		},
		{
			`
function "neg" {
  params = ["val"]
  result = -val
}
function "add" {
  params = ["a", "b"]
  result = a + b
}
`,
			`neg(add(1, 3))`,
			nil,
			cty.NumberIntVal(-4),
			0,
		},
		{
			`
function "neg" {
  parrams = ["val"]
  result = -val
}
`,
			`null`,
			nil,
			cty.NullVal(cty.DynamicPseudoType),
			2, // missing attribute "params", and unknown attribute "parrams"
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			f, diags := hclsyntax.ParseConfig([]byte(test.src), "config", hcl.Pos{Line: 1, Column: 1})
			if f == nil || f.Body == nil {
				t.Fatalf("got nil file or body")
			}

			funcs, _, funcsDiags := decodeUserFunctions(f.Body, "function", func() *hcl.EvalContext {
				return test.baseCtx
			})
			diags = append(diags, funcsDiags...)

			expr, exprParseDiags := hclsyntax.ParseExpression([]byte(test.testExpr), "testexpr", hcl.Pos{Line: 1, Column: 1})
			diags = append(diags, exprParseDiags...)
			if expr == nil {
				t.Fatalf("parsing test expr returned nil")
			}

			got, exprDiags := expr.Value(&hcl.EvalContext{
				Functions: funcs,
			})
			diags = append(diags, exprDiags...)

			if len(diags) != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf("- %s", diag)
				}
			}

			if !got.RawEquals(test.want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.want)
			}
		})
	}
}
