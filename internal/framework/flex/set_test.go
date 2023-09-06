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
		expected flex.Set[string]
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

func TestFlattenFrameworkStringValueSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []string
		expected types.Set
	}
	tests := map[string]testCase{
		"two elements": {
			input: []string{"GET", "HEAD"},
			expected: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
		},
		"zero elements": {
			input:    []string{},
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

	type testCase struct {
		input    []string
		expected types.Set
	}
	tests := map[string]testCase{
		"two elements": {
			input: []string{"GET", "HEAD"},
			expected: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
		},
		"zero elements": {
			input:    []string{},
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
