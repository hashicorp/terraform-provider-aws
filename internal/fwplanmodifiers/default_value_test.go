package fwplanmodifiers

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDefaultValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plannedValue  attr.Value
		currentValue  attr.Value
		defaultValue  attr.Value
		expectedValue attr.Value
		expectError   bool
	}
	tests := map[string]testCase{
		"non-default non-Null string": {
			plannedValue:  types.String{Value: "gamma"},
			currentValue:  types.String{Value: "beta"},
			defaultValue:  types.String{Value: "alpha"},
			expectedValue: types.String{Value: "gamma"},
		},
		"non-default non-Null string, current Null": {
			plannedValue:  types.String{Value: "gamma"},
			currentValue:  types.String{Null: true},
			defaultValue:  types.String{Value: "alpha"},
			expectedValue: types.String{Value: "gamma"},
		},
		"non-default Null string, current Null": {
			plannedValue:  types.String{Null: true},
			currentValue:  types.String{Value: "beta"},
			defaultValue:  types.String{Value: "alpha"},
			expectedValue: types.String{Value: "alpha"},
		},
		"default string": {
			plannedValue:  types.String{Null: true},
			currentValue:  types.String{Value: "alpha"},
			defaultValue:  types.String{Value: "alpha"},
			expectedValue: types.String{Value: "alpha"},
		},
		"default string on create": {
			plannedValue:  types.String{Null: true},
			currentValue:  types.String{Null: true},
			defaultValue:  types.String{Value: "alpha"},
			expectedValue: types.String{Value: "alpha"},
		},
		"non-default non-Null number": {
			plannedValue:  types.Number{Value: big.NewFloat(30)},
			currentValue:  types.Number{Value: big.NewFloat(10)},
			defaultValue:  types.Number{Value: big.NewFloat(-10)},
			expectedValue: types.Number{Value: big.NewFloat(30)},
		},
		"non-default non-Null number, current Null": {
			plannedValue:  types.Number{Value: big.NewFloat(30)},
			currentValue:  types.Number{Null: true},
			defaultValue:  types.Number{Value: big.NewFloat(-10)},
			expectedValue: types.Number{Value: big.NewFloat(30)},
		},
		"non-default Null number, current Null": {
			plannedValue:  types.Number{Null: true},
			currentValue:  types.Number{Value: big.NewFloat(10)},
			defaultValue:  types.Number{Value: big.NewFloat(-10)},
			expectedValue: types.Number{Value: big.NewFloat(-10)},
		},
		"default number": {
			plannedValue:  types.Number{Null: true},
			currentValue:  types.Number{Value: big.NewFloat(-10)},
			defaultValue:  types.Number{Value: big.NewFloat(-10)},
			expectedValue: types.Number{Value: big.NewFloat(-10)},
		},
		"non-default string list": {
			plannedValue: types.List{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "POST"},
			}},
			currentValue: types.List{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "PUT"},
			}},
			defaultValue: types.List{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
			expectedValue: types.List{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "POST"},
			}},
		},
		"non-default string list, current out of order": {
			plannedValue: types.List{ElemType: types.StringType, Null: true},
			currentValue: types.List{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "HEAD"},
				types.String{Value: "GET"},
			}},
			defaultValue: types.List{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
			expectedValue: types.List{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
		},
		"default string list": {
			plannedValue: types.List{ElemType: types.StringType, Null: true},
			currentValue: types.List{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
			defaultValue: types.List{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
			expectedValue: types.List{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
		},
		"non-default string set": {
			plannedValue: types.Set{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "POST"},
			}},
			currentValue: types.Set{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "PUT"},
			}},
			defaultValue: types.Set{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
			expectedValue: types.Set{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "POST"},
			}},
		},
		"default string set, current out of order": {
			plannedValue: types.Set{ElemType: types.StringType, Null: true},
			currentValue: types.Set{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "HEAD"},
				types.String{Value: "GET"},
			}},
			defaultValue: types.Set{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
			expectedValue: types.Set{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "HEAD"},
				types.String{Value: "GET"},
			}},
		},
		"default string set": {
			plannedValue: types.Set{ElemType: types.StringType, Null: true},
			currentValue: types.Set{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
			defaultValue: types.Set{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
			expectedValue: types.Set{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
		},
		"non-default object": {
			plannedValue: types.Object{
				AttrTypes: map[string]attr.Type{
					"value": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"value": types.String{Value: "gamma"},
				},
			},
			currentValue: types.Object{
				AttrTypes: map[string]attr.Type{
					"value": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"value": types.String{Value: "beta"},
				},
			},
			defaultValue: types.Object{
				AttrTypes: map[string]attr.Type{
					"value": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"value": types.String{Value: "alpha"},
				},
			},
			expectedValue: types.Object{
				AttrTypes: map[string]attr.Type{
					"value": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"value": types.String{Value: "gamma"},
				},
			},
		},
		"non-default object, different value": {
			plannedValue: types.Object{
				AttrTypes: map[string]attr.Type{
					"value": types.StringType,
				},
				Null: true,
			},
			currentValue: types.Object{
				AttrTypes: map[string]attr.Type{
					"value": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"value": types.String{Value: "beta"},
				},
			},
			defaultValue: types.Object{
				AttrTypes: map[string]attr.Type{
					"value": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"value": types.String{Value: "alpha"},
				},
			},
			expectedValue: types.Object{
				AttrTypes: map[string]attr.Type{
					"value": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"value": types.String{Value: "alpha"},
				},
			},
		},
		"default object": {
			plannedValue: types.Object{
				AttrTypes: map[string]attr.Type{
					"value": types.StringType,
				},
				Null: true,
			},
			currentValue: types.Object{
				AttrTypes: map[string]attr.Type{
					"value": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"value": types.String{Value: "alpha"},
				},
			},
			defaultValue: types.Object{
				AttrTypes: map[string]attr.Type{
					"value": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"value": types.String{Value: "alpha"},
				},
			},
			expectedValue: types.Object{
				AttrTypes: map[string]attr.Type{
					"value": types.StringType,
				},
				Attrs: map[string]attr.Value{
					"value": types.String{Value: "alpha"},
				},
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			request := tfsdk.ModifyAttributePlanRequest{
				AttributePath:  path.Root("test"),
				AttributePlan:  test.plannedValue,
				AttributeState: test.currentValue,
			}
			response := tfsdk.ModifyAttributePlanResponse{
				AttributePlan: request.AttributePlan,
			}
			DefaultValue(test.defaultValue).Modify(ctx, request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}

			if diff := cmp.Diff(response.AttributePlan, test.expectedValue); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
