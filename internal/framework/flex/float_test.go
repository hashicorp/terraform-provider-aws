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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Float64ToFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkFloat64ToFramework(b *testing.B) {
	ctx := context.Background()
	input := aws.Float64(42.1)
	for n := 0; n < b.N; n++ {
		r := flex.Float64ToFramework(ctx, input)
		if r.IsNull() {
			b.Fatal("should never see this")
		}
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Float64ToFrameworkLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkFloat64ToFrameworkLegacy(b *testing.B) {
	ctx := context.Background()
	input := aws.Float64(42.1)
	for n := 0; n < b.N; n++ {
		r := flex.Float64ToFrameworkLegacy(ctx, input)
		if r.IsNull() {
			b.Fatal("should never see this")
		}
	}
}

func TestFloat32ToFrameworkFloat64(t *testing.T) {
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Float32ToFrameworkFloat64(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkFloat32ToFrameworkFloat64(b *testing.B) {
	ctx := context.Background()
	input := aws.Float32(42.1)
	for n := 0; n < b.N; n++ {
		r := flex.Float32ToFrameworkFloat64(ctx, input)
		if r.IsNull() {
			b.Fatal("should never see this")
		}
	}
}

func TestFloat32ToFrameworkFloat64Legacy(t *testing.T) {
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Float32ToFrameworkFloat64Legacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkFloat32ToFrameworkFloat64Legacy(b *testing.B) {
	ctx := context.Background()
	input := aws.Float32(42.1)
	for n := 0; n < b.N; n++ {
		r := flex.Float32ToFrameworkFloat64Legacy(ctx, input)
		if r.IsNull() {
			b.Fatal("should never see this")
		}
	}
}
