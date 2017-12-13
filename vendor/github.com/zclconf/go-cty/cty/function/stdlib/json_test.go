package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestJSONEncode(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		// This does not comprehensively test all possible inputs because
		// the underlying functions in package json already have tests of
		// their own. Here we are mainly concerned with seeing that the
		// function's definition accepts all reasonable values.
		{
			cty.NumberIntVal(15),
			cty.StringVal(`15`),
		},
		{
			cty.StringVal("hello"),
			cty.StringVal(`"hello"`),
		},
		{
			cty.True,
			cty.StringVal(`true`),
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.StringVal(`[]`),
		},
		{
			cty.ListVal([]cty.Value{cty.True, cty.False}),
			cty.StringVal(`[true,false]`),
		},
		{
			cty.ObjectVal(map[string]cty.Value{"true": cty.True, "false": cty.False}),
			cty.StringVal(`{"false":false,"true":true}`),
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.String),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.String),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("JSONEncode(%#v)", test.Input), func(t *testing.T) {
			got, err := JSONEncode(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestJSONDecode(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal(`15`),
			cty.NumberIntVal(15),
		},
		{
			cty.StringVal(`"hello"`),
			cty.StringVal("hello"),
		},
		{
			cty.StringVal(`true`),
			cty.True,
		},
		{
			cty.StringVal(`[]`),
			cty.EmptyTupleVal,
		},
		{
			cty.StringVal(`[true,false]`),
			cty.TupleVal([]cty.Value{cty.True, cty.False}),
		},
		{
			cty.StringVal(`{"false":false,"true":true}`),
			cty.ObjectVal(map[string]cty.Value{"true": cty.True, "false": cty.False}),
		},
		{
			cty.UnknownVal(cty.String),
			cty.DynamicVal, // need to know the value to determine the type
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("JSONDecode(%#v)", test.Input), func(t *testing.T) {
			got, err := JSONDecode(test.Input)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
