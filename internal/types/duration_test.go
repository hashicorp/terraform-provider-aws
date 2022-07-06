package types_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
)

func TestDurationTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val         tftypes.Value
		expected    attr.Value
		expectError bool
	}{
		"null value": {
			val:      tftypes.NewValue(tftypes.String, nil),
			expected: types.Duration{Null: true},
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: types.Duration{Unknown: true},
		},
		"valid duration": {
			val:      tftypes.NewValue(tftypes.String, "2h"),
			expected: types.Duration{Value: 2 * time.Hour},
		},
		"invalid duration": {
			val:         tftypes.NewValue(tftypes.String, "not ok"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			val, err := types.DurationType.ValueFromTerraform(ctx, test.val)

			if err == nil && test.expectError {
				t.Fatal("expected error, got no error")
			}
			if err != nil && !test.expectError {
				t.Fatalf("got unexpected error: %s", err)
			}

			if diff := cmp.Diff(val, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestDurationTypeValidate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         tftypes.Value
		expectError bool
	}
	tests := map[string]testCase{
		"not a string": {
			val:         tftypes.NewValue(tftypes.Bool, true),
			expectError: true,
		},
		"unknown string": {
			val: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
		"null string": {
			val: tftypes.NewValue(tftypes.String, nil),
		},
		"valid string": {
			val: tftypes.NewValue(tftypes.String, "2h"),
		},
		"invalid string": {
			val:         tftypes.NewValue(tftypes.String, "not ok"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()

			attributePath := tftypes.NewAttributePath().WithAttributeName("test")
			diags := types.DurationType.Validate(ctx, test.val, attributePath)

			if !diags.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if diags.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %#v", diags)
			}
		})
	}
}
