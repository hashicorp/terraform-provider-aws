package hcldec

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		config    string
		spec      Spec
		ctx       *hcl.EvalContext
		want      cty.Value
		diagCount int
	}{
		{
			``,
			&ObjectSpec{},
			nil,
			cty.EmptyObjectVal,
			0,
		},
		{
			"a = 1\n",
			&ObjectSpec{},
			nil,
			cty.EmptyObjectVal,
			1, // attribute named "a" is not expected here
		},
		{
			"a = 1\n",
			&ObjectSpec{
				"a": &AttrSpec{
					Name: "a",
					Type: cty.Number,
				},
			},
			nil,
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.NumberIntVal(1),
			}),
			0,
		},
		{
			"a = 1\n",
			&AttrSpec{
				Name: "a",
				Type: cty.Number,
			},
			nil,
			cty.NumberIntVal(1),
			0,
		},
		{
			"a = 1\n",
			&DefaultSpec{
				Primary: &AttrSpec{
					Name: "a",
					Type: cty.Number,
				},
				Default: &LiteralSpec{
					Value: cty.NumberIntVal(10),
				},
			},
			nil,
			cty.NumberIntVal(1),
			0,
		},
		{
			"",
			&DefaultSpec{
				Primary: &AttrSpec{
					Name: "a",
					Type: cty.Number,
				},
				Default: &LiteralSpec{
					Value: cty.NumberIntVal(10),
				},
			},
			nil,
			cty.NumberIntVal(10),
			0,
		},
		{
			"a = 1\n",
			ObjectSpec{
				"foo": &DefaultSpec{
					Primary: &AttrSpec{
						Name: "a",
						Type: cty.Number,
					},
					Default: &LiteralSpec{
						Value: cty.NumberIntVal(10),
					},
				},
			},
			nil,
			cty.ObjectVal(map[string]cty.Value{"foo": cty.NumberIntVal(1)}),
			0,
		},
		{
			"a = \"1\"\n",
			&AttrSpec{
				Name: "a",
				Type: cty.Number,
			},
			nil,
			cty.NumberIntVal(1),
			0,
		},
		{
			"a = true\n",
			&AttrSpec{
				Name: "a",
				Type: cty.Number,
			},
			nil,
			cty.UnknownVal(cty.Number),
			1, // incorrect type - number required.
		},
		{
			``,
			&AttrSpec{
				Name:     "a",
				Type:     cty.Number,
				Required: true,
			},
			nil,
			cty.NullVal(cty.Number),
			1, // attribute "a" is required
		},

		{
			`
b {
}
`,
			&BlockSpec{
				TypeName: "b",
				Nested:   ObjectSpec{},
			},
			nil,
			cty.EmptyObjectVal,
			0,
		},
		{
			`
b "baz" {
}
`,
			&BlockSpec{
				TypeName: "b",
				Nested: &BlockLabelSpec{
					Index: 0,
					Name:  "name",
				},
			},
			nil,
			cty.StringVal("baz"),
			0,
		},
		{
			`
b "baz" {}
b "foo" {}
`,
			&BlockSpec{
				TypeName: "b",
				Nested: &BlockLabelSpec{
					Index: 0,
					Name:  "name",
				},
			},
			nil,
			cty.StringVal("baz"),
			1, // duplicate "b" block
		},
		{
			`
b {
}
`,
			&BlockSpec{
				TypeName: "b",
				Nested: &BlockLabelSpec{
					Index: 0,
					Name:  "name",
				},
			},
			nil,
			cty.NullVal(cty.String),
			1, // missing name label
		},
		{
			``,
			&BlockSpec{
				TypeName: "b",
				Nested:   ObjectSpec{},
			},
			nil,
			cty.NullVal(cty.EmptyObject),
			0,
		},
		{
			"a {}\n",
			&BlockSpec{
				TypeName: "b",
				Nested:   ObjectSpec{},
			},
			nil,
			cty.NullVal(cty.EmptyObject),
			1, // blocks of type "a" are not supported
		},
		{
			``,
			&BlockSpec{
				TypeName: "b",
				Nested:   ObjectSpec{},
				Required: true,
			},
			nil,
			cty.NullVal(cty.EmptyObject),
			1, // a block of type "b" is required
		},
		{
			`
b {}
b {}
`,
			&BlockSpec{
				TypeName: "b",
				Nested:   ObjectSpec{},
				Required: true,
			},
			nil,
			cty.EmptyObjectVal,
			1, // only one "b" block is allowed
		},
		{
			`
b {}
b {}
`,
			&BlockListSpec{
				TypeName: "b",
				Nested:   ObjectSpec{},
			},
			nil,
			cty.ListVal([]cty.Value{cty.EmptyObjectVal, cty.EmptyObjectVal}),
			0,
		},
		{
			``,
			&BlockListSpec{
				TypeName: "b",
				Nested:   ObjectSpec{},
			},
			nil,
			cty.ListValEmpty(cty.EmptyObject),
			0,
		},
		{
			`
b "foo" {}
b "bar" {}
`,
			&BlockListSpec{
				TypeName: "b",
				Nested: &BlockLabelSpec{
					Name:  "name",
					Index: 0,
				},
			},
			nil,
			cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}),
			0,
		},
		{
			`
b {}
b {}
b {}
`,
			&BlockListSpec{
				TypeName: "b",
				Nested:   ObjectSpec{},
				MaxItems: 2,
			},
			nil,
			cty.ListVal([]cty.Value{cty.EmptyObjectVal, cty.EmptyObjectVal, cty.EmptyObjectVal}),
			1, // too many b blocks
		},
		{
			`
b {}
b {}
`,
			&BlockListSpec{
				TypeName: "b",
				Nested:   ObjectSpec{},
				MinItems: 10,
			},
			nil,
			cty.ListVal([]cty.Value{cty.EmptyObjectVal, cty.EmptyObjectVal}),
			1, // insufficient b blocks
		},
		{
			`
b {}
b {}
`,
			&BlockSetSpec{
				TypeName: "b",
				Nested:   ObjectSpec{},
				MaxItems: 2,
			},
			nil,
			cty.SetVal([]cty.Value{cty.EmptyObjectVal, cty.EmptyObjectVal}),
			0,
		},
		{
			`
b "foo" "bar" {}
b "bar" "baz" {}
`,
			&BlockSetSpec{
				TypeName: "b",
				Nested: TupleSpec{
					&BlockLabelSpec{
						Name:  "name",
						Index: 1,
					},
					&BlockLabelSpec{
						Name:  "type",
						Index: 0,
					},
				},
			},
			nil,
			cty.SetVal([]cty.Value{
				cty.TupleVal([]cty.Value{cty.StringVal("bar"), cty.StringVal("foo")}),
				cty.TupleVal([]cty.Value{cty.StringVal("baz"), cty.StringVal("bar")}),
			}),
			0,
		},
		{
			`
b "foo" {}
b "bar" {}
`,
			&BlockMapSpec{
				TypeName:   "b",
				LabelNames: []string{"key"},
				Nested:     ObjectSpec{},
			},
			nil,
			cty.MapVal(map[string]cty.Value{"foo": cty.EmptyObjectVal, "bar": cty.EmptyObjectVal}),
			0,
		},
		{
			`
b "foo" "bar" {}
b "bar" "baz" {}
`,
			&BlockMapSpec{
				TypeName:   "b",
				LabelNames: []string{"key1", "key2"},
				Nested:     ObjectSpec{},
			},
			nil,
			cty.MapVal(map[string]cty.Value{
				"foo": cty.MapVal(map[string]cty.Value{
					"bar": cty.EmptyObjectVal,
				}),
				"bar": cty.MapVal(map[string]cty.Value{
					"baz": cty.EmptyObjectVal,
				}),
			}),
			0,
		},
		{
			`
b "foo" "bar" {}
b "bar" "bar" {}
`,
			&BlockMapSpec{
				TypeName:   "b",
				LabelNames: []string{"key1", "key2"},
				Nested:     ObjectSpec{},
			},
			nil,
			cty.MapVal(map[string]cty.Value{
				"foo": cty.MapVal(map[string]cty.Value{
					"bar": cty.EmptyObjectVal,
				}),
				"bar": cty.MapVal(map[string]cty.Value{
					"bar": cty.EmptyObjectVal,
				}),
			}),
			0,
		},
		{
			`
b "foo" "bar" {}
b "foo" "baz" {}
`,
			&BlockMapSpec{
				TypeName:   "b",
				LabelNames: []string{"key1", "key2"},
				Nested:     ObjectSpec{},
			},
			nil,
			cty.MapVal(map[string]cty.Value{
				"foo": cty.MapVal(map[string]cty.Value{
					"bar": cty.EmptyObjectVal,
					"baz": cty.EmptyObjectVal,
				}),
			}),
			0,
		},
		{
			`
b "foo" "bar" {}
`,
			&BlockMapSpec{
				TypeName:   "b",
				LabelNames: []string{"key"},
				Nested:     ObjectSpec{},
			},
			nil,
			cty.MapValEmpty(cty.EmptyObject),
			1, // too many labels
		},
		{
			`
b "bar" {}
`,
			&BlockMapSpec{
				TypeName:   "b",
				LabelNames: []string{"key1", "key2"},
				Nested:     ObjectSpec{},
			},
			nil,
			cty.MapValEmpty(cty.EmptyObject),
			1, // not enough labels
		},
		{
			`
b "foo" {}
b "foo" {}
`,
			&BlockMapSpec{
				TypeName:   "b",
				LabelNames: []string{"key"},
				Nested:     ObjectSpec{},
			},
			nil,
			cty.MapVal(map[string]cty.Value{"foo": cty.EmptyObjectVal}),
			1, // duplicate b block
		},
		{
			`
b "foo" "bar" {}
b "foo" "bar" {}
`,
			&BlockMapSpec{
				TypeName:   "b",
				LabelNames: []string{"key1", "key2"},
				Nested:     ObjectSpec{},
			},
			nil,
			cty.MapVal(map[string]cty.Value{"foo": cty.MapVal(map[string]cty.Value{"bar": cty.EmptyObjectVal})}),
			1, // duplicate b block
		},
		{
			`
b "foo" "bar" {}
b "bar" "baz" {}
`,
			&BlockMapSpec{
				TypeName:   "b",
				LabelNames: []string{"type"},
				Nested: &BlockLabelSpec{
					Name:  "name",
					Index: 0,
				},
			},
			nil,
			cty.MapVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
				"bar": cty.StringVal("baz"),
			}),
			0,
		},
		{
			`
b "foo" {}
`,
			&BlockMapSpec{
				TypeName:   "b",
				LabelNames: []string{"type"},
				Nested: &BlockLabelSpec{
					Name:  "name",
					Index: 0,
				},
			},
			nil,
			cty.MapValEmpty(cty.String),
			1, // missing name
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d-%s", i, test.config), func(t *testing.T) {
			file, parseDiags := hclsyntax.ParseConfig([]byte(test.config), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
			body := file.Body
			got, valDiags := Decode(body, test.spec, test.ctx)

			var diags hcl.Diagnostics
			diags = append(diags, parseDiags...)
			diags = append(diags, valDiags...)

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

func TestSourceRange(t *testing.T) {
	tests := []struct {
		config string
		spec   Spec
		want   hcl.Range
	}{
		{
			"a = 1\n",
			&AttrSpec{
				Name: "a",
			},
			hcl.Range{
				Start: hcl.Pos{Line: 1, Column: 5, Byte: 4},
				End:   hcl.Pos{Line: 1, Column: 6, Byte: 5},
			},
		},
		{
			`
b {
  a = 1
}
`,
			&BlockSpec{
				TypeName: "b",
				Nested: &AttrSpec{
					Name: "a",
				},
			},
			hcl.Range{
				Start: hcl.Pos{Line: 3, Column: 7, Byte: 11},
				End:   hcl.Pos{Line: 3, Column: 8, Byte: 12},
			},
		},
		{
			`
b {
  c {
    a = 1
  }
}
`,
			&BlockSpec{
				TypeName: "b",
				Nested: &BlockSpec{
					TypeName: "c",
					Nested: &AttrSpec{
						Name: "a",
					},
				},
			},
			hcl.Range{
				Start: hcl.Pos{Line: 4, Column: 9, Byte: 19},
				End:   hcl.Pos{Line: 4, Column: 10, Byte: 20},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d-%s", i, test.config), func(t *testing.T) {
			file, diags := hclsyntax.ParseConfig([]byte(test.config), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})
			if len(diags) != 0 {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), 0)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}
			body := file.Body

			got := SourceRange(body, test.spec)

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.want)
			}
		})
	}

}
