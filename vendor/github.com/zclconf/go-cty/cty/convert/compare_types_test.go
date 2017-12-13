package convert

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestCompareTypes(t *testing.T) {
	tests := []struct {
		A    cty.Type
		B    cty.Type
		Want int
	}{
		// Primitives
		{
			cty.String,
			cty.String,
			0,
		},
		{
			cty.String,
			cty.Number,
			-1,
		},
		{
			cty.Number,
			cty.String,
			1,
		},
		{
			cty.String,
			cty.Bool,
			-1,
		},
		{
			cty.Bool,
			cty.String,
			1,
		},
		{
			cty.Bool,
			cty.Number,
			0,
		},
		{
			cty.Number,
			cty.Bool,
			0,
		},

		// Lists
		{
			cty.List(cty.String),
			cty.List(cty.String),
			0,
		},
		{
			cty.List(cty.String),
			cty.List(cty.Number),
			-1,
		},
		{
			cty.List(cty.Number),
			cty.List(cty.String),
			1,
		},
		{
			cty.List(cty.String),
			cty.String,
			0,
		},

		// Sets
		{
			cty.Set(cty.String),
			cty.Set(cty.String),
			0,
		},
		{
			cty.Set(cty.String),
			cty.Set(cty.Number),
			-1,
		},
		{
			cty.Set(cty.Number),
			cty.Set(cty.String),
			1,
		},
		{
			cty.Set(cty.String),
			cty.String,
			0,
		},

		// Maps
		{
			cty.Map(cty.String),
			cty.Map(cty.String),
			0,
		},
		{
			cty.Map(cty.String),
			cty.Map(cty.Number),
			-1,
		},
		{
			cty.Map(cty.Number),
			cty.Map(cty.String),
			1,
		},
		{
			cty.Map(cty.String),
			cty.String,
			0,
		},

		// Objects
		{
			cty.EmptyObject,
			cty.EmptyObject,
			0,
		},
		{
			cty.EmptyObject,
			cty.Object(map[string]cty.Type{
				"name": cty.String,
			}),
			0,
		},
		{
			cty.Object(map[string]cty.Type{
				"name": cty.String,
			}),
			cty.Object(map[string]cty.Type{
				"name": cty.String,
			}),
			0,
		},
		{
			cty.Object(map[string]cty.Type{
				"name":   cty.String,
				"number": cty.Number,
			}),
			cty.Object(map[string]cty.Type{
				"name": cty.String,
			}),
			0,
		},
		{
			cty.Object(map[string]cty.Type{
				"number": cty.Number,
			}),
			cty.Object(map[string]cty.Type{
				"name": cty.String,
			}),
			0,
		},
		{
			cty.Object(map[string]cty.Type{
				"name":   cty.String,
				"number": cty.Number,
			}),
			cty.Object(map[string]cty.Type{
				"name":   cty.String,
				"number": cty.Number,
			}),
			0,
		},
		{
			cty.Object(map[string]cty.Type{
				"name":   cty.String,
				"number": cty.String,
			}),
			cty.Object(map[string]cty.Type{
				"name":   cty.String,
				"number": cty.Number,
			}),
			-1,
		},
		{
			cty.Object(map[string]cty.Type{
				"name":   cty.String,
				"number": cty.Number,
			}),
			cty.Object(map[string]cty.Type{
				"name":   cty.String,
				"number": cty.String,
			}),
			1,
		},
		{
			// This is the tricky case where comparing types doesn't tell
			// the whole story, because there is a third type C where both
			// attributes are strings which would be a common base type
			// of these.
			cty.Object(map[string]cty.Type{
				"a": cty.String,
				"b": cty.Number,
			}),
			cty.Object(map[string]cty.Type{
				"a": cty.Number,
				"b": cty.String,
			}),
			0,
		},

		// Tuples
		{
			cty.EmptyTuple,
			cty.EmptyTuple,
			0,
		},
		{
			cty.EmptyTuple,
			cty.Tuple([]cty.Type{cty.String}),
			0,
		},
		{
			cty.Tuple([]cty.Type{cty.String}),
			cty.Tuple([]cty.Type{cty.String}),
			0,
		},
		{
			cty.Tuple([]cty.Type{cty.String, cty.Number}),
			cty.Tuple([]cty.Type{cty.String}),
			0,
		},
		{
			cty.Tuple([]cty.Type{cty.String, cty.Number}),
			cty.Tuple([]cty.Type{cty.String, cty.Number}),
			0,
		},
		{
			cty.Tuple([]cty.Type{cty.String, cty.String}),
			cty.Tuple([]cty.Type{cty.String, cty.Number}),
			-1,
		},
		{
			cty.Tuple([]cty.Type{cty.String, cty.Number}),
			cty.Tuple([]cty.Type{cty.String, cty.String}),
			1,
		},
		{
			// This is the tricky case where comparing types doesn't tell
			// the whole story, because there is a third type C where both
			// elements are strings which would be a common base type
			// of these.
			cty.Tuple([]cty.Type{cty.String, cty.Number}),
			cty.Tuple([]cty.Type{cty.Number, cty.String}),
			0,
		},

		// Lists and Sets
		{
			cty.Set(cty.String),
			cty.List(cty.String),
			1,
		},
		{
			cty.List(cty.String),
			cty.Set(cty.String),
			-1,
		},
		{
			cty.List(cty.String),
			cty.Set(cty.Number),
			-1,
		},
		{
			cty.Set(cty.Number),
			cty.List(cty.String),
			1,
		},
		{
			cty.List(cty.Number),
			cty.Set(cty.String),
			-1,
		},
		{
			cty.Set(cty.String),
			cty.List(cty.Number),
			1,
		},

		// Dynamics
		{
			cty.DynamicPseudoType,
			cty.DynamicPseudoType,
			0,
		},
		{
			cty.DynamicPseudoType,
			cty.String,
			1,
		},
		{
			cty.String,
			cty.DynamicPseudoType,
			-1,
		},
		{
			cty.Number,
			cty.DynamicPseudoType,
			-1,
		},
		{
			cty.DynamicPseudoType,
			cty.Number,
			1,
		},
		{
			cty.Bool,
			cty.DynamicPseudoType,
			-1,
		},
		{
			cty.DynamicPseudoType,
			cty.Bool,
			1,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v,%#v", test.A, test.B), func(t *testing.T) {
			got := compareTypes(test.A, test.B)
			if got != test.Want {
				t.Errorf(
					"wrong result\nA: %#v\nB: %#v\ngot:  %#v\nwant: %#v",
					test.A, test.B,
					got, test.Want,
				)
			}
		})
	}
}
