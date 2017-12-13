package json

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestImpliedType(t *testing.T) {
	tests := []struct {
		Input string
		Want  cty.Type
	}{
		{
			"null",
			cty.DynamicPseudoType,
		},
		{
			"1",
			cty.Number,
		},
		{
			"1.2222222222222222222222222222222222",
			cty.Number,
		},
		{
			"999999999999999999999999999999999999999999999999999999999999",
			cty.Number,
		},
		{
			`""`,
			cty.String,
		},
		{
			`"hello"`,
			cty.String,
		},
		{
			"true",
			cty.Bool,
		},
		{
			"false",
			cty.Bool,
		},
		{
			"{}",
			cty.EmptyObject,
		},
		{
			`{"true": true}`,
			cty.Object(map[string]cty.Type{
				"true": cty.Bool,
			}),
		},
		{
			`{"true": true, "name": "Ermintrude", "null": null}`,
			cty.Object(map[string]cty.Type{
				"true": cty.Bool,
				"name": cty.String,
				"null": cty.DynamicPseudoType,
			}),
		},
		{
			"[]",
			cty.EmptyTuple,
		},
		{
			"[true, 1.2, null]",
			cty.Tuple([]cty.Type{cty.Bool, cty.Number, cty.DynamicPseudoType}),
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			got, err := ImpliedType([]byte(test.Input))

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.Equals(test.Want) {
				t.Errorf(
					"wrong type\ninput: %s\ngot:   %#v\nwant:  %#v",
					test.Input, got, test.Want,
				)
			}
		})
	}
}
