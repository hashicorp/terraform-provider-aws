// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestDurationTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val      tftypes.Value
		expected attr.Value
	}{
		"null value": {
			val:      tftypes.NewValue(tftypes.String, nil),
			expected: fwtypes.DurationNull(),
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: fwtypes.DurationUnknown(),
		},
		"valid duration": {
			val:      tftypes.NewValue(tftypes.String, "2h"),
			expected: fwtypes.DurationValue(2 * time.Hour),
		},
		"invalid duration": {
			val:      tftypes.NewValue(tftypes.String, "not ok"),
			expected: fwtypes.DurationUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			val, err := fwtypes.DurationType.ValueFromTerraform(ctx, test.val)

			if err != nil {
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
			t.Parallel()

			ctx := context.Background()

			diags := fwtypes.DurationType.Validate(ctx, test.val, path.Root("test"))

			if !diags.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if diags.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %#v", diags)
			}
		})
	}
}

func TestDurationToStringValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		duration fwtypes.Duration
		expected types.String
	}{
		"value": {
			// TODO: StringValue does not round-trip
			duration: durationFromString(t, "2h"),
			expected: types.StringValue("2h0m0s"),
		},
		"null": {
			duration: fwtypes.DurationNull(),
			expected: types.StringNull(),
		},
		"unknown": {
			duration: fwtypes.DurationUnknown(),
			expected: types.StringUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			s, _ := test.duration.ToStringValue(ctx)

			if !test.expected.Equal(s) {
				t.Fatalf("expected %#v to equal %#v", s, test.expected)
			}
		})
	}
}

func durationFromString(t *testing.T, s string) fwtypes.Duration {
	ctx := context.Background()

	val := tftypes.NewValue(tftypes.String, s)

	attr, err := fwtypes.DurationType.ValueFromTerraform(ctx, val)
	if err != nil {
		t.Fatalf("setting Duration: %s", err)
	}

	return attr.(fwtypes.Duration)
}
