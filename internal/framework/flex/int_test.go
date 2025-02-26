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

func TestInt64FromFramework(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Int64
		expected *int64
	}
	tests := map[string]testCase{
		"valid int64": {
			input:    types.Int64Value(42),
			expected: aws.Int64(42),
		},
		"zero int64": {
			input:    types.Int64Value(0),
			expected: aws.Int64(0),
		},
		"null int64": {
			input:    types.Int64Null(),
			expected: nil,
		},
		"unknown int64": {
			input:    types.Int64Unknown(),
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Int64FromFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkInt64FromFramework(b *testing.B) {
	ctx := context.Background()
	input := types.Int64Value(42)
	for n := 0; n < b.N; n++ {
		r := flex.Int64FromFramework(ctx, input)
		if r == nil {
			b.Fatal("should never see this")
		}
	}
}

func TestInt64ValueFromFramework(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Int64
		expected int64
	}
	tests := map[string]testCase{
		"valid int64": {
			input:    types.Int64Value(42),
			expected: 42,
		},
		"zero int64": {
			input:    types.Int64Value(0),
			expected: 0,
		},
		"null int64": {
			input:    types.Int64Null(),
			expected: 0,
		},
		"unknown int64": {
			input:    types.Int64Unknown(),
			expected: 0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Int64ValueFromFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkInt64ValueFromFramework(b *testing.B) {
	ctx := context.Background()
	input := types.Int64Value(42)
	for n := 0; n < b.N; n++ {
		r := flex.Int64ValueFromFramework(ctx, input)
		if r == 0 {
			b.Fatal("should never see this")
		}
	}
}

func TestInt64ToFramework(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *int64
		expected types.Int64
	}
	tests := map[string]testCase{
		"valid int64": {
			input:    aws.Int64(42),
			expected: types.Int64Value(42),
		},
		"zero int64": {
			input:    aws.Int64(0),
			expected: types.Int64Value(0),
		},
		"nil int64": {
			input:    nil,
			expected: types.Int64Null(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Int64ToFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkInt64ToFramework(b *testing.B) {
	ctx := context.Background()
	input := aws.Int64(42)
	for n := 0; n < b.N; n++ {
		r := flex.Int64ToFramework(ctx, input)
		if r.IsNull() {
			b.Fatal("should never see this")
		}
	}
}

func TestInt64ToFrameworkLegacy(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *int64
		expected types.Int64
	}
	tests := map[string]testCase{
		"valid int64": {
			input:    aws.Int64(42),
			expected: types.Int64Value(42),
		},
		"zero int64": {
			input:    aws.Int64(0),
			expected: types.Int64Value(0),
		},
		"nil int64": {
			input:    nil,
			expected: types.Int64Value(0),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Int64ToFrameworkLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkInt64ToFrameworkLegacy(b *testing.B) {
	ctx := context.Background()
	input := aws.Int64(42)
	for n := 0; n < b.N; n++ {
		r := flex.Int64ToFrameworkLegacy(ctx, input)
		if r.IsNull() {
			b.Fatal("should never see this")
		}
	}
}

func TestInt32ToFrameworkInt64(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *int32
		expected types.Int64
	}
	tests := map[string]testCase{
		"valid int64": {
			input:    aws.Int32(42),
			expected: types.Int64Value(42),
		},
		"zero int64": {
			input:    aws.Int32(0),
			expected: types.Int64Value(0),
		},
		"nil int64": {
			input:    nil,
			expected: types.Int64Null(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Int32ToFrameworkInt64(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkInt32ToFrameworkInt64(b *testing.B) {
	ctx := context.Background()
	input := aws.Int32(42)
	for n := 0; n < b.N; n++ {
		r := flex.Int32ToFrameworkInt64(ctx, input)
		if r.IsNull() {
			b.Fatal("should never see this")
		}
	}
}

func TestInt32ValueToFrameworkInt64(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    int32
		expected types.Int64
	}
	tests := map[string]testCase{
		"valid int64": {
			input:    42,
			expected: types.Int64Value(42),
		},
		"zero int64": {
			input:    0,
			expected: types.Int64Value(0),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Int32ValueToFrameworkInt64(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkInt32ValueToFrameworkInt64(b *testing.B) {
	ctx := context.Background()
	input := int32(42)
	for n := 0; n < b.N; n++ {
		r := flex.Int32ValueToFrameworkInt64(ctx, input)
		if r.IsNull() {
			b.Fatal("should never see this")
		}
	}
}

func TestInt32ToFrameworkInt64Legacy(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *int32
		expected types.Int64
	}
	tests := map[string]testCase{
		"valid int64": {
			input:    aws.Int32(42),
			expected: types.Int64Value(42),
		},
		"zero int64": {
			input:    aws.Int32(0),
			expected: types.Int64Value(0),
		},
		"nil int64": {
			input:    nil,
			expected: types.Int64Value(0),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Int32ToFrameworkInt64Legacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkInt32ValueToFrameworkInt64Legacy(b *testing.B) {
	ctx := context.Background()
	input := aws.Int32(42)
	for n := 0; n < b.N; n++ {
		r := flex.Int32ToFrameworkInt64Legacy(ctx, input)
		if r.IsNull() {
			b.Fatal("should never see this")
		}
	}
}

func TestInt32FromFrameworkInt64(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Int64
		expected *int32
	}
	tests := map[string]testCase{
		"valid int64": {
			input:    types.Int64Value(42),
			expected: aws.Int32(42),
		},
		"zero int64": {
			input:    types.Int64Value(0),
			expected: aws.Int32(0),
		},
		"null int64": {
			input:    types.Int64Null(),
			expected: nil,
		},
		"unknown int64": {
			input:    types.Int64Unknown(),
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Int32FromFrameworkInt64(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkInt32FromFrameworkInt64(b *testing.B) {
	ctx := context.Background()
	input := types.Int64Value(42)
	for n := 0; n < b.N; n++ {
		r := flex.Int32FromFrameworkInt64(ctx, input)
		if r == nil {
			b.Fatal("should never see this")
		}
	}
}

func TestInt32FromFrameworkLegacy(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Int32
		expected *int32
	}
	tests := map[string]testCase{
		"valid int32": {
			input:    types.Int32Value(42),
			expected: aws.Int32(42),
		},
		"zero int32": {
			input:    types.Int32Value(0),
			expected: nil,
		},
		"null int32": {
			input:    types.Int32Null(),
			expected: nil,
		},
		"unknown int32": {
			input:    types.Int32Unknown(),
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Int32FromFrameworkLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestInt32ValueFromFrameworkInt64(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Int64
		expected int32
	}
	tests := map[string]testCase{
		"valid int64": {
			input:    types.Int64Value(42),
			expected: 42,
		},
		"zero int64": {
			input:    types.Int64Value(0),
			expected: 0,
		},
		"null int64": {
			input:    types.Int64Null(),
			expected: 0,
		},
		"unknown int64": {
			input:    types.Int64Unknown(),
			expected: 0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Int32ValueFromFrameworkInt64(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkInt32ValueFromFrameworkInt64(b *testing.B) {
	ctx := context.Background()
	input := types.Int64Value(42)
	for n := 0; n < b.N; n++ {
		r := flex.Int32ValueFromFrameworkInt64(ctx, input)
		if r == 0 {
			b.Fatal("should never see this")
		}
	}
}

func TestInt32FromFramework(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Int32
		expected *int32
	}
	tests := map[string]testCase{
		"valid int64": {
			input:    types.Int32Value(42),
			expected: aws.Int32(42),
		},
		"zero int64": {
			input:    types.Int32Value(0),
			expected: aws.Int32(0),
		},
		"null int64": {
			input:    types.Int32Null(),
			expected: nil,
		},
		"unknown int64": {
			input:    types.Int32Unknown(),
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.Int32FromFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func BenchmarkInt32FromFramework(b *testing.B) {
	ctx := context.Background()
	input := types.Int32Value(42)
	for n := 0; n < b.N; n++ {
		r := flex.Int32FromFramework(ctx, input)
		if r == nil {
			b.Fatal("should never see this")
		}
	}
}
