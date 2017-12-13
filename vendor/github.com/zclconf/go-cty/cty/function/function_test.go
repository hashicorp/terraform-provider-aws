package function

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestReturnTypeForValues(t *testing.T) {
	tests := []struct {
		Spec     *Spec
		Args     []cty.Value
		WantType cty.Type
		WantErr  bool
	}{
		{
			Spec: &Spec{
				Params: []Parameter{},
				Type:   StaticReturnType(cty.Number),
				Impl:   stubImpl,
			},
			Args:     []cty.Value{},
			WantType: cty.Number,
		},
		{
			Spec: &Spec{
				Params: []Parameter{},
				Type:   StaticReturnType(cty.Number),
				Impl:   stubImpl,
			},
			Args:    []cty.Value{cty.NumberIntVal(2)},
			WantErr: true,
		},
		{
			Spec: &Spec{
				Params: []Parameter{},
				Type:   StaticReturnType(cty.Number),
				Impl:   stubImpl,
			},
			Args:    []cty.Value{cty.UnknownVal(cty.Number)},
			WantErr: true,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type: cty.Number,
					},
				},
				Type: StaticReturnType(cty.Number),
				Impl: stubImpl,
			},
			Args:     []cty.Value{cty.NumberIntVal(2)},
			WantType: cty.Number,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type: cty.Number,
					},
				},
				Type: StaticReturnType(cty.Number),
				Impl: stubImpl,
			},
			Args:     []cty.Value{cty.UnknownVal(cty.Number)},
			WantType: cty.Number,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type: cty.Number,
					},
				},
				Type: StaticReturnType(cty.Number),
				Impl: stubImpl,
			},
			Args:     []cty.Value{cty.DynamicVal},
			WantType: cty.DynamicPseudoType,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type:             cty.Number,
						AllowDynamicType: true,
					},
				},
				Type: StaticReturnType(cty.Number),
				Impl: stubImpl,
			},
			Args:     []cty.Value{cty.DynamicVal},
			WantType: cty.Number,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type:             cty.Number,
						AllowDynamicType: true,
					},
				},
				Type: StaticReturnType(cty.Number),
				Impl: stubImpl,
			},
			Args:    []cty.Value{cty.UnknownVal(cty.String)},
			WantErr: true,
		},
		{
			Spec: &Spec{
				Params: []Parameter{
					{
						Type:             cty.Number,
						AllowDynamicType: true,
					},
				},
				Type: StaticReturnType(cty.Number),
				Impl: stubImpl,
			},
			Args:    []cty.Value{cty.StringVal("hello")},
			WantErr: true,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			f := New(test.Spec)
			gotType, gotErr := f.ReturnTypeForValues(test.Args)

			if test.WantErr {
				if gotErr == nil {
					t.Errorf("succeeded with %#v; want error", gotType)
				}
			} else {
				if gotErr != nil {
					t.Fatalf("unexpected error\nspec: %#v\nargs: %#v\nerr:  %s\nwant: %#v", test.Spec, test.Args, gotErr, test.WantType)
				}

				if gotType == cty.NilType {
					t.Fatalf("returned type is invalid")
				}

				if !gotType.Equals(test.WantType) {
					t.Errorf("wrong return type\nspec: %#v\nargs: %#v\ngot:  %#v\nwant: %#v", test.Spec, test.Args, gotType, test.WantType)
				}
			}
		})
	}
}

func stubImpl([]cty.Value, cty.Type) (cty.Value, error) {
	return cty.NilVal, fmt.Errorf("should not be called")
}
