// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestTimestampTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val      tftypes.Value
		expected attr.Value
	}{
		"null value": {
			val:      tftypes.NewValue(tftypes.String, nil),
			expected: fwtypes.NewTimestampNull(),
		},
		"unknown value": {
			val:      tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expected: fwtypes.NewTimestampUnknown(),
		},
		"valid timestamp UTC": {
			val:      tftypes.NewValue(tftypes.String, "2023-06-07T15:11:34Z"),
			expected: fwtypes.NewTimestampValue(time.Date(2023, time.June, 7, 15, 11, 34, 0, time.UTC)),
		},
		"valid timestamp zone": {
			val:      tftypes.NewValue(tftypes.String, "2023-06-07T15:11:34-06:00"),
			expected: fwtypes.NewTimestampValue(time.Date(2023, time.June, 7, 15, 11, 34, 0, locationFromString(t, "America/Regina"))), // No DST
		},
		"invalid value": {
			val:      tftypes.NewValue(tftypes.String, "not ok"),
			expected: fwtypes.NewTimestampUnknown(),
		},
		"invalid no zone": {
			val:      tftypes.NewValue(tftypes.String, "2023-06-07T15:11:34"),
			expected: fwtypes.NewTimestampUnknown(),
		},
		"invalid date only": {
			val:      tftypes.NewValue(tftypes.String, "2023-06-07Z"),
			expected: fwtypes.NewTimestampUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			val, err := fwtypes.TimestampType{}.ValueFromTerraform(ctx, test.val)

			if err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if !test.expected.Equal(val) {
				t.Errorf("unexpected diff\nwanted: %s\ngot:    %s", test.expected, val)
			}
		})
	}
}

func TestTimestampTypeValidate(t *testing.T) {
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
		"valid timestamp UTC": {
			val: tftypes.NewValue(tftypes.String, "2023-06-07T15:11:34Z"),
		},
		"valid timestamp zone": {
			val: tftypes.NewValue(tftypes.String, "2023-06-07T15:11:34-06:00"),
		},
		"invalid string": {
			val:         tftypes.NewValue(tftypes.String, "not ok"),
			expectError: true,
		},
		"invalid no zone": {
			val:         tftypes.NewValue(tftypes.String, "2023-06-07T15:11:34"),
			expectError: true,
		},
		"invalid date only": {
			val:         tftypes.NewValue(tftypes.String, "2023-06-07Z"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			diags := fwtypes.TimestampType{}.Validate(ctx, test.val, path.Root("test"))

			if !diags.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if diags.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %#v", diags)
			}
		})
	}
}

func locationFromString(t *testing.T, s string) *time.Location {
	location, err := time.LoadLocation(s)
	if err != nil {
		t.Fatalf("loading time.Location %q: %s", s, err)
	}

	return location
}
