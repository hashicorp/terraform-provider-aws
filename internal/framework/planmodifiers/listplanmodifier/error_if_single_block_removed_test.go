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

func TestErrorIfSingleBlockRemoved(t *testing.T) {
	t.Parallel()

	type testCase struct {
		request planmodifier.ListRequest
		wantErr bool
	}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"config": schema.ListAttribute{
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
	testPlan := func(config types.List) tfsdk.Plan {
		tfCOnfig, err := config.ToTerraformValue(t.Context())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		return tfsdk.Plan{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(t.Context()),
				map[string]tftypes.Value{
					"config": tfCOnfig,
				},
			),
		}
	}
	testState := func(config types.List) tfsdk.State {
		tfCfg, err := config.ToTerraformValue(t.Context())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		return tfsdk.State{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(t.Context()),
				map[string]tftypes.Value{
					"config": tfCfg,
				},
			),
		}
	}

	emptyList := types.ListValueMust(types.StringType, []attr.Value{})
	oneElement := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("value1")})
	twoElements := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("value1"), types.StringValue("value2")})

	tests := map[string]testCase{
		"unknown value on create": {
			request: planmodifier.ListRequest{
				State:      nullState,
				StateValue: types.ListNull(types.StringType),
				Plan:       testPlan(types.ListUnknown(types.StringType)),
				PlanValue:  types.ListUnknown(types.StringType),
			},
		},
		"unknown value on update": {
			request: planmodifier.ListRequest{
				State:      testState(oneElement),
				StateValue: oneElement,
				Plan:       testPlan(types.ListUnknown(types.StringType)),
				PlanValue:  types.ListUnknown(types.StringType),
			},
		},
		"null known value on create": {
			request: planmodifier.ListRequest{
				State:      nullState,
				StateValue: types.ListNull(types.StringType),
				Plan:       testPlan(types.ListNull(types.StringType)),
				PlanValue:  types.ListNull(types.StringType),
			},
		},
		"null known value on update no error": {
			request: planmodifier.ListRequest{
				State:      testState(twoElements),
				StateValue: twoElements,
				Plan:       testPlan(types.ListNull(types.StringType)),
				PlanValue:  types.ListNull(types.StringType),
			},
		},
		"empty known value on update no error": {
			request: planmodifier.ListRequest{
				State:      testState(twoElements),
				StateValue: twoElements,
				Plan:       testPlan(emptyList),
				PlanValue:  emptyList,
			},
		},
		"addition from null": {
			request: planmodifier.ListRequest{
				State:      nullState,
				StateValue: types.ListNull(types.StringType),
				Plan:       testPlan(oneElement),
				PlanValue:  oneElement,
			},
		},
		"addition from empty": {
			request: planmodifier.ListRequest{
				State:      testState(emptyList),
				StateValue: emptyList,
				Plan:       testPlan(oneElement),
				PlanValue:  oneElement,
			},
		},
		"addition from single": {
			request: planmodifier.ListRequest{
				State:      testState(oneElement),
				StateValue: oneElement,
				Plan:       testPlan(twoElements),
				PlanValue:  twoElements,
			},
		},
		"removal from multiple": {
			request: planmodifier.ListRequest{
				State:      testState(twoElements),
				StateValue: twoElements,
				Plan:       testPlan(oneElement),
				PlanValue:  oneElement,
			},
		},
		"removal to empty from single": {
			request: planmodifier.ListRequest{
				State:      testState(oneElement),
				StateValue: oneElement,
				Plan:       testPlan(emptyList),
				PlanValue:  emptyList,
			},
			wantErr: true,
		},
		"removal to null from single": {
			request: planmodifier.ListRequest{
				State:      testState(oneElement),
				StateValue: oneElement,
				Plan:       testPlan(types.ListNull(types.StringType)),
				PlanValue:  types.ListNull(types.StringType),
			},
			wantErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := planmodifier.ListResponse{}
			tflistplanmodifier.ErrorIfSingleBlockRemoved().PlanModifyList(t.Context(), test.request, &response)

			if got, want := response.Diagnostics.HasError(), test.wantErr; !cmp.Equal(got, want) {
				t.Errorf("ErrorIfSingleBlockRemoved() err %t, want %t", got, want)
			}
		})
	}
}
