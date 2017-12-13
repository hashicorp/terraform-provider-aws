package convert

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestSortTypes(t *testing.T) {
	tests := []struct {
		Input []cty.Type
		Want  []cty.Type
	}{
		{
			[]cty.Type{},
			[]cty.Type{},
		},
		{
			[]cty.Type{cty.String, cty.Number},
			[]cty.Type{cty.String, cty.Number},
		},
		{
			[]cty.Type{cty.Number, cty.String},
			[]cty.Type{cty.String, cty.Number},
		},
		{
			[]cty.Type{cty.String, cty.Bool},
			[]cty.Type{cty.String, cty.Bool},
		},
		{
			[]cty.Type{cty.Bool, cty.String},
			[]cty.Type{cty.String, cty.Bool},
		},
		{
			[]cty.Type{cty.Bool, cty.String, cty.Number},
			[]cty.Type{cty.String, cty.Bool, cty.Number},
		},
		{
			[]cty.Type{cty.Number, cty.String, cty.Bool},
			[]cty.Type{cty.String, cty.Number, cty.Bool},
		},
		{
			[]cty.Type{cty.String, cty.String},
			[]cty.Type{cty.String, cty.String},
		},
		{
			[]cty.Type{cty.Number, cty.String, cty.Number},
			[]cty.Type{cty.String, cty.Number, cty.Number},
		},
		{
			[]cty.Type{cty.String, cty.List(cty.String)},
			[]cty.Type{cty.String, cty.List(cty.String)},
		},
		{
			[]cty.Type{cty.List(cty.String), cty.String},
			[]cty.Type{cty.List(cty.String), cty.String},
		},
		{
			// This result is somewhat arbitrary, but the important thing
			// is that it is consistent.
			[]cty.Type{cty.Bool, cty.List(cty.String), cty.String},
			[]cty.Type{cty.List(cty.String), cty.String, cty.Bool},
		},
		{
			[]cty.Type{cty.String, cty.DynamicPseudoType},
			[]cty.Type{cty.String, cty.DynamicPseudoType},
		},
		{
			[]cty.Type{cty.DynamicPseudoType, cty.String},
			[]cty.Type{cty.String, cty.DynamicPseudoType},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Input), func(t *testing.T) {
			idxs := sortTypes(test.Input)

			if len(idxs) != len(test.Input) {
				t.Fatalf("wrong number of indexes %q; want %q", len(idxs), len(test.Input))
			}

			got := make([]cty.Type, len(idxs))

			for i, idx := range idxs {
				got[i] = test.Input[idx]
			}

			for i := range test.Want {
				if !got[i].Equals(test.Want[i]) {
					t.Errorf(
						"wrong order\ninput: %#v\ngot:   %#v\nwant:  %#v",
						test.Input,
						got, test.Want,
					)
					break
				}
			}
		})
	}
}
