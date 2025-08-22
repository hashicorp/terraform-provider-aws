// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkStringValueList(context.Background(), test.input)

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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringValueListLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
