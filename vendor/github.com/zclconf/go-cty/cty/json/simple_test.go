package json

import (
	"encoding/json"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestSimpleJSONValue(t *testing.T) {
	tests := []struct {
		Input cty.Value
		JSON  string
		Want  cty.Value
	}{
		{
			cty.NumberIntVal(5),
			`5`,
			cty.NumberIntVal(5),
		},
		{
			cty.True,
			`true`,
			cty.True,
		},
		{
			cty.StringVal("hello"),
			`"hello"`,
			cty.StringVal("hello"),
		},
		{
			cty.TupleVal([]cty.Value{cty.StringVal("hello"), cty.True}),
			`["hello",true]`,
			cty.TupleVal([]cty.Value{cty.StringVal("hello"), cty.True}),
		},
		{
			cty.ListVal([]cty.Value{cty.False, cty.True}),
			`[false,true]`,
			cty.TupleVal([]cty.Value{cty.False, cty.True}),
		},
		{
			cty.SetVal([]cty.Value{cty.False, cty.True}),
			`[false,true]`,
			cty.TupleVal([]cty.Value{cty.False, cty.True}),
		},
		{
			cty.ObjectVal(map[string]cty.Value{"true": cty.True, "greet": cty.StringVal("hello")}),
			`{"greet":"hello","true":true}`,
			cty.ObjectVal(map[string]cty.Value{"true": cty.True, "greet": cty.StringVal("hello")}),
		},
		{
			cty.MapVal(map[string]cty.Value{"true": cty.True, "false": cty.False}),
			`{"false":false,"true":true}`,
			cty.ObjectVal(map[string]cty.Value{"true": cty.True, "false": cty.False}),
		},
		{
			cty.NullVal(cty.Bool),
			`null`,
			cty.NullVal(cty.DynamicPseudoType), // type is lost in the round-trip
		},
	}

	for _, test := range tests {
		t.Run(test.Input.GoString(), func(t *testing.T) {
			wrappedInput := SimpleJSONValue{test.Input}
			buf, err := json.Marshal(wrappedInput)
			if err != nil {
				t.Fatalf("unexpected error from json.Marshal: %s", err)
			}
			if string(buf) != test.JSON {
				t.Fatalf(
					"incorrect JSON\ninput: %#v\ngot:   %s\nwant:  %s",
					test.Input, buf, test.JSON,
				)
			}

			var wrappedOutput SimpleJSONValue
			err = json.Unmarshal(buf, &wrappedOutput)
			if err != nil {
				t.Fatalf("unexpected error from json.Unmarshal: %s", err)
			}

			if !wrappedOutput.Value.RawEquals(test.Want) {
				t.Fatalf(
					"incorrect result\nJSON:  %s\ngot:   %#v\nwant:  %#v",
					buf, wrappedOutput.Value, test.Want,
				)
			}
		})
	}
}
