// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package listplanmodifier_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tflistplanmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers/listplanmodifier"
)

func TestRequiresReplaceIfEmptied(t *testing.T) {
	t.Parallel()

	type testCase struct {
		request  planmodifier.ListRequest
		expected planmodifier.ListResponse
	}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"value": schema.ListAttribute{
				ElementType: types.StringType,
			},
		},
	}
	nullState := tfsdk.State{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			testSchema.Type().TerraformType(t.Context()),
			nil,
		),
	}
	testPlan := func(value types.List) tfsdk.Plan {
		tfValue, err := value.ToTerraformValue(t.Context())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		return tfsdk.Plan{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(t.Context()),
				map[string]tftypes.Value{
					"value": tfValue,
				},
			),
		}
	}
	testState := func(value types.List) tfsdk.State {
		tfValue, err := value.ToTerraformValue(t.Context())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		return tfsdk.State{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(t.Context()),
				map[string]tftypes.Value{
					"value": tfValue,
				},
			),
		}
	}
	emptyValue := types.ListValueMust(types.StringType, []attr.Value{})
	oldValue := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("old-value1"), types.StringValue("old-value2")})
	newValue := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("new-value")})

	tests := map[string]testCase{
		"unknown value on update": {
			request: planmodifier.ListRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(types.ListUnknown(types.StringType)),
				PlanValue:  types.ListUnknown(types.StringType),
			},
			expected: planmodifier.ListResponse{
				PlanValue: types.ListUnknown(types.StringType),
			},
		},
		"empty value on create": {
			request: planmodifier.ListRequest{
				State:      nullState,
				StateValue: types.ListNull(types.StringType),
				Plan:       testPlan(emptyValue),
				PlanValue:  emptyValue,
			},
			expected: planmodifier.ListResponse{
				PlanValue: emptyValue,
			},
		},
		"non-empty value to empty value on update": {
			request: planmodifier.ListRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(emptyValue),
				PlanValue:  emptyValue,
			},
			expected: planmodifier.ListResponse{
				PlanValue:       emptyValue,
				RequiresReplace: true,
			},
		},
		"non-empty value to null value on update": {
			request: planmodifier.ListRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(types.ListNull(types.StringType)),
				PlanValue:  types.ListNull(types.StringType),
			},
			expected: planmodifier.ListResponse{
				PlanValue:       types.ListNull(types.StringType),
				RequiresReplace: true,
			},
		},
		"non-empty value to different non-empty value on update": {
			request: planmodifier.ListRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(newValue),
				PlanValue:  newValue,
			},
			expected: planmodifier.ListResponse{
				PlanValue: newValue,
			},
		},
		"non-empty value to same non-empty value on update": {
			request: planmodifier.ListRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(oldValue),
				PlanValue:  oldValue,
			},
			expected: planmodifier.ListResponse{
				PlanValue: oldValue,
			},
		},
		"unchanged non-empty value on update": {
			request: planmodifier.ListRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(oldValue),
				PlanValue:  oldValue,
			},
			expected: planmodifier.ListResponse{
				PlanValue: oldValue,
			},
		},
		"empty value to non-empty value on update": {
			request: planmodifier.ListRequest{
				State:      testState(emptyValue),
				StateValue: emptyValue,
				Plan:       testPlan(newValue),
				PlanValue:  newValue,
			},
			expected: planmodifier.ListResponse{
				PlanValue: newValue,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := planmodifier.ListResponse{
				PlanValue: test.request.PlanValue,
			}
			tflistplanmodifier.RequiresReplaceIfEmptied.PlanModifyList(t.Context(), test.request, &response)

			if diff := cmp.Diff(test.expected, response); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
