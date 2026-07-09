// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
)

func testConfig(t *testing.T, gate any, a any, b any) tfsdk.Config {
	t.Helper()

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"gate": tftypes.String,
			"a":    tftypes.String,
			"b":    tftypes.String,
		},
	}

	return tfsdk.Config{
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"gate": schema.StringAttribute{},
				"a":    schema.StringAttribute{},
				"b":    schema.StringAttribute{},
			},
		},
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"gate": tftypes.NewValue(tftypes.String, gate),
			"a":    tftypes.NewValue(tftypes.String, a),
			"b":    tftypes.NewValue(tftypes.String, b),
		}),
	}
}

func TestAlsoRequiresWhenEquals(t *testing.T) {
	t.Parallel()

	type testCase struct {
		configValue types.String
		config      func(t *testing.T) tfsdk.Config
		expressions []path.Expression
		expErrors   int
	}

	testCases := map[string]testCase{
		"self-null": {
			configValue: types.StringNull(),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, nil, nil, nil) },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   0,
		},
		"self-unknown": {
			configValue: types.StringUnknown(),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, tftypes.UnknownValue, nil, nil) },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   0,
		},
		"self-not-equals-target-null": {
			configValue: types.StringValue("B"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "B", nil, nil) },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   0,
		},
		"self-not-equals-target-set": {
			configValue: types.StringValue("B"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "B", "value", nil) },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   0,
		},
		"self-equals-target-set": {
			configValue: types.StringValue("A"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "A", "value", nil) },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   0,
		},
		"self-equals-target-null": {
			configValue: types.StringValue("A"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "A", nil, "value") },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   1,
		},
		"self-equals-target-unknown": {
			configValue: types.StringValue("A"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "A", tftypes.UnknownValue, nil) },
			expressions: []path.Expression{path.MatchRoot("a")},
			expErrors:   0,
		},
		"self-equals-target-multiple-targets-one-null": {
			configValue: types.StringValue("A"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "A", "value", nil) },
			expressions: []path.Expression{path.MatchRoot("a"), path.MatchRoot("b")},
			expErrors:   1,
		},
		"self-equals-target-multiple-targets-other-null": {
			configValue: types.StringValue("A"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "A", nil, "value") },
			expressions: []path.Expression{path.MatchRoot("a"), path.MatchRoot("b")},
			expErrors:   1,
		},
		"self-equals-target-multiple-targets-all-null": {
			configValue: types.StringValue("A"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "A", nil, nil) },
			expressions: []path.Expression{path.MatchRoot("a"), path.MatchRoot("b")},
			expErrors:   2,
		},
		"self-equals-target-multiple-targets-all-set": {
			configValue: types.StringValue("A"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "A", "valueA", "value") },
			expressions: []path.Expression{path.MatchRoot("a"), path.MatchRoot("b")},
			expErrors:   0,
		},
		"self-equals-target-multiple-targets-one-unknown-other-null": {
			configValue: types.StringValue("A"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "A", nil, tftypes.UnknownValue) },
			expressions: []path.Expression{path.MatchRoot("a"), path.MatchRoot("b")},
			expErrors:   0,
		},
		"self-equals-target-multiple-targets-one-null-other-unknown": {
			configValue: types.StringValue("A"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "A", tftypes.UnknownValue, nil) },
			expressions: []path.Expression{path.MatchRoot("a"), path.MatchRoot("b")},
			expErrors:   0,
		},
		"self-equals-target-multiple-targets-all-unknown": {
			configValue: types.StringValue("A"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "A", tftypes.UnknownValue, tftypes.UnknownValue) },
			expressions: []path.Expression{path.MatchRoot("a"), path.MatchRoot("b")},
			expErrors:   0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := validator.StringRequest{
				ConfigValue:    tc.configValue,
				Config:         tc.config(t),
				Path:           path.Root("gate"),
				PathExpression: path.MatchRoot("gate"),
			}
			response := validator.StringResponse{}

			tfstringvalidator.AlsoRequiresWhenEquals("A", tc.expressions...).ValidateString(t.Context(), request, &response)

			if got := response.Diagnostics.ErrorsCount(); got != tc.expErrors {
				t.Fatalf("expected %d error(s), got %d: %s", tc.expErrors, got, response.Diagnostics)
			}

		})
	}
}
