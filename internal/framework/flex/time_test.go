// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

func TestTimeFromFramework(t *testing.T) {
	t.Parallel()

	v, _ := time.Parse(time.RFC3339, "2023-07-25T20:43:16Z")
	type testCase struct {
		input    timetypes.RFC3339
		expected *time.Time
	}
	tests := map[string]testCase{
		"valid time": {
			input:    timetypes.NewRFC3339ValueMust("2023-07-25T20:43:16Z"),
			expected: aws.Time(v),
		},
		"null time": {
			input:    timetypes.NewRFC3339Null(),
			expected: nil,
		},
		"unknown time": {
			input:    timetypes.NewRFC3339Unknown(),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.TimeFromFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestTimeToFramework(t *testing.T) {
	t.Parallel()

	v, _ := time.Parse(time.RFC3339, "2023-07-25T20:43:16Z")
	type testCase struct {
		input    *time.Time
		expected timetypes.RFC3339
	}
	tests := map[string]testCase{
		"valid time": {
			input:    &v,
			expected: timetypes.NewRFC3339ValueMust("2023-07-25T20:43:16Z"),
		},
		// "nil time": {
		// 	input:    nil,
		// 	expected: timetypes.NewRFC3339Null(),
		// },
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.TimeToFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
