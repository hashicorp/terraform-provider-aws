package gocty

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestImpliedType(t *testing.T) {
	tests := []struct {
		Input interface{}
		Want  cty.Type
	}{
		// Primitive types
		{
			int(0),
			cty.Number,
		},
		{
			int8(0),
			cty.Number,
		},
		{
			int16(0),
			cty.Number,
		},
		{
			int32(0),
			cty.Number,
		},
		{
			int64(0),
			cty.Number,
		},
		{
			uint(0),
			cty.Number,
		},
		{
			uint8(0),
			cty.Number,
		},
		{
			uint16(0),
			cty.Number,
		},
		{
			uint32(0),
			cty.Number,
		},
		{
			uint64(0),
			cty.Number,
		},
		{
			float32(0),
			cty.Number,
		},
		{
			float64(0),
			cty.Number,
		},
		{
			false,
			cty.Bool,
		},
		{
			"",
			cty.String,
		},

		// Collection types
		{
			[]int(nil),
			cty.List(cty.Number),
		},
		{
			[][]int(nil),
			cty.List(cty.List(cty.Number)),
		},
		{
			map[string]int(nil),
			cty.Map(cty.Number),
		},
		{
			map[string]map[string]int(nil),
			cty.Map(cty.Map(cty.Number)),
		},
		{
			map[string][]int(nil),
			cty.Map(cty.List(cty.Number)),
		},

		// Structs
		{
			testStruct{},
			cty.Object(map[string]cty.Type{
				"name":   cty.String,
				"number": cty.Number,
			}),
		},

		// Pointers (unwrapped and ignored)
		{
			ptrToInt(0),
			cty.Number,
		},
		{
			ptrToBool(false),
			cty.Bool,
		},
		{
			ptrToString(""),
			cty.String,
		},
		{
			&testStruct{},
			cty.Object(map[string]cty.Type{
				"name":   cty.String,
				"number": cty.Number,
			}),
		},

		// Dynamic
		{
			cty.NilVal,
			cty.DynamicPseudoType,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Input), func(t *testing.T) {
			got, err := ImpliedType(test.Input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.Equals(test.Want) {
				t.Fatalf(
					"wrong result\ninput: %#v\ngot:   %#v\nwant:  %#v",
					test.Input, got, test.Want,
				)
			}
		})
	}
}
