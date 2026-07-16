// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
)

func TestDiscriminatorRequires(t *testing.T) {
	t.Parallel()

	type testCase struct {
		configValue types.String
		config      func(t *testing.T) tfsdk.Config
		expErrors   int
	}

	testCases := map[string]testCase{
		"discriminator-null": {
			configValue: types.StringNull(),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, nil, nil, nil) },
			expErrors:   0,
		},
		"discriminator-unknown": {
			configValue: types.StringUnknown(),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, tftypes.UnknownValue, nil, nil) },
			expErrors:   0,
		},
		"unmapped-discriminator": {
			configValue: types.StringValue("C"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "C", nil, nil) },
			expErrors:   0,
		},
		"discriminator-a-mapped-target-configured": {
			configValue: types.StringValue("A"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "A", "value", nil) },
			expErrors:   0,
		},
		"discriminator-a-mapped-target-null": {
			configValue: types.StringValue("A"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "A", nil, "value") },
			expErrors:   1,
		},
		"discriminator-b-mapped-target-null": {
			configValue: types.StringValue("B"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "B", "value", nil) },
			expErrors:   1,
		},
		"mapped-target-unknown": {
			configValue: types.StringValue("A"),
			config:      func(t *testing.T) tfsdk.Config { return testConfig(t, "A", tftypes.UnknownValue, nil) },
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

			tfstringvalidator.DiscriminatorRequires(map[string]path.Expression{
				"A": path.MatchRoot("a"),
				"B": path.MatchRoot("b"),
			}).ValidateString(t.Context(), request, &response)

			if got := response.Diagnostics.ErrorsCount(); got != tc.expErrors {
				t.Fatalf("expected %d error(s), got %d: %s", tc.expErrors, got, response.Diagnostics)
			}
		})
	}
}
