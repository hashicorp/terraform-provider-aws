package convert

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		Value     cty.Value
		Type      cty.Type
		Want      cty.Value
		WantError bool
	}{
		{
			Value: cty.StringVal("hello"),
			Type:  cty.String,
			Want:  cty.StringVal("hello"),
		},
		{
			Value: cty.StringVal("1"),
			Type:  cty.Number,
			Want:  cty.NumberIntVal(1),
		},
		{
			Value: cty.StringVal("1.5"),
			Type:  cty.Number,
			Want:  cty.NumberFloatVal(1.5),
		},
		{
			Value:     cty.StringVal("hello"),
			Type:      cty.Number,
			WantError: true,
		},
		{
			Value: cty.StringVal("true"),
			Type:  cty.Bool,
			Want:  cty.True,
		},
		{
			Value: cty.StringVal("1"),
			Type:  cty.Bool,
			Want:  cty.True,
		},
		{
			Value: cty.StringVal("false"),
			Type:  cty.Bool,
			Want:  cty.False,
		},
		{
			Value: cty.StringVal("0"),
			Type:  cty.Bool,
			Want:  cty.False,
		},
		{
			Value:     cty.StringVal("hello"),
			Type:      cty.Bool,
			WantError: true,
		},
		{
			Value: cty.NumberIntVal(4),
			Type:  cty.String,
			Want:  cty.StringVal("4"),
		},
		{
			Value: cty.NumberFloatVal(3.14159265359),
			Type:  cty.String,
			Want:  cty.StringVal("3.14159265359"),
		},
		{
			Value: cty.True,
			Type:  cty.String,
			Want:  cty.StringVal("true"),
		},
		{
			Value: cty.False,
			Type:  cty.String,
			Want:  cty.StringVal("false"),
		},
		{
			Value: cty.UnknownVal(cty.String),
			Type:  cty.Number,
			Want:  cty.UnknownVal(cty.Number),
		},
		{
			Value: cty.UnknownVal(cty.Number),
			Type:  cty.String,
			Want:  cty.UnknownVal(cty.String),
		},
		{
			Value: cty.DynamicVal,
			Type:  cty.String,
			Want:  cty.UnknownVal(cty.String),
		},
		{
			Value: cty.StringVal("hello"),
			Type:  cty.DynamicPseudoType,
			Want:  cty.StringVal("hello"),
		},
		{
			Value: cty.ListVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NumberIntVal(10),
			}),
			Type: cty.List(cty.String),
			Want: cty.ListVal([]cty.Value{
				cty.StringVal("5"),
				cty.StringVal("10"),
			}),
		},
		{
			Value: cty.SetVal([]cty.Value{
				cty.StringVal("5"),
				cty.StringVal("10"),
			}),
			Type: cty.List(cty.String),
			Want: cty.ListVal([]cty.Value{
				// NOTE: This results depends on the traversal order of the
				// set, which may change if the set implementation changes.
				cty.StringVal("10"),
				cty.StringVal("5"),
			}),
		},
		{
			Value: cty.SetVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NumberIntVal(10),
			}),
			Type: cty.List(cty.String),
			Want: cty.ListVal([]cty.Value{
				// NOTE: This results depends on the traversal order of the
				// set, which may change if the set implementation changes.
				cty.StringVal("5"),
				cty.StringVal("10"),
			}),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.StringVal("hello"),
			}),
			Type: cty.List(cty.String),
			Want: cty.ListVal([]cty.Value{
				cty.StringVal("5"),
				cty.StringVal("hello"),
			}),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.StringVal("12"),
			}),
			Type: cty.List(cty.Number),
			Want: cty.ListVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NumberIntVal(12),
			}),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NumberIntVal(10),
			}),
			Type: cty.List(cty.DynamicPseudoType),
			Want: cty.ListVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NumberIntVal(10),
			}),
		},
		{
			Value: cty.TupleVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.StringVal("hello"),
			}),
			Type: cty.List(cty.DynamicPseudoType),
			Want: cty.ListVal([]cty.Value{
				cty.StringVal("5"),
				cty.StringVal("hello"),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"num": cty.NumberIntVal(5),
				"str": cty.StringVal("hello"),
			}),
			Type: cty.Map(cty.String),
			Want: cty.MapVal(map[string]cty.Value{
				"num": cty.StringVal("5"),
				"str": cty.StringVal("hello"),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"num": cty.NumberIntVal(5),
				"str": cty.StringVal("12"),
			}),
			Type: cty.Map(cty.Number),
			Want: cty.MapVal(map[string]cty.Value{
				"num": cty.NumberIntVal(5),
				"str": cty.NumberIntVal(12),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"num1": cty.NumberIntVal(5),
				"num2": cty.NumberIntVal(10),
			}),
			Type: cty.Map(cty.DynamicPseudoType),
			Want: cty.MapVal(map[string]cty.Value{
				"num1": cty.NumberIntVal(5),
				"num2": cty.NumberIntVal(10),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"num": cty.NumberIntVal(5),
				"str": cty.StringVal("hello"),
			}),
			Type: cty.Map(cty.DynamicPseudoType),
			Want: cty.MapVal(map[string]cty.Value{
				"num": cty.StringVal("5"),
				"str": cty.StringVal("hello"),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"list":  cty.ListValEmpty(cty.Bool),
				"tuple": cty.EmptyTupleVal,
			}),
			Type: cty.Map(cty.DynamicPseudoType),
			Want: cty.MapVal(map[string]cty.Value{
				"list":  cty.ListValEmpty(cty.Bool),
				"tuple": cty.ListValEmpty(cty.Bool),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"map": cty.MapValEmpty(cty.String),
				"obj": cty.EmptyObjectVal,
			}),
			Type: cty.Map(cty.DynamicPseudoType),
			Want: cty.MapVal(map[string]cty.Value{
				"map": cty.MapValEmpty(cty.String),
				"obj": cty.MapValEmpty(cty.String),
			}),
		},
		{
			Value: cty.ObjectVal(map[string]cty.Value{
				"num":  cty.NumberIntVal(5),
				"bool": cty.True,
			}),
			Type:      cty.Map(cty.DynamicPseudoType),
			WantError: true, // no common base type to unify to
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v to %#v", test.Value, test.Type), func(t *testing.T) {
			got, err := Convert(test.Value, test.Type)

			switch {
			case test.WantError:
				if err == nil {
					t.Errorf("conversion succeeded with %#v; want error", got)
				}
			default:
				if err != nil {
					t.Fatalf("conversion failed: %s", err)
				}

				if !got.RawEquals(test.Want) {
					t.Errorf(
						"wrong result\nvalue: %#v\ntype:  %#v\ngot:   %#v\nwant:  %#v",
						test.Value, test.Type,
						got, test.Want,
					)
				}
			}
		})
	}
}
