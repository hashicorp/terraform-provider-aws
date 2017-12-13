package convert

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestUnify(t *testing.T) {
	tests := []struct {
		Input           []cty.Type
		WantType        cty.Type
		WantConversions []bool
	}{
		{
			[]cty.Type{},
			cty.NilType,
			nil,
		},
		{
			[]cty.Type{cty.String},
			cty.String,
			[]bool{false},
		},
		{
			[]cty.Type{cty.Number},
			cty.Number,
			[]bool{false},
		},
		{
			[]cty.Type{cty.Number, cty.Number},
			cty.Number,
			[]bool{false, false},
		},
		{
			[]cty.Type{cty.Number, cty.String},
			cty.String,
			[]bool{true, false},
		},
		{
			[]cty.Type{cty.String, cty.Number},
			cty.String,
			[]bool{false, true},
		},
		{
			[]cty.Type{cty.Bool, cty.String, cty.Number},
			cty.String,
			[]bool{true, false, true},
		},
		{
			[]cty.Type{cty.Bool, cty.Number},
			cty.NilType,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Input), func(t *testing.T) {
			gotType, gotConvs := Unify(test.Input)
			if gotType != test.WantType {
				t.Errorf("wrong result type\ngot:  %#v\nwant: %#v", gotType, test.WantType)
			}

			gotConvsNil := gotConvs == nil
			wantConvsNil := test.WantConversions == nil
			if gotConvsNil && wantConvsNil {
				// Success!
				return
			}

			if gotConvsNil != wantConvsNil {
				if gotConvsNil {
					t.Fatalf("got nil conversions; want %#v", test.WantConversions)
				} else {
					t.Fatalf("got conversions; want nil")
				}
			}

			gotConvsBool := make([]bool, len(gotConvs))
			for i, f := range gotConvs {
				gotConvsBool[i] = f != nil
			}

			if !reflect.DeepEqual(gotConvsBool, test.WantConversions) {
				t.Fatalf(
					"wrong conversions\ngot:  %#v\nwant: %#v",
					gotConvsBool, test.WantConversions,
				)
			}
		})
	}
}
