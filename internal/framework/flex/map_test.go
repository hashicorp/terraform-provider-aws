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

func TestExpandFrameworkStringMap(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Map
		expected map[string]*string
	}
	tests := map[string]testCase{
		"null": {
			input:    types.MapNull(types.StringType),
			expected: nil,
		},
		"unknown": {
			input:    types.MapUnknown(types.StringType),
			expected: nil,
		},
		"two elements": {
			input: types.MapValueMust(types.StringType, map[string]attr.Value{
				"one": types.StringValue("GET"),
				"two": types.StringValue("HEAD"),
			}),
			expected: aws.StringMap(map[string]string{
				"one": "GET",
				"two": "HEAD",
			}),
		},
		"zero elements": {
			input:    types.MapValueMust(types.StringType, map[string]attr.Value{}),
			expected: aws.StringMap(map[string]string{}),
		},
		"invalid element type": {
			input: types.MapValueMust(types.BoolType, map[string]attr.Value{
				"one": types.BoolValue(true),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkStringMap(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandFrameworkStringValueMap(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Map
		expected map[string]string
	}
	tests := map[string]testCase{
		"null": {
			input:    types.MapNull(types.StringType),
			expected: nil,
		},
		"unknown": {
			input:    types.MapUnknown(types.StringType),
			expected: nil,
		},
		"two elements": {
			input: types.MapValueMust(types.StringType, map[string]attr.Value{
				"one": types.StringValue("GET"),
				"two": types.StringValue("HEAD"),
			}),
			expected: map[string]string{
				"one": "GET",
				"two": "HEAD",
			},
		},
		"zero elements": {
			input:    types.MapValueMust(types.StringType, map[string]attr.Value{}),
			expected: map[string]string{},
		},
		"invalid element type": {
			input: types.MapValueMust(types.BoolType, map[string]attr.Value{
				"one": types.BoolValue(true),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkStringValueMap(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkStringMap(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    map[string]*string
		expected types.Map
	}
	tests := map[string]testCase{
		"two elements": {
			input: aws.StringMap(map[string]string{
				"one": "GET",
				"two": "HEAD",
			}),
			expected: types.MapValueMust(types.StringType, map[string]attr.Value{
				"one": types.StringValue("GET"),
				"two": types.StringValue("HEAD"),
			}),
		},
		"zero elements": {
			input:    aws.StringMap(map[string]string{}),
			expected: types.MapNull(types.StringType),
		},
		"nil map": {
			input:    nil,
			expected: types.MapNull(types.StringType),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringMap(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkStringValueMap(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    map[string]string
		expected types.Map
	}
	tests := map[string]testCase{
		"two elements": {
			input: map[string]string{
				"one": "GET",
				"two": "HEAD",
			},
			expected: types.MapValueMust(types.StringType, map[string]attr.Value{
				"one": types.StringValue("GET"),
				"two": types.StringValue("HEAD"),
			}),
		},
		"zero elements": {
			input:    map[string]string{},
			expected: types.MapNull(types.StringType),
		},
		"nil map": {
			input:    nil,
			expected: types.MapNull(types.StringType),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringValueMap(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkStringValueMapLegacy(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    map[string]string
		expected types.Map
	}
	tests := map[string]testCase{
		"two elements": {
			input: map[string]string{
				"one": "GET",
				"two": "HEAD",
			},
			expected: types.MapValueMust(types.StringType, map[string]attr.Value{
				"one": types.StringValue("GET"),
				"two": types.StringValue("HEAD"),
			}),
		},
		"zero elements": {
			input:    map[string]string{},
			expected: types.MapValueMust(types.StringType, map[string]attr.Value{}),
		},
		"nil map": {
			input:    nil,
			expected: types.MapValueMust(types.StringType, map[string]attr.Value{}),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringValueMapLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
