// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestTimestampValueToTerraformValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		timestamp fwtypes.TimestampValue
		expected  tftypes.Value
	}{
		"value": {
			timestamp: errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			expected:  tftypes.NewValue(tftypes.String, "2023-06-07T15:11:34Z"),
		},
		"null": {
			timestamp: fwtypes.NewTimestampNull(),
			expected:  tftypes.NewValue(tftypes.String, nil),
		},
		"unknown": {
			timestamp: fwtypes.NewTimestampUnknown(),
			expected:  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			got, err := test.timestamp.ToTerraformValue(ctx)
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}

			if !test.expected.Equal(got) {
				t.Fatalf("expected %#v to equal %#v", got, test.expected)
			}
		})
	}
}

func TestTimestampValueToStringValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		timestamp fwtypes.TimestampValue
		expected  types.String
	}{
		"value": {
			timestamp: errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			expected:  types.StringValue("2023-06-07T15:11:34Z"),
		},
		"null": {
			timestamp: fwtypes.NewTimestampNull(),
			expected:  types.StringNull(),
		},
		"unknown": {
			timestamp: fwtypes.NewTimestampUnknown(),
			expected:  types.StringUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			s, _ := test.timestamp.ToStringValue(ctx)

			if !test.expected.Equal(s) {
				t.Fatalf("expected %#v to equal %#v", s, test.expected)
			}
		})
	}
}

func TestTimestampValueEqual(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     fwtypes.TimestampValue
		candidate attr.Value
		expected  bool
	}{
		"known-known-same": {
			input:     errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			candidate: errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			expected:  true,
		},
		"known-known-diff": {
			input:     errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			candidate: errs.Must(fwtypes.NewTimestampValueString("1999-06-07T15:11:34Z")),
			expected:  false,
		},
		"known-unknown": {
			input:     errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			candidate: fwtypes.NewTimestampUnknown(),
			expected:  false,
		},
		"known-null": {
			input:     errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			candidate: fwtypes.NewTimestampNull(),
			expected:  false,
		},
		"unknown-known": {
			input:     fwtypes.NewTimestampUnknown(),
			candidate: errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			expected:  false,
		},
		"unknown-unknown": {
			input:     fwtypes.NewTimestampUnknown(),
			candidate: fwtypes.NewTimestampUnknown(),
			expected:  true,
		},
		"unknown-null": {
			input:     fwtypes.NewTimestampUnknown(),
			candidate: fwtypes.NewTimestampNull(),
			expected:  false,
		},
		"null-known": {
			input:     fwtypes.NewTimestampNull(),
			candidate: errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			expected:  false,
		},
		"null-unknown": {
			input:     fwtypes.NewTimestampNull(),
			candidate: fwtypes.NewTimestampUnknown(),
			expected:  false,
		},
		"null-null": {
			input:     fwtypes.NewTimestampNull(),
			candidate: fwtypes.NewTimestampNull(),
			expected:  true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.Equal(test.candidate)
			if got != test.expected {
				t.Errorf("expected %t, got %t", test.expected, got)
			}
		})
	}
}

func TestTimestampValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    fwtypes.TimestampValue
		expected bool
	}{
		"known": {
			input:    errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			expected: false,
		},
		"null": {
			input:    fwtypes.NewTimestampNull(),
			expected: true,
		},
		"unknown": {
			input:    fwtypes.NewTimestampUnknown(),
			expected: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.IsNull()

			if got != testCase.expected {
				t.Error("expected Null")
			}
		})
	}
}

func TestTimestampValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    fwtypes.TimestampValue
		expected bool
	}{
		"known": {
			input:    errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			expected: false,
		},
		"null": {
			input:    fwtypes.NewTimestampNull(),
			expected: false,
		},
		"unknown": {
			input:    fwtypes.NewTimestampUnknown(),
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.IsUnknown()

			if got != testCase.expected {
				t.Error("expected Unknown")
			}
		})
	}
}

func TestTimestampValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    fwtypes.TimestampValue
		expected string
	}
	tests := map[string]testCase{
		"known-non-empty": {
			input:    errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			expected: `"2023-06-07T15:11:34Z"`,
		},
		"unknown": {
			input:    fwtypes.NewTimestampUnknown(),
			expected: "<unknown>",
		},
		"null": {
			input:    fwtypes.NewTimestampNull(),
			expected: "<null>",
		},
		"zero-value": {
			input:    fwtypes.TimestampValue{},
			expected: `<null>`,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.String()
			if got != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, got)
			}
		})
	}
}

func TestTimestampValueValueString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    fwtypes.TimestampValue
		expected string
	}{
		"known": {
			input:    errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			expected: "2023-06-07T15:11:34Z",
		},
		"null": {
			input:    fwtypes.NewTimestampNull(),
			expected: "",
		},
		"unknown": {
			input:    fwtypes.NewTimestampUnknown(),
			expected: "",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueString()

			if got != testCase.expected {
				t.Errorf("Expected %q, got %q", testCase.expected, got)
			}
		})
	}
}

func TestTimestampValueValueTimestamp(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    fwtypes.TimestampValue
		expected time.Time
	}{
		"known": {
			input:    errs.Must(fwtypes.NewTimestampValueString("2023-06-07T15:11:34Z")),
			expected: errs.Must(time.Parse(time.RFC3339, "2023-06-07T15:11:34Z")),
		},
		"null": {
			input:    fwtypes.NewTimestampNull(),
			expected: time.Time{},
		},
		"unknown": {
			input:    fwtypes.NewTimestampUnknown(),
			expected: time.Time{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueTimestamp()

			if !testCase.expected.Equal(got) {
				t.Errorf("Expected %q, got %q", testCase.expected, got)
			}
		})
	}
}
