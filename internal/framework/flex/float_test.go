// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

func TestFloat64ToFramework(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *float64
		expected types.Float64
	}
	tests := map[string]testCase{
		"valid float64": {
			input:    aws.Float64(42.1),
			expected: types.Float64Value(42.1),
		},
		"zero float64": {
			input:    aws.Float64(0),
			expected: types.Float64Value(0),
		},
		"nil float64": {
			input:    nil,
			expected: types.Float64Null(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Float64ToFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFloat64ToFrameworkLegacy(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *float64
		expected types.Float64
	}
	tests := map[string]testCase{
		"valid float64": {
			input:    aws.Float64(42.1),
			expected: types.Float64Value(42.1),
		},
		"zero float64": {
			input:    aws.Float64(0),
			expected: types.Float64Value(0),
		},
		"nil float64": {
			input:    nil,
			expected: types.Float64Value(0),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Float64ToFrameworkLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFloat32ToFramework(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *float32
		expected types.Float64
	}
	tests := map[string]testCase{
		"valid float32": {
			input:    aws.Float32(42.0),
			expected: types.Float64Value(42.0),
		},
		"zero float32": {
			input:    aws.Float32(0),
			expected: types.Float64Value(0),
		},
		"nil float32": {
			input:    nil,
			expected: types.Float64Null(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Float32ToFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFloat32ToFrameworkLegacy(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *float32
		expected types.Float64
	}
	tests := map[string]testCase{
		"valid float32": {
			input:    aws.Float32(42.0),
			expected: types.Float64Value(42.0),
		},
		"zero float32": {
			input:    aws.Float32(0),
			expected: types.Float64Value(0),
		},
		"nil float32": {
			input:    nil,
			expected: types.Float64Value(0),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Float32ToFrameworkLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
