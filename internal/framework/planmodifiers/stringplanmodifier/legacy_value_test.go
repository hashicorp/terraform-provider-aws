// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package stringplanmodifier_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers/stringplanmodifier"
)

func TestLegacyValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  planmodifier.StringRequest
		expected planmodifier.StringResponse
	}{
		"null-state": {
			// when we first create the resource, use the zero value
			request: planmodifier.StringRequest{
				StateValue:  types.StringNull(),
				PlanValue:   types.StringUnknown(),
				ConfigValue: types.StringNull(),
			},
			expected: planmodifier.StringResponse{
				PlanValue: types.StringValue(""),
			},
		},
		"known-plan": {
			// if another plan modifier has set the planned value, use that
			request: planmodifier.StringRequest{
				StateValue:  types.StringValue("state"),
				PlanValue:   types.StringValue("plan"),
				ConfigValue: types.StringNull(),
			},
			expected: planmodifier.StringResponse{
				PlanValue: types.StringValue("plan"),
			},
		},
		"non-null-state-unknown-plan": {
			// if no value is set in the config on update, use the zero value.
			request: planmodifier.StringRequest{
				StateValue:  types.StringValue("state"),
				PlanValue:   types.StringUnknown(),
				ConfigValue: types.StringNull(),
			},
			expected: planmodifier.StringResponse{
				PlanValue: types.StringValue(""),
			},
		},
		"unknown-config": {
			// this is the situation in which a user is
			// interpolating into a field. We want that to still
			// show up as unknown, otherwise they'll get apply-time
			// errors for changing the value even though we knew it
			// was legitimately possible for it to change and the
			// provider can't prevent this from happening
			request: planmodifier.StringRequest{
				StateValue:  types.StringValue("state"),
				PlanValue:   types.StringUnknown(),
				ConfigValue: types.StringUnknown(),
			},
			expected: planmodifier.StringResponse{
				PlanValue: types.StringUnknown(),
			},
		},
		"under-list": {
			request: planmodifier.StringRequest{
				ConfigValue: types.StringNull(),
				Path:        path.Root("test").AtListIndex(0).AtName("nested_test"),
				PlanValue:   types.StringUnknown(),
				StateValue:  types.StringNull(),
			},
			expected: planmodifier.StringResponse{
				PlanValue: types.StringValue(""),
			},
		},
		"under-set": {
			request: planmodifier.StringRequest{
				ConfigValue: types.StringNull(),
				Path: path.Root("test").AtSetValue(
					types.SetValueMust(
						types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_test": types.StringType,
							},
						},
						[]attr.Value{
							types.ObjectValueMust(
								map[string]attr.Type{
									"nested_test": types.StringType,
								},
								map[string]attr.Value{
									"nested_test": types.StringUnknown(),
								},
							),
						},
					),
				).AtName("nested_test"),
				PlanValue:  types.StringUnknown(),
				StateValue: types.StringNull(),
			},
			expected: planmodifier.StringResponse{
				PlanValue: types.StringValue(""),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp := planmodifier.StringResponse{
				PlanValue: testCase.request.PlanValue,
			}

			stringplanmodifier.LegacyValue().PlanModifyString(context.Background(), testCase.request, &resp)

			if diff := cmp.Diff(testCase.expected, resp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
