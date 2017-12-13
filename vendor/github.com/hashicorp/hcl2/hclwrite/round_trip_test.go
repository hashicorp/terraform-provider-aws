package hclwrite

import (
	"bytes"
	"testing"

	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func TestRoundTripVerbatim(t *testing.T) {
	tests := []string{
		``,
		`foo = 1
`,
		`
foobar = 1
baz    = 1
`,
		`
# this file is awesome

# tossed salads and scrambled eggs
foobar = 1
baz    = 1

block {
  a = "a"
  b = "b"
  c = "c"
  d = "d"

  subblock {
  }

  subblock {
    e = "e"
  }
}

# and they all lived happily ever after
`,
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			src := []byte(test)
			file, diags := parse(src, "", hcl.Pos{Line: 1, Column: 1})
			if len(diags) != 0 {
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
				t.Fatalf("unexpected diagnostics")
			}

			wr := &bytes.Buffer{}
			n, err := file.WriteTo(wr)
			if n != len(test) {
				t.Errorf("wrong number of bytes %d; want %d", n, len(test))
			}
			if err != nil {
				t.Fatalf("error from WriteTo")
			}

			result := wr.Bytes()

			if !bytes.Equal(result, src) {
				t.Errorf("wrong result\nresult:\n%s\ninput:\n%s", result, src)
			}
		})
	}
}

func TestRoundTripFormat(t *testing.T) {
	// The goal of this test is to verify that the formatter doesn't change
	// the semantics of any expressions when it adds and removes whitespace.
	// String templates are the primary area of concern here, but we also
	// test some other things for completeness sake.
	//
	// The tests here must define zero or more attributes, which will be
	// extract with JustAttributes and evaluated both before and after
	// formatting.

	tests := []string{
		"",
		"\n\n\n",
		"a=1\n",
		"a=\"hello\"\n",
		"a=\"${hello} world\"\n",
		"a=upper(\"hello\")\n",
		"a=upper(hello)\n",
		"a=[1,2,3,4,five]\n",
		"a={greeting=hello}\n",
		"a={\ngreeting=hello\n}\n",
		"a={\ngreeting=hello}\n",
		"a={greeting=hello\n}\n",
		"a={greeting=hello,number=five,sarcastic=\"${upper(hello)}\"\n}\n",
		"a={\ngreeting=hello\nnumber=five\nsarcastic=\"${upper(hello)}\"\n}\n",
		"a=<<EOT\nhello\nEOT\n\n",
		"a=[<<EOT\nhello\nEOT\n]\n",
		"a=[\n<<EOT\nhello\nEOT\n]\n",
		"a=[\n]\n",
		"a=1\nb=2\nc=3\n",
		"a=\"${\n5\n}\"\n",
	}

	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"hello": cty.StringVal("hello"),
			"five":  cty.NumberIntVal(5),
		},
		Functions: map[string]function.Function{
			"upper": stdlib.UpperFunc,
		},
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {

			attrsAsObj := func(src []byte, phase string) cty.Value {
				t.Logf("source %s:\n%s", phase, src)
				f, diags := hclsyntax.ParseConfig(src, "", hcl.Pos{Line: 1, Column: 1})
				if len(diags) != 0 {
					for _, diag := range diags {
						t.Logf(" - %s", diag.Error())
					}
					t.Fatalf("unexpected diagnostics in parse %s", phase)
				}

				attrs, diags := f.Body.JustAttributes()
				if len(diags) != 0 {
					for _, diag := range diags {
						t.Logf(" - %s", diag.Error())
					}
					t.Fatalf("unexpected diagnostics in JustAttributes %s", phase)
				}

				vals := map[string]cty.Value{}
				for k, attr := range attrs {
					val, diags := attr.Expr.Value(ctx)
					if len(diags) != 0 {
						for _, diag := range diags {
							t.Logf(" - %s", diag.Error())
						}
						t.Fatalf("unexpected diagnostics evaluating %s", phase)
					}
					vals[k] = val
				}
				return cty.ObjectVal(vals)
			}

			src := []byte(test)
			before := attrsAsObj(src, "before")

			formatted := Format(src)
			after := attrsAsObj(formatted, "after")

			if !after.RawEquals(before) {
				t.Errorf("mismatching after format\nbefore: %#v\nafter:  %#v", before, after)
			}
		})
	}

}
