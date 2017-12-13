package json

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestValueJSONable(t *testing.T) {
	bytesType := cty.Capsule("bytes", reflect.TypeOf([]byte(nil)))
	buf := []byte("hello")
	bytesVal := cty.CapsuleVal(bytesType, &buf)

	tests := []struct {
		Value  cty.Value
		Type   cty.Type
		Want   string
		DecVal cty.Value
	}{
		// Primitives
		{
			cty.StringVal("hello"),
			cty.String,
			`"hello"`,
			cty.StringVal("hello"),
		},
		{
			cty.StringVal(""),
			cty.String,
			`""`,
			cty.StringVal(""),
		},
		{
			cty.StringVal("15"),
			cty.Number,
			`15`,
			cty.NumberIntVal(15),
		},
		{
			cty.StringVal("true"),
			cty.Bool,
			`true`,
			cty.True,
		},
		{
			cty.StringVal("1"),
			cty.Bool,
			`true`,
			cty.True,
		},
		{
			cty.NullVal(cty.String),
			cty.String,
			`null`,
			cty.NullVal(cty.String),
		},
		{
			cty.NumberIntVal(2),
			cty.Number,
			`2`,
			cty.NumberIntVal(2),
		},
		{
			cty.NumberFloatVal(2.5),
			cty.Number,
			`2.5`,
			cty.NumberFloatVal(2.5),
		},
		{
			cty.NumberIntVal(5),
			cty.String,
			`"5"`,
			cty.StringVal("5"),
		},
		{
			cty.True,
			cty.Bool,
			`true`,
			cty.True,
		},
		{
			cty.False,
			cty.Bool,
			`false`,
			cty.False,
		},
		{
			cty.True,
			cty.String,
			`"true"`,
			cty.StringVal("true"),
		},

		// Lists
		{
			cty.ListVal([]cty.Value{cty.True, cty.False}),
			cty.List(cty.Bool),
			`[true,false]`,
			cty.ListVal([]cty.Value{cty.True, cty.False}),
		},
		{
			cty.ListValEmpty(cty.Bool),
			cty.List(cty.Bool),
			`[]`,
			cty.ListValEmpty(cty.Bool),
		},
		{
			cty.ListVal([]cty.Value{cty.True, cty.False}),
			cty.List(cty.String),
			`["true","false"]`,
			cty.ListVal([]cty.Value{cty.StringVal("true"), cty.StringVal("false")}),
		},

		// Sets
		{
			cty.SetVal([]cty.Value{cty.True, cty.False}),
			cty.Set(cty.Bool),
			`[false,true]`,
			cty.SetVal([]cty.Value{cty.True, cty.False}),
		},
		{
			cty.SetValEmpty(cty.Bool),
			cty.Set(cty.Bool),
			`[]`,
			cty.SetValEmpty(cty.Bool),
		},

		// Tuples
		{
			cty.TupleVal([]cty.Value{cty.True, cty.NumberIntVal(5)}),
			cty.Tuple([]cty.Type{cty.Bool, cty.Number}),
			`[true,5]`,
			cty.TupleVal([]cty.Value{cty.True, cty.NumberIntVal(5)}),
		},
		{
			cty.EmptyTupleVal,
			cty.EmptyTuple,
			`[]`,
			cty.EmptyTupleVal,
		},

		// Maps
		{
			cty.MapValEmpty(cty.Bool),
			cty.Map(cty.Bool),
			`{}`,
			cty.MapValEmpty(cty.Bool),
		},
		{
			cty.MapVal(map[string]cty.Value{"yes": cty.True, "no": cty.False}),
			cty.Map(cty.Bool),
			`{"no":false,"yes":true}`,
			cty.MapVal(map[string]cty.Value{"yes": cty.True, "no": cty.False}),
		},
		{
			cty.NullVal(cty.Map(cty.Bool)),
			cty.Map(cty.Bool),
			`null`,
			cty.NullVal(cty.Map(cty.Bool)),
		},

		// Objects
		{
			cty.EmptyObjectVal,
			cty.EmptyObject,
			`{}`,
			cty.EmptyObjectVal,
		},
		{
			cty.ObjectVal(map[string]cty.Value{"bool": cty.True, "number": cty.Zero}),
			cty.Object(map[string]cty.Type{"bool": cty.Bool, "number": cty.Number}),
			`{"bool":true,"number":0}`,
			cty.ObjectVal(map[string]cty.Value{"bool": cty.True, "number": cty.Zero}),
		},

		// Capsules
		{
			bytesVal,
			bytesType,
			`"aGVsbG8="`,
			bytesVal,
		},

		// Encoding into dynamic produces type information wrapper
		{
			cty.True,
			cty.DynamicPseudoType,
			`{"value":true,"type":"bool"}`,
			cty.True,
		},
		{
			cty.StringVal("hello"),
			cty.DynamicPseudoType,
			`{"value":"hello","type":"string"}`,
			cty.StringVal("hello"),
		},
		{
			cty.NumberIntVal(5),
			cty.DynamicPseudoType,
			`{"value":5,"type":"number"}`,
			cty.NumberIntVal(5),
		},
		{
			cty.ListVal([]cty.Value{cty.True, cty.False}),
			cty.DynamicPseudoType,
			`{"value":[true,false],"type":["list","bool"]}`,
			cty.ListVal([]cty.Value{cty.True, cty.False}),
		},
		{
			cty.ListVal([]cty.Value{cty.True, cty.False}),
			cty.List(cty.DynamicPseudoType),
			`[{"value":true,"type":"bool"},{"value":false,"type":"bool"}]`,
			cty.ListVal([]cty.Value{cty.True, cty.False}),
		},
		{
			cty.ObjectVal(map[string]cty.Value{"static": cty.True, "dynamic": cty.True}),
			cty.Object(map[string]cty.Type{"static": cty.Bool, "dynamic": cty.DynamicPseudoType}),
			`{"dynamic":{"value":true,"type":"bool"},"static":true}`,
			cty.ObjectVal(map[string]cty.Value{"static": cty.True, "dynamic": cty.True}),
		},
		{
			cty.ObjectVal(map[string]cty.Value{"static": cty.True, "dynamic": cty.True}),
			cty.DynamicPseudoType,
			`{"value":{"dynamic":true,"static":true},"type":["object",{"dynamic":"bool","static":"bool"}]}`,
			cty.ObjectVal(map[string]cty.Value{"static": cty.True, "dynamic": cty.True}),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v to %#v", test.Value, test.Type), func(t *testing.T) {
			gotBuf, err := Marshal(test.Value, test.Type)

			if err != nil {
				t.Fatalf("unexpected error from Marshal: %s", err)
			}

			got := string(gotBuf)

			if got != test.Want {
				t.Errorf(
					"wrong serialization\nvalue: %#v\ntype:  %#v\ngot:   %s\nwant:  %s",
					test.Value, test.Type, got, test.Want,
				)
			}

			newVal, err := Unmarshal(gotBuf, test.Type)
			if err != nil {
				t.Fatalf("unexpected error from Unmarshal: %s", err)
			}

			// If we're dealing with our capsule type then we need to do some
			// more manual comparison because capsule values compare by
			// pointer identity but pointers don't survive marshalling.
			if newVal.Type().Equals(bytesType) {
				gotBuf := newVal.EncapsulatedValue()
				wantBuf := test.DecVal.EncapsulatedValue()
				if !reflect.DeepEqual(gotBuf, wantBuf) {
					t.Errorf(
						"mismatch after Unmarshal\njson: %s\ntype: %#v\ngot:  %#v\nwant: %#v",
						got, test.Type, newVal, test.Value,
					)
				}
			} else if !newVal.RawEquals(test.DecVal) {
				t.Errorf(
					"mismatch after Unmarshal\njson: %s\ntype: %#v\ngot:  %#v\nwant: %#v",
					got, test.Type, newVal, test.Value,
				)
			}
		})
	}
}
