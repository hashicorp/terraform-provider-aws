package gohcl

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/hcl2/hcl"
	zclJSON "github.com/hashicorp/hcl2/hcl/json"
	"github.com/zclconf/go-cty/cty"
)

func TestDecodeBody(t *testing.T) {
	deepEquals := func(other interface{}) func(v interface{}) bool {
		return func(v interface{}) bool {
			return reflect.DeepEqual(v, other)
		}
	}

	type withNameExpression struct {
		Name hcl.Expression `hcl:"name"`
	}

	tests := []struct {
		Body      map[string]interface{}
		Target    interface{}
		Check     func(v interface{}) bool
		DiagCount int
	}{
		{
			map[string]interface{}{},
			struct{}{},
			deepEquals(struct{}{}),
			0,
		},
		{
			map[string]interface{}{},
			struct {
				Name string `hcl:"name"`
			}{},
			deepEquals(struct {
				Name string `hcl:"name"`
			}{}),
			1, // name is required
		},
		{
			map[string]interface{}{},
			struct {
				Name *string `hcl:"name"`
			}{},
			deepEquals(struct {
				Name *string `hcl:"name"`
			}{}),
			0,
		},
		{
			map[string]interface{}{},
			withNameExpression{},
			func(v interface{}) bool {
				if v == nil {
					return false
				}

				wne, valid := v.(withNameExpression)
				if !valid {
					return false
				}

				if wne.Name == nil {
					return false
				}

				nameVal, _ := wne.Name.Value(nil)
				if !nameVal.IsNull() {
					return false
				}

				return true
			},
			0,
		},
		{
			map[string]interface{}{
				"name": "Ermintrude",
			},
			withNameExpression{},
			func(v interface{}) bool {
				if v == nil {
					return false
				}

				wne, valid := v.(withNameExpression)
				if !valid {
					return false
				}

				if wne.Name == nil {
					return false
				}

				nameVal, _ := wne.Name.Value(nil)
				if !nameVal.Equals(cty.StringVal("Ermintrude")).True() {
					return false
				}

				return true
			},
			0,
		},
		{
			map[string]interface{}{
				"name": "Ermintrude",
			},
			struct {
				Name string `hcl:"name"`
			}{},
			deepEquals(struct {
				Name string `hcl:"name"`
			}{"Ermintrude"}),
			0,
		},
		{
			map[string]interface{}{
				"name": "Ermintrude",
				"age":  23,
			},
			struct {
				Name string `hcl:"name"`
			}{},
			deepEquals(struct {
				Name string `hcl:"name"`
			}{"Ermintrude"}),
			1, // Extraneous "age" property
		},
		{
			map[string]interface{}{
				"name": "Ermintrude",
				"age":  50,
			},
			struct {
				Name  string         `hcl:"name"`
				Attrs hcl.Attributes `hcl:",remain"`
			}{},
			func(gotI interface{}) bool {
				got := gotI.(struct {
					Name  string         `hcl:"name"`
					Attrs hcl.Attributes `hcl:",remain"`
				})
				return got.Name == "Ermintrude" && len(got.Attrs) == 1 && got.Attrs["age"] != nil
			},
			0,
		},
		{
			map[string]interface{}{
				"name": "Ermintrude",
				"age":  50,
			},
			struct {
				Name   string   `hcl:"name"`
				Remain hcl.Body `hcl:",remain"`
			}{},
			func(gotI interface{}) bool {
				got := gotI.(struct {
					Name   string   `hcl:"name"`
					Remain hcl.Body `hcl:",remain"`
				})

				attrs, _ := got.Remain.JustAttributes()

				return got.Name == "Ermintrude" && len(attrs) == 1 && attrs["age"] != nil
			},
			0,
		},
		{
			map[string]interface{}{
				"name": "Ermintrude",
				"age":  51,
			},
			struct {
				Name   string               `hcl:"name"`
				Remain map[string]cty.Value `hcl:",remain"`
			}{},
			deepEquals(struct {
				Name   string               `hcl:"name"`
				Remain map[string]cty.Value `hcl:",remain"`
			}{
				Name: "Ermintrude",
				Remain: map[string]cty.Value{
					"age": cty.NumberIntVal(51),
				},
			}),
			0,
		},
		{
			map[string]interface{}{
				"noodle": map[string]interface{}{},
			},
			struct {
				Noodle struct{} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				// Generating no diagnostics is good enough for this one.
				return true
			},
			0,
		},
		{
			map[string]interface{}{
				"noodle": []map[string]interface{}{{}},
			},
			struct {
				Noodle struct{} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				// Generating no diagnostics is good enough for this one.
				return true
			},
			0,
		},
		{
			map[string]interface{}{
				"noodle": []map[string]interface{}{{}, {}},
			},
			struct {
				Noodle struct{} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				// Generating one diagnostic is good enough for this one.
				return true
			},
			1,
		},
		{
			map[string]interface{}{},
			struct {
				Noodle struct{} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				// Generating one diagnostic is good enough for this one.
				return true
			},
			1,
		},
		{
			map[string]interface{}{
				"noodle": []map[string]interface{}{},
			},
			struct {
				Noodle struct{} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				// Generating one diagnostic is good enough for this one.
				return true
			},
			1,
		},
		{
			map[string]interface{}{
				"noodle": map[string]interface{}{},
			},
			struct {
				Noodle *struct{} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				return gotI.(struct {
					Noodle *struct{} `hcl:"noodle,block"`
				}).Noodle != nil
			},
			0,
		},
		{
			map[string]interface{}{
				"noodle": []map[string]interface{}{{}},
			},
			struct {
				Noodle *struct{} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				return gotI.(struct {
					Noodle *struct{} `hcl:"noodle,block"`
				}).Noodle != nil
			},
			0,
		},
		{
			map[string]interface{}{
				"noodle": []map[string]interface{}{},
			},
			struct {
				Noodle *struct{} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				return gotI.(struct {
					Noodle *struct{} `hcl:"noodle,block"`
				}).Noodle == nil
			},
			0,
		},
		{
			map[string]interface{}{
				"noodle": []map[string]interface{}{{}, {}},
			},
			struct {
				Noodle *struct{} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				// Generating one diagnostic is good enough for this one.
				return true
			},
			1,
		},
		{
			map[string]interface{}{
				"noodle": []map[string]interface{}{},
			},
			struct {
				Noodle []struct{} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				noodle := gotI.(struct {
					Noodle []struct{} `hcl:"noodle,block"`
				}).Noodle
				return len(noodle) == 0
			},
			0,
		},
		{
			map[string]interface{}{
				"noodle": []map[string]interface{}{{}},
			},
			struct {
				Noodle []struct{} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				noodle := gotI.(struct {
					Noodle []struct{} `hcl:"noodle,block"`
				}).Noodle
				return len(noodle) == 1
			},
			0,
		},
		{
			map[string]interface{}{
				"noodle": []map[string]interface{}{{}, {}},
			},
			struct {
				Noodle []struct{} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				noodle := gotI.(struct {
					Noodle []struct{} `hcl:"noodle,block"`
				}).Noodle
				return len(noodle) == 2
			},
			0,
		},
		{
			map[string]interface{}{
				"noodle": map[string]interface{}{},
			},
			struct {
				Noodle struct {
					Name string `hcl:"name,label"`
				} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				// Generating two diagnostics is good enough for this one.
				// (one for the missing noodle block and the other for
				// the JSON serialization detecting the missing level of
				// heirarchy for the label.)
				return true
			},
			2,
		},
		{
			map[string]interface{}{
				"noodle": map[string]interface{}{
					"foo_foo": map[string]interface{}{},
				},
			},
			struct {
				Noodle struct {
					Name string `hcl:"name,label"`
				} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				noodle := gotI.(struct {
					Noodle struct {
						Name string `hcl:"name,label"`
					} `hcl:"noodle,block"`
				}).Noodle
				return noodle.Name == "foo_foo"
			},
			0,
		},
		{
			map[string]interface{}{
				"noodle": map[string]interface{}{
					"foo_foo": map[string]interface{}{},
					"bar_baz": map[string]interface{}{},
				},
			},
			struct {
				Noodle struct {
					Name string `hcl:"name,label"`
				} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				// One diagnostic is enough for this one.
				return true
			},
			1,
		},
		{
			map[string]interface{}{
				"noodle": map[string]interface{}{
					"foo_foo": map[string]interface{}{},
					"bar_baz": map[string]interface{}{},
				},
			},
			struct {
				Noodles []struct {
					Name string `hcl:"name,label"`
				} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				noodles := gotI.(struct {
					Noodles []struct {
						Name string `hcl:"name,label"`
					} `hcl:"noodle,block"`
				}).Noodles
				return len(noodles) == 2 && (noodles[0].Name == "foo_foo" || noodles[0].Name == "bar_baz") && (noodles[1].Name == "foo_foo" || noodles[1].Name == "bar_baz") && noodles[0].Name != noodles[1].Name
			},
			0,
		},
		{
			map[string]interface{}{
				"noodle": map[string]interface{}{
					"foo_foo": map[string]interface{}{
						"type": "rice",
					},
				},
			},
			struct {
				Noodle struct {
					Name string `hcl:"name,label"`
					Type string `hcl:"type"`
				} `hcl:"noodle,block"`
			}{},
			func(gotI interface{}) bool {
				noodle := gotI.(struct {
					Noodle struct {
						Name string `hcl:"name,label"`
						Type string `hcl:"type"`
					} `hcl:"noodle,block"`
				}).Noodle
				return noodle.Name == "foo_foo" && noodle.Type == "rice"
			},
			0,
		},

		{
			map[string]interface{}{
				"name": "Ermintrude",
				"age":  34,
			},
			map[string]string(nil),
			deepEquals(map[string]string{
				"name": "Ermintrude",
				"age":  "34",
			}),
			0,
		},
		{
			map[string]interface{}{
				"name": "Ermintrude",
				"age":  89,
			},
			map[string]*hcl.Attribute(nil),
			func(gotI interface{}) bool {
				got := gotI.(map[string]*hcl.Attribute)
				return len(got) == 2 && got["name"] != nil && got["age"] != nil
			},
			0,
		},
		{
			map[string]interface{}{
				"name": "Ermintrude",
				"age":  13,
			},
			map[string]hcl.Expression(nil),
			func(gotI interface{}) bool {
				got := gotI.(map[string]hcl.Expression)
				return len(got) == 2 && got["name"] != nil && got["age"] != nil
			},
			0,
		},
		{
			map[string]interface{}{
				"name": "Ermintrude",
				"age":  13,
			},
			map[string]cty.Value(nil),
			deepEquals(map[string]cty.Value{
				"name": cty.StringVal("Ermintrude"),
				"age":  cty.NumberIntVal(13),
			}),
			0,
		},
	}

	for i, test := range tests {
		// For convenience here we're going to use the JSON parser
		// to process the given body.
		buf, err := json.Marshal(test.Body)
		if err != nil {
			t.Fatalf("error JSON-encoding body for test %d: %s", i, err)
		}

		t.Run(string(buf), func(t *testing.T) {
			file, diags := zclJSON.Parse(buf, "test.json")
			if len(diags) != 0 {
				t.Fatalf("diagnostics while parsing: %s", diags.Error())
			}

			targetVal := reflect.New(reflect.TypeOf(test.Target))

			diags = DecodeBody(file.Body, nil, targetVal.Interface())
			if len(diags) != test.DiagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.DiagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}
			got := targetVal.Elem().Interface()
			if !test.Check(got) {
				t.Errorf("wrong result\ngot:  %s", spew.Sdump(got))
			}
		})
	}

}

func TestDecodeExpression(t *testing.T) {
	tests := []struct {
		Value     cty.Value
		Target    interface{}
		Want      interface{}
		DiagCount int
	}{
		{
			cty.StringVal("hello"),
			"",
			"hello",
			0,
		},
		{
			cty.StringVal("hello"),
			cty.NilVal,
			cty.StringVal("hello"),
			0,
		},
		{
			cty.NumberIntVal(2),
			"",
			"2",
			0,
		},
		{
			cty.StringVal("true"),
			false,
			true,
			0,
		},
		{
			cty.NullVal(cty.String),
			"",
			"",
			1, // null value is not allowed
		},
		{
			cty.UnknownVal(cty.String),
			"",
			"",
			1, // value must be known
		},
		{
			cty.ListVal([]cty.Value{cty.True}),
			false,
			false,
			1, // bool required
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			expr := &fixedExpression{test.Value}

			targetVal := reflect.New(reflect.TypeOf(test.Target))

			diags := DecodeExpression(expr, nil, targetVal.Interface())
			if len(diags) != test.DiagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.DiagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}
			got := targetVal.Elem().Interface()
			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

type fixedExpression struct {
	val cty.Value
}

func (e *fixedExpression) Value(ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	return e.val, nil
}

func (e *fixedExpression) Range() (r hcl.Range) {
	return
}
func (e *fixedExpression) StartRange() (r hcl.Range) {
	return
}

func (e *fixedExpression) Variables() []hcl.Traversal {
	return nil
}
