// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package setplanmodifier_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfsetplanmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers/setplanmodifier"
)

func TestDefaultValueFromPath(t *testing.T) {
	t.Parallel()

	type testCase struct {
		request  planmodifier.SetRequest
		expected planmodifier.SetResponse
	}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"src": schema.SetAttribute{
				ElementType: types.StringType,
			},
			"dst": schema.SetAttribute{
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
	testPlan := func(src, dst types.Set) tfsdk.Plan {
		tfSrc, err := src.ToTerraformValue(t.Context())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		tfDst, err := dst.ToTerraformValue(t.Context())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		return tfsdk.Plan{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(t.Context()),
				map[string]tftypes.Value{
					"src": tfSrc,
					"dst": tfDst,
				},
			),
		}
	}
	testState := func(src, dst types.Set) tfsdk.State {
		tfSrc, err := src.ToTerraformValue(t.Context())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		tfDst, err := dst.ToTerraformValue(t.Context())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		return tfsdk.State{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(t.Context()),
				map[string]tftypes.Value{
					"src": tfSrc,
					"dst": tfDst,
				},
			),
		}
	}
	defaultValue := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("default-value")})
	computedValue := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("computed-value")})
	configuredValue := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("configured-value")})

	tests := map[string]testCase{
		"unknown value on create": {
			request: planmodifier.SetRequest{
				State:      nullState,
				StateValue: types.SetNull(types.StringType),
				Plan:       testPlan(defaultValue, types.SetUnknown(types.StringType)),
				PlanValue:  types.SetUnknown(types.StringType),
			},
			expected: planmodifier.SetResponse{
				PlanValue: defaultValue,
			},
		},
		"unknown value on update": {
			request: planmodifier.SetRequest{
				State:      testState(defaultValue, computedValue),
				StateValue: computedValue,
				Plan:       testPlan(defaultValue, types.SetUnknown(types.StringType)),
				PlanValue:  types.SetUnknown(types.StringType),
			},
			expected: planmodifier.SetResponse{
				PlanValue: defaultValue,
			},
		},
		"null known value on create": {
			request: planmodifier.SetRequest{
				State:      nullState,
				StateValue: types.SetNull(types.StringType),
				Plan:       testPlan(types.SetNull(types.StringType), types.SetUnknown(types.StringType)),
				PlanValue:  types.SetUnknown(types.StringType),
			},
			expected: planmodifier.SetResponse{
				PlanValue: types.SetNull(types.StringType),
			},
		},
		"null known value on update": {
			request: planmodifier.SetRequest{
				State:      testState(defaultValue, computedValue),
				StateValue: computedValue,
				Plan:       testPlan(types.SetNull(types.StringType), types.SetUnknown(types.StringType)),
				PlanValue:  types.SetUnknown(types.StringType),
			},
			expected: planmodifier.SetResponse{
				PlanValue: types.SetNull(types.StringType),
			},
		},
		"non-null known value on create": {
			request: planmodifier.SetRequest{
				State:      nullState,
				StateValue: types.SetNull(types.StringType),
				Plan:       testPlan(defaultValue, configuredValue),
				PlanValue:  configuredValue,
			},
			expected: planmodifier.SetResponse{
				PlanValue: configuredValue,
			},
		},
		"non-null known value on update": {
			request: planmodifier.SetRequest{
				State:      testState(defaultValue, computedValue),
				StateValue: computedValue,
				Plan:       testPlan(defaultValue, configuredValue),
				PlanValue:  configuredValue,
			},
			expected: planmodifier.SetResponse{
				PlanValue: configuredValue,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := planmodifier.SetResponse{
				PlanValue: test.request.PlanValue,
			}
			tfsetplanmodifier.DefaultValueFromPath[types.Set](path.Root("src")).PlanModifySet(t.Context(), test.request, &response)

			if diff := cmp.Diff(test.expected, response); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
