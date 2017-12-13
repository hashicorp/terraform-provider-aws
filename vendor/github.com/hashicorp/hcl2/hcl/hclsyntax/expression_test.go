package hclsyntax

import (
	"testing"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func TestExpressionParseAndValue(t *testing.T) {
	// This is a combo test that exercises both the parser and the Value
	// method, with the focus on the latter but indirectly testing the former.
	tests := []struct {
		input     string
		ctx       *hcl.EvalContext
		want      cty.Value
		diagCount int
	}{
		{
			`1`,
			nil,
			cty.NumberIntVal(1),
			0,
		},
		{
			`(1)`,
			nil,
			cty.NumberIntVal(1),
			0,
		},
		{
			`(2+3)`,
			nil,
			cty.NumberIntVal(5),
			0,
		},
		{
			`(2+unk)`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.UnknownVal(cty.Number),
				},
			},
			cty.UnknownVal(cty.Number),
			0,
		},
		{
			`(2+unk)`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.DynamicVal,
				},
			},
			cty.UnknownVal(cty.Number),
			0,
		},
		{
			`(unk+unk)`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.DynamicVal,
				},
			},
			cty.UnknownVal(cty.Number),
			0,
		},
		{
			`(2+true)`,
			nil,
			cty.UnknownVal(cty.Number),
			1, // unsuitable type for right operand
		},
		{
			`(false+true)`,
			nil,
			cty.UnknownVal(cty.Number),
			2, // unsuitable type for each operand
		},
		{
			`(5 == 5)`,
			nil,
			cty.True,
			0,
		},
		{
			`(5 == 4)`,
			nil,
			cty.False,
			0,
		},
		{
			`(1 == true)`,
			nil,
			cty.False,
			0,
		},
		{
			`("true" == true)`,
			nil,
			cty.False,
			0,
		},
		{
			`(true == "true")`,
			nil,
			cty.False,
			0,
		},
		{
			`(true != "true")`,
			nil,
			cty.True,
			0,
		},
		{
			`(- 2)`,
			nil,
			cty.NumberIntVal(-2),
			0,
		},
		{
			`(! true)`,
			nil,
			cty.False,
			0,
		},
		{
			`(
    1
)`,
			nil,
			cty.NumberIntVal(1),
			0,
		},
		{
			`(1`,
			nil,
			cty.NumberIntVal(1),
			1, // Unbalanced parentheses
		},
		{
			`true`,
			nil,
			cty.True,
			0,
		},
		{
			`false`,
			nil,
			cty.False,
			0,
		},
		{
			`null`,
			nil,
			cty.NullVal(cty.DynamicPseudoType),
			0,
		},
		{
			`true true`,
			nil,
			cty.True,
			1, // extra characters after expression
		},
		{
			`"hello"`,
			nil,
			cty.StringVal("hello"),
			0,
		},
		{
			`"hello\nworld"`,
			nil,
			cty.StringVal("hello\nworld"),
			0,
		},
		{
			`"unclosed`,
			nil,
			cty.StringVal("unclosed"),
			1, // Unterminated template string
		},
		{
			`"hello ${"world"}"`,
			nil,
			cty.StringVal("hello world"),
			0,
		},
		{
			`"hello ${12.5}"`,
			nil,
			cty.StringVal("hello 12.5"),
			0,
		},
		{
			`"silly ${"${"nesting"}"}"`,
			nil,
			cty.StringVal("silly nesting"),
			0,
		},
		{
			`"silly ${"${true}"}"`,
			nil,
			cty.StringVal("silly true"),
			0,
		},
		{
			`"hello $${escaped}"`,
			nil,
			cty.StringVal("hello ${escaped}"),
			0,
		},
		{
			`"hello $$nonescape"`,
			nil,
			cty.StringVal("hello $$nonescape"),
			0,
		},
		{
			`upper("foo")`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.StringVal("FOO"),
			0,
		},
		{
			`
upper(
    "foo"
)
`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.StringVal("FOO"),
			0,
		},
		{
			`upper(["foo"]...)`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.StringVal("FOO"),
			0,
		},
		{
			`upper("foo", []...)`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.StringVal("FOO"),
			0,
		},
		{
			`upper("foo", "bar")`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.DynamicVal,
			1, // too many function arguments
		},
		{
			`upper(["foo", "bar"]...)`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.DynamicVal,
			1, // too many function arguments
		},
		{
			`[]`,
			nil,
			cty.EmptyTupleVal,
			0,
		},
		{
			`[1]`,
			nil,
			cty.TupleVal([]cty.Value{cty.NumberIntVal(1)}),
			0,
		},
		{
			`[1,]`,
			nil,
			cty.TupleVal([]cty.Value{cty.NumberIntVal(1)}),
			0,
		},
		{
			`[1,true]`,
			nil,
			cty.TupleVal([]cty.Value{cty.NumberIntVal(1), cty.True}),
			0,
		},
		{
			`[
  1,
  true
]`,
			nil,
			cty.TupleVal([]cty.Value{cty.NumberIntVal(1), cty.True}),
			0,
		},
		{
			`{}`,
			nil,
			cty.EmptyObjectVal,
			0,
		},
		{
			`{"hello": "world"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{"hello" = "world"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{hello = "world"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{hello: "world"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{"hello" = "world", "goodbye" = "cruel world"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello":   cty.StringVal("world"),
				"goodbye": cty.StringVal("cruel world"),
			}),
			0,
		},
		{
			`{
  "hello" = "world"
}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{
  "hello" = "world"
  "goodbye" = "cruel world"
}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello":   cty.StringVal("world"),
				"goodbye": cty.StringVal("cruel world"),
			}),
			0,
		},
		{
			`{
  "hello" = "world",
  "goodbye" = "cruel world"
}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello":   cty.StringVal("world"),
				"goodbye": cty.StringVal("cruel world"),
			}),
			0,
		},
		{
			`{
  "hello" = "world",
  "goodbye" = "cruel world",
}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello":   cty.StringVal("world"),
				"goodbye": cty.StringVal("cruel world"),
			}),
			0,
		},

		{
			`{for k, v in {hello: "world"}: k => v if k == "hello"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{for k, v in {hello: "world"}: upper(k) => upper(v) if k == "hello"}`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"upper": stdlib.UpperFunc,
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"HELLO": cty.StringVal("WORLD"),
			}),
			0,
		},
		{
			`{for k, v in ["world"]: k => v if k == 0}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"0": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{for v in ["world"]: v => v}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"world": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{for k, v in {hello: "world"}: k => v if k == "foo"}`,
			nil,
			cty.EmptyObjectVal,
			0,
		},
		{
			`{for k, v in {hello: "world"}: 5 => v}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"5": cty.StringVal("world"),
			}),
			0,
		},
		{
			`{for k, v in {hello: "world"}: [] => v}`,
			nil,
			cty.DynamicVal,
			1, // key expression has the wrong type
		},
		{
			`{for k, v in {hello: "world"}: k => k if k == "hello"}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("hello"),
			}),
			0,
		},
		{
			`{for k, v in {hello: "world"}: k => foo}`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.StringVal("foo"),
				},
			},
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("foo"),
			}),
			0,
		},
		{
			`[for k, v in {hello: "world"}: "${k}=${v}"]`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello=world"),
			}),
			0,
		},
		{
			`[for k, v in {hello: "world"}: k => v]`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"hello": cty.StringVal("world"),
			}),
			1, // can't have a key expr when producing a tuple
		},
		{
			`{for v in {hello: "world"}: v}`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("world"),
			}),
			1, // must have a key expr when producing a map
		},
		{
			`{for i, v in ["a", "b", "c", "b", "d"]: v => i...}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(0),
				}),
				"b": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(1),
					cty.NumberIntVal(3),
				}),
				"c": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(2),
				}),
				"d": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(4),
				}),
			}),
			0,
		},
		{
			`{for i, v in ["a", "b", "c", "b", "d"]: v => i... if i <= 2}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(0),
				}),
				"b": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(1),
				}),
				"c": cty.TupleVal([]cty.Value{
					cty.NumberIntVal(2),
				}),
			}),
			0,
		},
		{
			`{for i, v in ["a", "b", "c", "b", "d"]: v => i}`,
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.NumberIntVal(0),
				"b": cty.NumberIntVal(1),
				"c": cty.NumberIntVal(2),
				"d": cty.NumberIntVal(4),
			}),
			1, // duplicate key "b"
		},
		{
			`[for v in {hello: "world"}: v...]`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("world"),
			}),
			1, // can't use grouping when producing a tuple
		},
		{
			`[for v in "hello": v]`,
			nil,
			cty.DynamicVal,
			1, // can't iterate over a string
		},
		{
			`[for v in null: v]`,
			nil,
			cty.DynamicVal,
			1, // can't iterate over a null value
		},
		{
			`[for v in unk: v]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.UnknownVal(cty.List(cty.String)),
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`[for v in unk: v]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.DynamicVal,
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`[for v in unk: v]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.UnknownVal(cty.String),
				},
			},
			cty.DynamicVal,
			1, // can't iterate over a string (even if it's unknown)
		},
		{
			`[for v in ["a", "b"]: v if unkbool]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unkbool": cty.UnknownVal(cty.Bool),
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`[for v in ["a", "b"]: v if nullbool]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"nullbool": cty.NullVal(cty.Bool),
				},
			},
			cty.DynamicVal,
			1, // value of if clause must not be null
		},
		{
			`[for v in ["a", "b"]: v if dyn]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"dyn": cty.DynamicVal,
				},
			},
			cty.DynamicVal,
			0,
		},
		{
			`[for v in ["a", "b"]: v if unknum]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unknum": cty.UnknownVal(cty.List(cty.Number)),
				},
			},
			cty.DynamicVal,
			1, // if expression must be bool
		},
		{
			`[for i, v in ["a", "b"]: v if i + i]`,
			nil,
			cty.DynamicVal,
			1, // if expression must be bool
		},
		{
			`[for v in ["a", "b"]: unkstr]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unkstr": cty.UnknownVal(cty.String),
				},
			},
			cty.TupleVal([]cty.Value{
				cty.UnknownVal(cty.String),
				cty.UnknownVal(cty.String),
			}),
			0,
		},

		{
			`[{name: "Steve"}, {name: "Ermintrude"}].*.name`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("Steve"),
				cty.StringVal("Ermintrude"),
			}),
			0,
		},
		{
			`{name: "Steve"}.*.name`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("Steve"),
			}),
			0,
		},
		{
			`["hello", "goodbye"].*`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
				cty.StringVal("goodbye"),
			}),
			0,
		},
		{
			`"hello".*`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
			}),
			0,
		},
		{
			`[["hello"], ["world", "unused"]].*.0`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("hello"),
				cty.StringVal("world"),
			}),
			0,
		},
		{
			`[[{name:"foo"}], [{name:"bar"}, {name:"baz"}]].*.0.name`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.StringVal("foo"),
				cty.StringVal("bar"),
			}),
			0,
		},
		{
			// For an "attribute-only" splat, an index operator applies to
			// the splat result as a whole, rather than being incorporated
			// into the splat traversal itself.
			`[{name: "Steve"}, {name: "Ermintrude"}].*.name[0]`,
			nil,
			cty.StringVal("Steve"),
			0,
		},
		{
			`[["hello"], ["goodbye"]].*.*`,
			nil,
			cty.TupleVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("hello")}),
				cty.TupleVal([]cty.Value{cty.StringVal("goodbye")}),
			}),
			1,
		},

		{
			`["hello"][0]`,
			nil,
			cty.StringVal("hello"),
			0,
		},
		{
			`[][0]`,
			nil,
			cty.DynamicVal,
			1, // invalid index
		},
		{
			`["hello"][negate(0)]`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"negate": stdlib.NegateFunc,
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`[][negate(0)]`,
			&hcl.EvalContext{
				Functions: map[string]function.Function{
					"negate": stdlib.NegateFunc,
				},
			},
			cty.DynamicVal,
			1, // invalid index
		},
		{
			`["hello"]["0"]`, // key gets converted to number
			nil,
			cty.StringVal("hello"),
			0,
		},

		{
			`foo`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.StringVal("hello"),
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`bar`,
			&hcl.EvalContext{},
			cty.DynamicVal,
			1, // variables not allowed here
		},
		{
			`foo.bar`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.StringVal("hello"),
				},
			},
			cty.DynamicVal,
			1, // foo does not have attributes
		},
		{
			`foo.baz`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.ObjectVal(map[string]cty.Value{
						"baz": cty.StringVal("hello"),
					}),
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`foo["baz"]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.ObjectVal(map[string]cty.Value{
						"baz": cty.StringVal("hello"),
					}),
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`foo[true]`, // key is converted to string
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.ObjectVal(map[string]cty.Value{
						"true": cty.StringVal("hello"),
					}),
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`foo[0].baz`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"foo": cty.ListVal([]cty.Value{
						cty.ObjectVal(map[string]cty.Value{
							"baz": cty.StringVal("hello"),
						}),
					}),
				},
			},
			cty.StringVal("hello"),
			0,
		},
		{
			`unk["baz"]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.UnknownVal(cty.String),
				},
			},
			cty.DynamicVal,
			1, // value does not have indices (because we know it's a string)
		},
		{
			`unk["boop"]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"unk": cty.UnknownVal(cty.Map(cty.String)),
				},
			},
			cty.UnknownVal(cty.String), // we know it's a map of string
			0,
		},
		{
			`dyn["boop"]`,
			&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"dyn": cty.DynamicVal,
				},
			},
			cty.DynamicVal, // don't know what it is yet
			0,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			expr, parseDiags := ParseExpression([]byte(test.input), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})

			got, valDiags := expr.Value(test.ctx)

			diagCount := len(parseDiags) + len(valDiags)

			if diagCount != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", diagCount, test.diagCount)
				for _, diag := range parseDiags {
					t.Logf(" - %s", diag.Error())
				}
				for _, diag := range valDiags {
					t.Logf(" - %s", diag.Error())
				}
			}

			if !got.RawEquals(test.want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.want)
			}
		})
	}

}

func TestFunctionCallExprValue(t *testing.T) {
	funcs := map[string]function.Function{
		"length":     stdlib.StrlenFunc,
		"jsondecode": stdlib.JSONDecodeFunc,
	}

	tests := map[string]struct {
		expr      *FunctionCallExpr
		ctx       *hcl.EvalContext
		want      cty.Value
		diagCount int
	}{
		"valid call with no conversions": {
			&FunctionCallExpr{
				Name: "length",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.StringVal("hello"),
					},
				},
			},
			&hcl.EvalContext{
				Functions: funcs,
			},
			cty.NumberIntVal(5),
			0,
		},
		"valid call with arg conversion": {
			&FunctionCallExpr{
				Name: "length",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.BoolVal(true),
					},
				},
			},
			&hcl.EvalContext{
				Functions: funcs,
			},
			cty.NumberIntVal(4), // length of string "true"
			0,
		},
		"valid call with unknown arg": {
			&FunctionCallExpr{
				Name: "length",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.UnknownVal(cty.String),
					},
				},
			},
			&hcl.EvalContext{
				Functions: funcs,
			},
			cty.UnknownVal(cty.Number),
			0,
		},
		"valid call with unknown arg needing conversion": {
			&FunctionCallExpr{
				Name: "length",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.UnknownVal(cty.Bool),
					},
				},
			},
			&hcl.EvalContext{
				Functions: funcs,
			},
			cty.UnknownVal(cty.Number),
			0,
		},
		"valid call with dynamic arg": {
			&FunctionCallExpr{
				Name: "length",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.DynamicVal,
					},
				},
			},
			&hcl.EvalContext{
				Functions: funcs,
			},
			cty.UnknownVal(cty.Number),
			0,
		},
		"invalid arg type": {
			&FunctionCallExpr{
				Name: "length",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.ListVal([]cty.Value{cty.StringVal("hello")}),
					},
				},
			},
			&hcl.EvalContext{
				Functions: funcs,
			},
			cty.DynamicVal,
			1,
		},
		"function with dynamic return type": {
			&FunctionCallExpr{
				Name: "jsondecode",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.StringVal(`"hello"`),
					},
				},
			},
			&hcl.EvalContext{
				Functions: funcs,
			},
			cty.StringVal("hello"),
			0,
		},
		"function with dynamic return type unknown arg": {
			&FunctionCallExpr{
				Name: "jsondecode",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.UnknownVal(cty.String),
					},
				},
			},
			&hcl.EvalContext{
				Functions: funcs,
			},
			cty.DynamicVal, // type depends on arg value
			0,
		},
		"error in function": {
			&FunctionCallExpr{
				Name: "jsondecode",
				Args: []Expression{
					&LiteralValueExpr{
						Val: cty.StringVal("invalid-json"),
					},
				},
			},
			&hcl.EvalContext{
				Functions: funcs,
			},
			cty.DynamicVal,
			1, // JSON parse error
		},
		"unknown function": {
			&FunctionCallExpr{
				Name: "lenth",
				Args: []Expression{},
			},
			&hcl.EvalContext{
				Functions: funcs,
			},
			cty.DynamicVal,
			1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, diags := test.expr.Value(test.ctx)

			if len(diags) != test.diagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.diagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}

			if !got.RawEquals(test.want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.want)
			}
		})
	}
}
