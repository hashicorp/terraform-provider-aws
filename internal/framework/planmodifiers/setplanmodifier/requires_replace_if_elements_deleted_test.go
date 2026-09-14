// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package setplanmodifier_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfsetplanmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers/setplanmodifier"
)

func TestRequiresReplaceIfElementsDeleted(t *testing.T) {
	t.Parallel()

	type testCase struct {
		request  planmodifier.SetRequest
		expected planmodifier.SetResponse
	}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"value": schema.SetAttribute{
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
	testPlan := func(value types.Set) tfsdk.Plan {
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
	testState := func(value types.Set) tfsdk.State {
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
	emptyValue := types.SetValueMust(types.StringType, []attr.Value{})
	oldValue := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("old-value1"), types.StringValue("old-value2")})
	newValueNoChange := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("old-value2"), types.StringValue("old-value1")})
	newValueAdd := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("old-value1"), types.StringValue("old-value2"), types.StringValue("new-value")})
	newValueAddAndDel := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("old-value1"), types.StringValue("new-value")})
	newValueDel := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("old-value2")})

	tests := map[string]testCase{
		"unknown value on update": {
			request: planmodifier.SetRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(types.SetUnknown(types.StringType)),
				PlanValue:  types.SetUnknown(types.StringType),
			},
			expected: planmodifier.SetResponse{
				PlanValue: types.SetUnknown(types.StringType),
			},
		},
		"empty value on create": {
			request: planmodifier.SetRequest{
				State:      nullState,
				StateValue: types.SetNull(types.StringType),
				Plan:       testPlan(emptyValue),
				PlanValue:  emptyValue,
			},
			expected: planmodifier.SetResponse{
				PlanValue: emptyValue,
			},
		},
		"non-empty value to empty value on update": {
			request: planmodifier.SetRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(emptyValue),
				PlanValue:  emptyValue,
			},
			expected: planmodifier.SetResponse{
				PlanValue:       emptyValue,
				RequiresReplace: true,
			},
		},
		"non-empty value to null value on update": {
			request: planmodifier.SetRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(types.SetNull(types.StringType)),
				PlanValue:  types.SetNull(types.StringType),
			},
			expected: planmodifier.SetResponse{
				PlanValue:       types.SetNull(types.StringType),
				RequiresReplace: true,
			},
		},
		"non-empty value to same non-empty value on update": {
			request: planmodifier.SetRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(oldValue),
				PlanValue:  oldValue,
			},
			expected: planmodifier.SetResponse{
				PlanValue: oldValue,
			},
		},
		"non-empty value to same reordered non-empty value on update": {
			request: planmodifier.SetRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(newValueNoChange),
				PlanValue:  newValueNoChange,
			},
			expected: planmodifier.SetResponse{
				PlanValue: newValueNoChange,
			},
		},
		"empty value to non-empty value on update": {
			request: planmodifier.SetRequest{
				State:      testState(emptyValue),
				StateValue: emptyValue,
				Plan:       testPlan(newValueAdd),
				PlanValue:  newValueAdd,
			},
			expected: planmodifier.SetResponse{
				PlanValue: newValueAdd,
			},
		},
		"non-empty value to non-empty value with addition on update": {
			request: planmodifier.SetRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(newValueAdd),
				PlanValue:  newValueAdd,
			},
			expected: planmodifier.SetResponse{
				PlanValue: newValueAdd,
			},
		},
		"non-empty value to non-empty value with addition and deletion on update": {
			request: planmodifier.SetRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(newValueAddAndDel),
				PlanValue:  newValueAddAndDel,
			},
			expected: planmodifier.SetResponse{
				PlanValue:       newValueAddAndDel,
				RequiresReplace: true,
			},
		},
		"non-empty value to non-empty value with deletion on update": {
			request: planmodifier.SetRequest{
				State:      testState(oldValue),
				StateValue: oldValue,
				Plan:       testPlan(newValueDel),
				PlanValue:  newValueDel,
			},
			expected: planmodifier.SetResponse{
				PlanValue:       newValueDel,
				RequiresReplace: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := planmodifier.SetResponse{
				PlanValue: test.request.PlanValue,
			}
			tfsetplanmodifier.RequiresReplaceIfElementsDeleted.PlanModifySet(t.Context(), test.request, &response)

			if diff := cmp.Diff(test.expected, response); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
