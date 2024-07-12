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
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func TestExpandFrameworkInt32Set(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Set
		expected []*int32
	}
	tests := map[string]testCase{
		"null": {
			input:    types.SetNull(types.Int64Type),
			expected: nil,
		},
		"unknown": {
			input:    types.SetUnknown(types.Int64Type),
			expected: nil,
		},
		"two elements": {
			input: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			expected: []*int32{aws.Int32(1), aws.Int32(-1)},
		},
		"zero elements": {
			input:    types.SetValueMust(types.Int64Type, []attr.Value{}),
			expected: []*int32{},
		},
		"invalid element type": {
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkInt32Set(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandFrameworkInt32ValueSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Set
		expected []int32
	}
	tests := map[string]testCase{
		"null": {
			input:    types.SetNull(types.Int64Type),
			expected: nil,
		},
		"unknown": {
			input:    types.SetUnknown(types.Int64Type),
			expected: nil,
		},
		"two elements": {
			input: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			expected: []int32{1, -1},
		},
		"zero elements": {
			input:    types.SetValueMust(types.Int64Type, []attr.Value{}),
			expected: []int32{},
		},
		"invalid element type": {
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkInt32ValueSet(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandFrameworkInt64Set(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Set
		expected []*int64
	}
	tests := map[string]testCase{
		"null": {
			input:    types.SetNull(types.Int64Type),
			expected: nil,
		},
		"unknown": {
			input:    types.SetUnknown(types.Int64Type),
			expected: nil,
		},
		"two elements": {
			input: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			expected: []*int64{aws.Int64(1), aws.Int64(-1)},
		},
		"zero elements": {
			input:    types.SetValueMust(types.Int64Type, []attr.Value{}),
			expected: []*int64{},
		},
		"invalid element type": {
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkInt64Set(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandFrameworkInt64ValueSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Set
		expected []int64
	}
	tests := map[string]testCase{
		"null": {
			input:    types.SetNull(types.Int64Type),
			expected: nil,
		},
		"unknown": {
			input:    types.SetUnknown(types.Int64Type),
			expected: nil,
		},
		"two elements": {
			input: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			expected: []int64{1, -1},
		},
		"zero elements": {
			input:    types.SetValueMust(types.Int64Type, []attr.Value{}),
			expected: []int64{},
		},
		"invalid element type": {
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkInt64ValueSet(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandFrameworkStringSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Set
		expected []*string
	}
	tests := map[string]testCase{
		"null": {
			input:    types.SetNull(types.StringType),
			expected: nil,
		},
		"unknown": {
			input:    types.SetUnknown(types.StringType),
			expected: nil,
		},
		"two elements": {
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
			expected: []*string{aws.String("GET"), aws.String("HEAD")},
		},
		"zero elements": {
			input:    types.SetValueMust(types.StringType, []attr.Value{}),
			expected: []*string{},
		},
		"invalid element type": {
			input: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(42),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkStringSet(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandFrameworkStringValueSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Set
		expected itypes.Set[string]
	}
	tests := map[string]testCase{
		"null": {
			input:    types.SetNull(types.StringType),
			expected: nil,
		},
		"unknown": {
			input:    types.SetUnknown(types.StringType),
			expected: nil,
		},
		"two elements": {
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
			expected: []string{"GET", "HEAD"},
		},
		"zero elements": {
			input:    types.SetValueMust(types.StringType, []attr.Value{}),
			expected: []string{},
		},
		"invalid element type": {
			input: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(42),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkStringValueSet(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandFrameworkStringyValueSet(t *testing.T) {
	t.Parallel()

	type testEnum string
	var testVal1 testEnum = "testVal1"
	var testVal2 testEnum = "testVal2"

	type testCase struct {
		input    types.Set
		expected itypes.Set[testEnum]
	}
	tests := map[string]testCase{
		"null": {
			input:    types.SetNull(types.StringType),
			expected: nil,
		},
		"unknown": {
			input:    types.SetUnknown(types.StringType),
			expected: nil,
		},
		"two elements": {
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue(string(testVal1)),
				types.StringValue(string(testVal2)),
			}),
			expected: []testEnum{testVal1, testVal2},
		},
		"zero elements": {
			input:    types.SetValueMust(types.StringType, []attr.Value{}),
			expected: []testEnum{},
		},
		"invalid element type": {
			input: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(42),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkStringyValueSet[testEnum](context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkInt32Set(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []*int32
		expected types.Set
	}
	tests := map[string]testCase{
		"two elements": {
			input: []*int32{aws.Int32(1), aws.Int32(-1)},
			expected: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
		},
		"zero elements": {
			input:    []*int32{},
			expected: types.SetNull(types.Int64Type),
		},
		"nil array": {
			input:    nil,
			expected: types.SetNull(types.Int64Type),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkInt32Set(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkInt32ValueSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []int32
		expected types.Set
	}
	tests := map[string]testCase{
		"two elements": {
			input: []int32{1, -1},
			expected: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
		},
		"zero elements": {
			input:    []int32{},
			expected: types.SetNull(types.Int64Type),
		},
		"nil array": {
			input:    nil,
			expected: types.SetNull(types.Int64Type),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkInt32ValueSet(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkInt64Set(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []*int64
		expected types.Set
	}
	tests := map[string]testCase{
		"two elements": {
			input: []*int64{aws.Int64(1), aws.Int64(-1)},
			expected: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
		},
		"zero elements": {
			input:    []*int64{},
			expected: types.SetNull(types.Int64Type),
		},
		"nil array": {
			input:    nil,
			expected: types.SetNull(types.Int64Type),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkInt64Set(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkInt64ValueSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []int64
		expected types.Set
	}
	tests := map[string]testCase{
		"two elements": {
			input: []int64{1, -1},
			expected: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
		},
		"zero elements": {
			input:    []int64{},
			expected: types.SetNull(types.Int64Type),
		},
		"nil array": {
			input:    nil,
			expected: types.SetNull(types.Int64Type),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkInt64ValueSet(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkStringSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []*string
		expected types.Set
	}
	tests := map[string]testCase{
		"two elements": {
			input: []*string{aws.String("GET"), aws.String("HEAD")},
			expected: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
		},
		"zero elements": {
			input:    []*string{},
			expected: types.SetNull(types.StringType),
		},
		"nil array": {
			input:    nil,
			expected: types.SetNull(types.StringType),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringSet(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkStringValueSet(t *testing.T) {
	t.Parallel()

	// AWS enums use custom types with an underlying string type
	type custom string

	type testCase struct {
		input    []custom
		expected types.Set
	}
	tests := map[string]testCase{
		"two elements": {
			input: []custom{"GET", "HEAD"},
			expected: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
		},
		"zero elements": {
			input:    []custom{},
			expected: types.SetNull(types.StringType),
		},
		"nil array": {
			input:    nil,
			expected: types.SetNull(types.StringType),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringValueSet(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkStringValueSetLegacy(t *testing.T) {
	t.Parallel()

	// AWS enums use custom types with an underlying string type
	type custom string

	type testCase struct {
		input    []custom
		expected types.Set
	}
	tests := map[string]testCase{
		"two elements": {
			input: []custom{"GET", "HEAD"},
			expected: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
		},
		"zero elements": {
			input:    []custom{},
			expected: types.SetValueMust(types.StringType, []attr.Value{}),
		},
		"nil array": {
			input:    nil,
			expected: types.SetValueMust(types.StringType, []attr.Value{}),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringValueSetLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
