// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package boolvalidator_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfboolvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/boolvalidator"
)

// Schema used in tests: a gating bool (`gate`) plus two siblings (`a` and `b`).
func testConfig(t *testing.T, gate any, a any, b any) tfsdk.Config {
	t.Helper()

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"gate": tftypes.Bool,
			"a":    tftypes.String,
			"b":    tftypes.String,
		},
	}

	return tfsdk.Config{
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"gate": schema.BoolAttribute{},
				"a":    schema.StringAttribute{},
				"b":    schema.StringAttribute{},
			},
		},
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"gate": tftypes.NewValue(tftypes.Bool, gate),
			"a":    tftypes.NewValue(tftypes.String, a),
			"b":    tftypes.NewValue(tftypes.String, b),
		}),
	}
}

func TestAlsoRequiresWhenTrue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		configValue types.Bool
		config      func(t *testing.T) tfsdk.Config
		expressions []path.Expression
		expErrors   int
	}

	testCases := map[string]testCase{
		"self-null": {
			configValue: types.BoolNull(),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, nil, nil, nil) },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   0,
		},
		"self-unknown": {
			configValue: types.BoolUnknown(),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, tftypes.UnknownValue, nil, nil) },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   0,
		},
		"self-false-target-null": {
			configValue: types.BoolValue(false),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, false, nil, nil) },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   0,
		},
		"self-false-target-set": {
			configValue: types.BoolValue(false),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, false, "value", nil) },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   0,
		},
		"self-true-target-set": {
			configValue: types.BoolValue(true),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, true, "value", nil) },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   0,
		},
		"self-true-target-null": {
			configValue: types.BoolValue(true),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, true, nil, nil) },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   1,
		},
		"self-true-target-unknown-defers": {
			configValue: types.BoolValue(true),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, true, tftypes.UnknownValue, nil) },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   0,
		},
		"self-true-multiple-targets-one-null": {
			configValue: types.BoolValue(true),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, true, "value", nil) },
			expressions: []path.Expression{path.MatchRoot("a"), path.MatchRoot("b")},
			expErrors:   1,
		},
		"self-true-multiple-targets-all-null": {
			configValue: types.BoolValue(true),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, true, nil, nil) },
			expressions: []path.Expression{path.MatchRoot("a"), path.MatchRoot("b")},
			expErrors:   2,
		},
		"self-true-self-reference-ignored": {
			// pointing at self is a no-op (does not crash or flag)
			configValue: types.BoolValue(true),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, true, nil, nil) },
			expressions: []path.Expression{path.MatchRoot("gate")},
			expErrors:   0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.BoolRequest{
				ConfigValue:    tc.configValue,
				Config:         tc.config(t),
				Path:           path.Root("gate"),
				PathExpression: path.MatchRoot("gate"),
			}
			resp := validator.BoolResponse{}

			tfboolvalidator.AlsoRequiresWhenTrue(tc.expressions...).ValidateBool(t.Context(), req, &resp)

			if got := resp.Diagnostics.ErrorsCount(); got != tc.expErrors {
				t.Fatalf("expected %d error(s), got %d: %s", tc.expErrors, got, resp.Diagnostics)
			}
		})
	}
}
