// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

func TestExpandFrameworkInt32List(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.List
		expected []*int32
	}
	tests := map[string]testCase{
		"null": {
			input:    types.ListNull(types.Int64Type),
			expected: nil,
		},
		"unknown": {
			input:    types.ListUnknown(types.Int64Type),
			expected: nil,
		},
		"two elements": {
			input: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			expected: []*int32{aws.Int32(1), aws.Int32(-1)},
		},
		"zero elements": {
			input:    types.ListValueMust(types.Int64Type, []attr.Value{}),
			expected: []*int32{},
		},
		"invalid element type": {
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkInt32List(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandFrameworkInt32ValueList(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.List
		expected []int32
	}
	tests := map[string]testCase{
		"null": {
			input:    types.ListNull(types.Int64Type),
			expected: nil,
		},
		"unknown": {
			input:    types.ListUnknown(types.Int64Type),
			expected: nil,
		},
		"two elements": {
			input: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			expected: []int32{1, -1},
		},
		"zero elements": {
			input:    types.ListValueMust(types.Int64Type, []attr.Value{}),
			expected: []int32{},
		},
		"invalid element type": {
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkInt32ValueList(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandFrameworkInt64List(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.List
		expected []*int64
	}
	tests := map[string]testCase{
		"null": {
			input:    types.ListNull(types.Int64Type),
			expected: nil,
		},
		"unknown": {
			input:    types.ListUnknown(types.Int64Type),
			expected: nil,
		},
		"two elements": {
			input: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			expected: []*int64{aws.Int64(1), aws.Int64(-1)},
		},
		"zero elements": {
			input:    types.ListValueMust(types.Int64Type, []attr.Value{}),
			expected: []*int64{},
		},
		"invalid element type": {
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkInt64List(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandFrameworkInt64ValueList(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.List
		expected []int64
	}
	tests := map[string]testCase{
		"null": {
			input:    types.ListNull(types.Int64Type),
			expected: nil,
		},
		"unknown": {
			input:    types.ListUnknown(types.Int64Type),
			expected: nil,
		},
		"two elements": {
			input: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			expected: []int64{1, -1},
		},
		"zero elements": {
			input:    types.ListValueMust(types.Int64Type, []attr.Value{}),
			expected: []int64{},
		},
		"invalid element type": {
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkInt64ValueList(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandFrameworkStringList(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.List
		expected []*string
	}
	tests := map[string]testCase{
		"null": {
			input:    types.ListNull(types.StringType),
			expected: nil,
		},
		"unknown": {
			input:    types.ListUnknown(types.StringType),
			expected: nil,
		},
		"two elements": {
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
			expected: []*string{aws.String("GET"), aws.String("HEAD")},
		},
		"zero elements": {
			input:    types.ListValueMust(types.StringType, []attr.Value{}),
			expected: []*string{},
		},
		"invalid element type": {
			input: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(42),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkStringList(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandFrameworkStringValueList(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.List
		expected []string
	}
	tests := map[string]testCase{
		"null": {
			input:    types.ListNull(types.StringType),
			expected: nil,
		},
		"unknown": {
			input:    types.ListUnknown(types.StringType),
			expected: nil,
		},
		"two elements": {
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
			expected: []string{"GET", "HEAD"},
		},
		"zero elements": {
			input:    types.ListValueMust(types.StringType, []attr.Value{}),
			expected: []string{},
		},
		"invalid element type": {
			input: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(42),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkStringValueList(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkInt32List(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []*int32
		expected types.List
	}
	tests := map[string]testCase{
		"two elements": {
			input: []*int32{aws.Int32(1), aws.Int32(-1)},
			expected: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
		},
		"zero elements": {
			input:    []*int32{},
			expected: types.ListNull(types.Int64Type),
		},
		"nil array": {
			input:    nil,
			expected: types.ListNull(types.Int64Type),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkInt32List(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkInt32ValueList(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []int32
		expected types.List
	}
	tests := map[string]testCase{
		"two elements": {
			input: []int32{1, -1},
			expected: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
		},
		"zero elements": {
			input:    []int32{},
			expected: types.ListNull(types.Int64Type),
		},
		"nil array": {
			input:    nil,
			expected: types.ListNull(types.Int64Type),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkInt32ValueList(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkInt64List(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []*int64
		expected types.List
	}
	tests := map[string]testCase{
		"two elements": {
			input: []*int64{aws.Int64(1), aws.Int64(-1)},
			expected: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
		},
		"zero elements": {
			input:    []*int64{},
			expected: types.ListNull(types.Int64Type),
		},
		"nil array": {
			input:    nil,
			expected: types.ListNull(types.Int64Type),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkInt64List(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkInt64ValueList(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []int64
		expected types.List
	}
	tests := map[string]testCase{
		"two elements": {
			input: []int64{1, -1},
			expected: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
		},
		"zero elements": {
			input:    []int64{},
			expected: types.ListNull(types.Int64Type),
		},
		"nil array": {
			input:    nil,
			expected: types.ListNull(types.Int64Type),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkInt64ValueList(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkStringList(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []*string
		expected types.List
	}
	tests := map[string]testCase{
		"two elements": {
			input: []*string{aws.String("GET"), aws.String("HEAD")},
			expected: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
		},
		"zero elements": {
			input:    []*string{},
			expected: types.ListNull(types.StringType),
		},
		"nil array": {
			input:    nil,
			expected: types.ListNull(types.StringType),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringList(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkStringListLegacy(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []*string
		expected types.List
	}
	tests := map[string]testCase{
		"two elements": {
			input: []*string{aws.String("GET"), aws.String("HEAD")},
			expected: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
		},
		"zero elements": {
			input:    []*string{},
			expected: types.ListValueMust(types.StringType, []attr.Value{}),
		},
		"nil array": {
			input:    nil,
			expected: types.ListValueMust(types.StringType, []attr.Value{}),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringListLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkStringValueList(t *testing.T) {
	t.Parallel()

	// AWS enums use custom types with an underlying string type
	type custom string

	type testCase struct {
		input    []custom
		expected types.List
	}
	tests := map[string]testCase{
		"two elements": {
			input: []custom{"GET", "HEAD"},
			expected: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
		},
		"zero elements": {
			input:    []custom{},
			expected: types.ListNull(types.StringType),
		},
		"nil array": {
			input:    nil,
			expected: types.ListNull(types.StringType),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringValueList(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkStringValueListLegacy(t *testing.T) {
	t.Parallel()

	// AWS enums use custom types with an underlying string type
	type custom string

	type testCase struct {
		input    []custom
		expected types.List
	}
	tests := map[string]testCase{
		"two elements": {
			input: []custom{"GET", "HEAD"},
			expected: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
		},
		"zero elements": {
			input:    []custom{},
			expected: types.ListValueMust(types.StringType, []attr.Value{}),
		},
		"nil array": {
			input:    nil,
			expected: types.ListValueMust(types.StringType, []attr.Value{}),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringValueListLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
