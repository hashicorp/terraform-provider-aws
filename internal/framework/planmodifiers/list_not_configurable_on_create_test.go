// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package planmodifiers_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwplanmodifiers "github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers"
)

func TestListNotConfigurableOnCreate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		request     planmodifier.ListRequest
		expectError bool
	}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.ListAttribute{
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
	testPlan := func(value types.List) tfsdk.Plan {
		tfValue, err := value.ToTerraformValue(context.Background())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		return tfsdk.Plan{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(context.Background()),
				map[string]tftypes.Value{
					"testattr": tfValue,
				},
			),
		}
	}
	testState := func(value types.List) tfsdk.State {
		tfValue, err := value.ToTerraformValue(context.Background())

		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}

		return tfsdk.State{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(context.Background()),
				map[string]tftypes.Value{
					"testattr": tfValue,
				},
			),
		}
	}

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
				State:      testState(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("computed-value")})),
				StateValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("computed-value")}),
				Plan:       testPlan(types.ListUnknown(types.StringType)),
				PlanValue:  types.ListUnknown(types.StringType),
			},
		},
		"null value on create": {
			request: planmodifier.ListRequest{
				State:      nullState,
				StateValue: types.ListNull(types.StringType),
				Plan:       testPlan(types.ListNull(types.StringType)),
				PlanValue:  types.ListNull(types.StringType),
			},
			expectError: true,
		},
		"null value on update": {
			request: planmodifier.ListRequest{
				State:      testState(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("computed-value")})),
				StateValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("computed-value")}),
				Plan:       testPlan(types.ListNull(types.StringType)),
				PlanValue:  types.ListNull(types.StringType),
			},
		},
		"non-null value on create": {
			request: planmodifier.ListRequest{
				State:      nullState,
				StateValue: types.ListNull(types.StringType),
				Plan:       testPlan(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("configured-value")})),
				PlanValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("configured-value")}),
			},
			expectError: true,
		},
		"non-null value on update": {
			request: planmodifier.ListRequest{
				State:      testState(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("computed-value")})),
				StateValue: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("computed-value")}),
				Plan:       testPlan(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("configured-value")})),
				PlanValue:  types.ListValueMust(types.StringType, []attr.Value{types.StringValue("configured-value")}),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := planmodifier.ListResponse{}
			fwplanmodifiers.ListNotConfigurableOnCreate().PlanModifyList(context.Background(), test.request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}
