// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestCloudVmClusterSystemVersionPlanModifier(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r := &resourceCloudVmCluster{}
	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

	attribute, ok := schemaResponse.Schema.Attributes["system_version"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("expected system_version to be a schema.StringAttribute, got %T", schemaResponse.Schema.Attributes["system_version"])
	}

	nonNullPlan := tfsdk.Plan{
		Raw: tftypes.NewValue(tftypes.String, "plan"),
	}
	nonNullState := tfsdk.State{
		Raw: tftypes.NewValue(tftypes.String, "state"),
	}

	tests := map[string]struct {
		config          types.String
		plan            types.String
		state           types.String
		requiresReplace bool
	}{
		"unconfigured value stays unknown": {
			config:          types.StringNull(),
			plan:            types.StringUnknown(),
			state:           types.StringValue("old-version"),
			requiresReplace: false,
		},
		"configured value change requires replacement": {
			config:          types.StringValue("new-version"),
			plan:            types.StringValue("new-version"),
			state:           types.StringValue("old-version"),
			requiresReplace: true,
		},
		"unchanged configured value does not require replacement": {
			config:          types.StringValue("old-version"),
			plan:            types.StringValue("old-version"),
			state:           types.StringValue("old-version"),
			requiresReplace: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := planmodifier.StringRequest{
				ConfigValue: test.config,
				Plan:        nonNullPlan,
				PlanValue:   test.plan,
				State:       nonNullState,
				StateValue:  test.state,
			}
			response := &planmodifier.StringResponse{
				PlanValue: test.plan,
			}

			for _, modifier := range attribute.PlanModifiers {
				modifier.PlanModifyString(ctx, request, response)
				request.PlanValue = response.PlanValue
			}

			if got, want := response.RequiresReplace, test.requiresReplace; got != want {
				t.Errorf("expected RequiresReplace to be %t, got %t", want, got)
			}
			if !response.PlanValue.Equal(test.plan) {
				t.Errorf("expected planned system_version %s, got %s", test.plan, response.PlanValue)
			}
		})
	}
}

func TestCloudVmClusterSystemVersionForCreate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tests := map[string]struct {
		config  types.String
		plan    types.String
		want    string
		wantNil bool
	}{
		"unconfigured omits stale planned value": {
			config:  types.StringNull(),
			plan:    types.StringValue("old-computed-version"),
			wantNil: true,
		},
		"configured uses apply-resolved planned value": {
			config: types.StringValue("configured-version"),
			plan:   types.StringValue("resolved-version"),
			want:   "resolved-version",
		},
		"unknown configuration uses apply-resolved planned value": {
			config: types.StringUnknown(),
			plan:   types.StringValue("resolved-version"),
			want:   "resolved-version",
		},
		"unknown planned value is omitted": {
			config:  types.StringValue("configured-version"),
			plan:    types.StringUnknown(),
			wantNil: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := cloudVmClusterSystemVersionForCreate(ctx, test.config, test.plan)

			switch {
			case test.wantNil && got != nil:
				t.Errorf("expected no system version, got %q", *got)
			case !test.wantNil && got == nil:
				t.Errorf("expected system version %q, got nil", test.want)
			case !test.wantNil && *got != test.want:
				t.Errorf("expected system version %q, got %q", test.want, *got)
			}
		})
	}
}

func TestCloudVmClusterSystemVersionForUpdate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config types.String
		plan   types.String
		state  types.String
		want   types.String
	}{
		"unconfigured unknown plan preserves state": {
			config: types.StringNull(),
			plan:   types.StringUnknown(),
			state:  types.StringValue("old-version"),
			want:   types.StringValue("old-version"),
		},
		"unknown configuration remains unknown": {
			config: types.StringUnknown(),
			plan:   types.StringUnknown(),
			state:  types.StringValue("old-version"),
			want:   types.StringUnknown(),
		},
		"configured plan is unchanged": {
			config: types.StringValue("configured-version"),
			plan:   types.StringValue("configured-version"),
			state:  types.StringValue("old-version"),
			want:   types.StringValue("configured-version"),
		},
		"known unconfigured plan is unchanged": {
			config: types.StringNull(),
			plan:   types.StringValue("old-version"),
			state:  types.StringValue("old-version"),
			want:   types.StringValue("old-version"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := cloudVmClusterSystemVersionForUpdate(test.config, test.plan, test.state)
			if !got.Equal(test.want) {
				t.Errorf("expected system version %s, got %s", test.want, got)
			}
		})
	}
}
