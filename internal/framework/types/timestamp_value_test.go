// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestTimestampValueToTerraformValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		timestamp fwtypes.Timestamp
		expected  tftypes.Value
	}{
		"value": {
			timestamp: fwtypes.TimestampValue("2023-06-07T15:11:34Z"),
			expected:  tftypes.NewValue(tftypes.String, "2023-06-07T15:11:34Z"),
		},
		"null": {
			timestamp: fwtypes.TimestampNull(),
			expected:  tftypes.NewValue(tftypes.String, nil),
		},
		"unknown": {
			timestamp: fwtypes.TimestampUnknown(),
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
		timestamp fwtypes.Timestamp
		expected  types.String
	}{
		"value": {
			timestamp: fwtypes.TimestampValue("2023-06-07T15:11:34Z"),
			expected:  types.StringValue("2023-06-07T15:11:34Z"),
		},
		"null": {
			timestamp: fwtypes.TimestampNull(),
			expected:  types.StringNull(),
		},
		"unknown": {
			timestamp: fwtypes.TimestampUnknown(),
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

func TestTimestampValueValueTimestamp(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    fwtypes.Timestamp
		expected time.Time
	}{
		"known": {
			input:    fwtypes.TimestampValue("2023-06-07T15:11:34Z"),
			expected: errs.Must(time.Parse(time.RFC3339, "2023-06-07T15:11:34Z")),
		},
		"null": {
			input:    fwtypes.TimestampNull(),
			expected: time.Time{},
		},
		"unknown": {
			input:    fwtypes.TimestampUnknown(),
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
