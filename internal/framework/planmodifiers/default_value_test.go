package planmodifiers

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
			plannedValue:  types.StringValue("gamma"),
			currentValue:  types.StringValue("beta"),
			defaultValue:  types.StringValue("alpha"),
			expectedValue: types.StringValue("gamma"),
		},
		"non-default non-Null string, current Null": {
			plannedValue:  types.StringValue("gamma"),
			currentValue:  types.StringNull(),
			defaultValue:  types.StringValue("alpha"),
			expectedValue: types.StringValue("gamma"),
		},
		"non-default Null string, current Null": {
			plannedValue:  types.StringNull(),
			currentValue:  types.StringValue("beta"),
			defaultValue:  types.StringValue("alpha"),
			expectedValue: types.StringValue("alpha"),
		},
		"default string": {
			plannedValue:  types.StringNull(),
			currentValue:  types.StringValue("alpha"),
			defaultValue:  types.StringValue("alpha"),
			expectedValue: types.StringValue("alpha"),
		},
		"default string on create": {
			plannedValue:  types.StringNull(),
			currentValue:  types.StringNull(),
			defaultValue:  types.StringValue("alpha"),
			expectedValue: types.StringValue("alpha"),
		},
		"non-default non-Null number": {
			plannedValue:  types.NumberValue(big.NewFloat(30)),
			currentValue:  types.NumberValue(big.NewFloat(10)),
			defaultValue:  types.NumberValue(big.NewFloat(-10)),
			expectedValue: types.NumberValue(big.NewFloat(30)),
		},
		"non-default non-Null number, current Null": {
			plannedValue:  types.NumberValue(big.NewFloat(30)),
			currentValue:  types.NumberNull(),
			defaultValue:  types.NumberValue(big.NewFloat(-10)),
			expectedValue: types.NumberValue(big.NewFloat(30)),
		},
		"non-default Null number, current Null": {
			plannedValue:  types.NumberNull(),
			currentValue:  types.NumberValue(big.NewFloat(10)),
			defaultValue:  types.NumberValue(big.NewFloat(-10)),
			expectedValue: types.NumberValue(big.NewFloat(-10)),
		},
		"default number": {
			plannedValue:  types.NumberNull(),
			currentValue:  types.NumberValue(big.NewFloat(-10)),
			defaultValue:  types.NumberValue(big.NewFloat(-10)),
			expectedValue: types.NumberValue(big.NewFloat(-10)),
		},
		"non-default string list": {
			plannedValue: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("POST"),
			}),
			currentValue: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("PUT"),
			}),
			defaultValue: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
			expectedValue: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("POST"),
			}),
		},
		"non-default string list, current out of order": {
			plannedValue: types.ListNull(types.StringType),
			currentValue: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("HEAD"),
				types.StringValue("GET"),
			}),
			defaultValue: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
			expectedValue: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
		},
		"default string list": {
			plannedValue: types.ListNull(types.StringType),
			currentValue: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
			defaultValue: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
			expectedValue: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
		},
		"non-default string set": {
			plannedValue: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("POST"),
			}),
			currentValue: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("PUT"),
			}),
			defaultValue: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
			expectedValue: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("POST"),
			}),
		},
		"default string set, current out of order": {
			plannedValue: types.SetNull(types.StringType),
			currentValue: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("HEAD"),
				types.StringValue("GET"),
			}),
			defaultValue: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
			expectedValue: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("HEAD"),
				types.StringValue("GET"),
			}),
		},
		"default string set": {
			plannedValue: types.SetNull(types.StringType),
			currentValue: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
			defaultValue: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
			expectedValue: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
		},
		"non-default object": {
			plannedValue: types.ObjectValueMust(map[string]attr.Type{
				"value": types.StringType,
			},
				map[string]attr.Value{
					"value": types.StringValue("gamma"),
				},
			),
			currentValue: types.ObjectValueMust(map[string]attr.Type{
				"value": types.StringType,
			},
				map[string]attr.Value{
					"value": types.StringValue("beta"),
				},
			),
			defaultValue: types.ObjectValueMust(map[string]attr.Type{
				"value": types.StringType,
			},
				map[string]attr.Value{
					"value": types.StringValue("alpha"),
				},
			),
			expectedValue: types.ObjectValueMust(map[string]attr.Type{
				"value": types.StringType,
			},
				map[string]attr.Value{
					"value": types.StringValue("gamma"),
				},
			),
		},
		"non-default object, different value": {
			plannedValue: types.ObjectNull(map[string]attr.Type{
				"value": types.StringType,
			}),
			currentValue: types.ObjectValueMust(map[string]attr.Type{
				"value": types.StringType,
			},
				map[string]attr.Value{
					"value": types.StringValue("beta"),
				},
			),
			defaultValue: types.ObjectValueMust(map[string]attr.Type{
				"value": types.StringType,
			},
				map[string]attr.Value{
					"value": types.StringValue("alpha"),
				},
			),
			expectedValue: types.ObjectValueMust(map[string]attr.Type{
				"value": types.StringType,
			},
				map[string]attr.Value{
					"value": types.StringValue("alpha"),
				},
			),
		},
		"default object": {
			plannedValue: types.ObjectNull(map[string]attr.Type{
				"value": types.StringType,
			}),
			currentValue: types.ObjectValueMust(map[string]attr.Type{
				"value": types.StringType,
			},
				map[string]attr.Value{
					"value": types.StringValue("alpha"),
				},
			),
			defaultValue: types.ObjectValueMust(map[string]attr.Type{
				"value": types.StringType,
			},
				map[string]attr.Value{
					"value": types.StringValue("alpha"),
				},
			),
			expectedValue: types.ObjectValueMust(map[string]attr.Type{
				"value": types.StringType,
			},
				map[string]attr.Value{
					"value": types.StringValue("alpha"),
				},
			),
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
