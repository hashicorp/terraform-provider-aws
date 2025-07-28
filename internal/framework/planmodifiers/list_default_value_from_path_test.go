// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package planmodifiers_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwplanmodifiers "github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers"
)

func TestListDefaultValueFromPath(t *testing.T) {
	t.Parallel()

	type testCase struct {
		request  planmodifier.ListRequest
		expected planmodifier.ListResponse
	}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"src": schema.ListAttribute{
				ElementType: types.StringType,
			},
			"dst": schema.ListAttribute{
				ElementType: types.StringType,
			},
		},
	}
	nullState := tfsdk.State{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			testSchema.Type().TerraformType(context.Background()),
			nil,
		),
	}
	testPlan := func(src, dst types.List) tfsdk.Plan {
		tfSrc, err := src.ToTerraformValue(context.Background())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		tfDst, err := dst.ToTerraformValue(context.Background())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		return tfsdk.Plan{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(context.Background()),
				map[string]tftypes.Value{
					"src": tfSrc,
					"dst": tfDst,
				},
			),
		}
	}
	testState := func(src, dst types.List) tfsdk.State {
		tfSrc, err := src.ToTerraformValue(context.Background())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		tfDst, err := dst.ToTerraformValue(context.Background())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		return tfsdk.State{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(context.Background()),
				map[string]tftypes.Value{
					"src": tfSrc,
					"dst": tfDst,
				},
			),
		}
	}
	defaultValue := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("default-value")})
	computedValue := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("computed-value")})
	configuredValue := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("configured-value")})

	tests := map[string]testCase{
		"unknown value on create": {
			request: planmodifier.ListRequest{
				State:      nullState,
				StateValue: types.ListNull(types.StringType),
				Plan:       testPlan(defaultValue, types.ListUnknown(types.StringType)),
				PlanValue:  types.ListUnknown(types.StringType),
			},
			expected: planmodifier.ListResponse{
				PlanValue: defaultValue,
			},
		},
		"unknown value on update": {
			request: planmodifier.ListRequest{
				State:      testState(defaultValue, computedValue),
				StateValue: computedValue,
				Plan:       testPlan(defaultValue, types.ListUnknown(types.StringType)),
				PlanValue:  types.ListUnknown(types.StringType),
			},
			expected: planmodifier.ListResponse{
				PlanValue: defaultValue,
			},
		},
		"null known value on create": {
			request: planmodifier.ListRequest{
				State:      nullState,
				StateValue: types.ListNull(types.StringType),
				Plan:       testPlan(types.ListNull(types.StringType), types.ListUnknown(types.StringType)),
				PlanValue:  types.ListUnknown(types.StringType),
			},
			expected: planmodifier.ListResponse{
				PlanValue: types.ListNull(types.StringType),
			},
		},
		"null known value on update": {
			request: planmodifier.ListRequest{
				State:      testState(defaultValue, computedValue),
				StateValue: computedValue,
				Plan:       testPlan(types.ListNull(types.StringType), types.ListUnknown(types.StringType)),
				PlanValue:  types.ListUnknown(types.StringType),
			},
			expected: planmodifier.ListResponse{
				PlanValue: types.ListNull(types.StringType),
			},
		},
		"non-null known value on create": {
			request: planmodifier.ListRequest{
				State:      nullState,
				StateValue: types.ListNull(types.StringType),
				Plan:       testPlan(defaultValue, configuredValue),
				PlanValue:  configuredValue,
			},
			expected: planmodifier.ListResponse{
				PlanValue: configuredValue,
			},
		},
		"non-null known value on update": {
			request: planmodifier.ListRequest{
				State:      testState(defaultValue, computedValue),
				StateValue: computedValue,
				Plan:       testPlan(defaultValue, configuredValue),
				PlanValue:  configuredValue,
			},
			expected: planmodifier.ListResponse{
				PlanValue: configuredValue,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := planmodifier.ListResponse{
				PlanValue: test.request.PlanValue,
			}
			fwplanmodifiers.ListDefaultValueFromPath[types.List](path.Root("src")).PlanModifyList(context.Background(), test.request, &response)

			if diff := cmp.Diff(test.expected, response); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
