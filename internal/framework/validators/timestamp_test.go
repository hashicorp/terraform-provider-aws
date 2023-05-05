package validators_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

func TestUTCTimestampValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val                 types.String
		expectedDiagnostics diag.Diagnostics
	}

	tests := map[string]testCase{
		"unknown String": {
			val: types.StringUnknown(),
		},
		"null String": {
			val: types.StringNull(),
		},
		"valid timestamp": {
			val: types.StringValue("2023-02-28T14:04:05Z"),
		},
		"invalid timestamp": {
			val: types.StringValue("02 Jan 06 15:04 -0700"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid UTC Timestamp, got: 02 Jan 06 15:04 -0700`,
				),
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			fwvalidators.UTCTimestamp().ValidateString(ctx, request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestOnceADayWindowFormatValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val                 types.String
		expectedDiagnostics diag.Diagnostics
	}

	tests := map[string]testCase{
		"unknown String": {
			val: types.StringUnknown(),
		},
		"null String": {
			val: types.StringNull(),
		},
		"valid format": {
			val: types.StringValue("04:00-05:00"),
		},
		"invalid format": {
			val: types.StringValue("24:00-25:00"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must satisfy the format of "hh24:mi-hh24:mi", got: 24:00-25:00`,
				),
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			fwvalidators.OnceADayWindowFormat().ValidateString(context.Background(), request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestOnceAWeekWindowFormatValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val                 types.String
		expectedDiagnostics diag.Diagnostics
	}

	tests := map[string]testCase{
		"unknown String": {
			val: types.StringUnknown(),
		},
		"null String": {
			val: types.StringNull(),
		},
		"valid format": {
			val: types.StringValue("sun:04:00-sun:05:00"),
		},
		"invalid format": {
			val: types.StringValue("sun:04:00-sun:04:60"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must satisfy the format of "ddd:hh24:mi-ddd:hh24:mi", got: sun:04:00-sun:04:60`,
				),
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			fwvalidators.OnceAWeekWindowFormat().ValidateString(context.Background(), request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
