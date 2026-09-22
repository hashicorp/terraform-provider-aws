// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package listplanmodifier_test

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
	tflistplanmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers/listplanmodifier"
)

func TestUnknownWhenOtherValueChanges(t *testing.T) {
	t.Parallel()

	type testCase struct {
		request  planmodifier.ListRequest
		expected planmodifier.ListResponse
	}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"src": schema.StringAttribute{},
			"dst": schema.ListAttribute{
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
	testPlan := func(src types.String, dst types.List) tfsdk.Plan {
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
	testState := func(src types.String, dst types.List) tfsdk.State {
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
	unknownValue := types.ListUnknown(types.StringType)
	configuredValue := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("configured-value")})

	tests := map[string]testCase{
		"on create": {
			request: planmodifier.ListRequest{
				State:      nullState,
				StateValue: types.ListNull(types.StringType),
				Plan:       testPlan(types.StringValue("test"), unknownValue),
				PlanValue:  unknownValue,
			},
			expected: planmodifier.ListResponse{
				PlanValue: unknownValue,
			},
		},
		"equal values on update": {
			request: planmodifier.ListRequest{
				State:      testState(types.StringValue("test"), configuredValue),
				StateValue: configuredValue,
				Plan:       testPlan(types.StringValue("test"), unknownValue),
				PlanValue:  unknownValue,
			},
			expected: planmodifier.ListResponse{
				PlanValue: configuredValue,
			},
		},
		"different values on update": {
			request: planmodifier.ListRequest{
				State:      testState(types.StringValue("old"), configuredValue),
				StateValue: configuredValue,
				Plan:       testPlan(types.StringValue("new"), unknownValue),
				PlanValue:  unknownValue,
			},
			expected: planmodifier.ListResponse{
				PlanValue: unknownValue,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := planmodifier.ListResponse{
				PlanValue: test.request.PlanValue,
			}
			tflistplanmodifier.UnknownWhenOtherValueChanges(path.Root("src")).PlanModifyList(t.Context(), test.request, &response)

			if diff := cmp.Diff(test.expected, response); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
