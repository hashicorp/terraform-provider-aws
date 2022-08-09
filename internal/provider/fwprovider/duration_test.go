// TODO: Move this to a shared 'types' package.
package fwprovider_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/fwprovider"
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
			expected: fwprovider.Duration{Null: true},
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: fwprovider.Duration{Unknown: true},
		},
		"valid duration": {
			val:      tftypes.NewValue(tftypes.String, "2h"),
			expected: fwprovider.Duration{Value: 2 * time.Hour},
		},
		"invalid duration": {
			val:         tftypes.NewValue(tftypes.String, "not ok"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			val, err := fwprovider.DurationType.ValueFromTerraform(ctx, test.val)

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
			ctx := context.Background()

			diags := fwprovider.DurationType.Validate(ctx, test.val, path.Root("test"))

			if !diags.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if diags.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %#v", diags)
			}
		})
	}
}
