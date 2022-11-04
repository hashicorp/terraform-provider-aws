package fwplanmodifiers

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNormalizeEmptyMap(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plannedValue  attr.Value
		currentValue  attr.Value
		expectedValue attr.Value
		expectError   bool
	}
	tests := map[string]testCase{
		"planned non-empty Map, current non-empty Map": {
			plannedValue: types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{
				"2": types.String{Value: "TWO"},
			}},
			currentValue: types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{
				"1": types.String{Value: "ONE"},
			}},
			expectedValue: types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{
				"2": types.String{Value: "TWO"},
			}},
		},
		"planned non-empty Map, current Null Map": {
			plannedValue: types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{
				"2": types.String{Value: "TWO"},
			}},
			currentValue: types.Map{ElemType: types.StringType, Null: true},
			expectedValue: types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{
				"2": types.String{Value: "TWO"},
			}},
		},
		"planned Null Map, current non-empty Map": {
			plannedValue: types.Map{ElemType: types.StringType, Null: true},
			currentValue: types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{
				"1": types.String{Value: "ONE"},
			}},
			expectedValue: types.Map{ElemType: types.StringType, Null: true},
		},
		"planned empty Map, current Null Map": {
			plannedValue:  types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{}},
			currentValue:  types.Map{ElemType: types.StringType, Null: true},
			expectedValue: types.Map{ElemType: types.StringType, Null: true},
		},
		"planned Null Map, current empty Map": {
			plannedValue:  types.Map{ElemType: types.StringType, Null: true},
			currentValue:  types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{}},
			expectedValue: types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{}},
		},
		"planned empty Map, current empty Map": {
			plannedValue:  types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{}},
			currentValue:  types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{}},
			expectedValue: types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{}},
		},
		"planned Null Map, current Null Map": {
			plannedValue:  types.Map{ElemType: types.StringType, Null: true},
			currentValue:  types.Map{ElemType: types.StringType, Null: true},
			expectedValue: types.Map{ElemType: types.StringType, Null: true},
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
			NormalizeEmptyMap().Modify(ctx, request, &response)

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
