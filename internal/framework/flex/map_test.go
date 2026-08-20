// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
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
		"null element": {
			input: types.MapValueMust(types.StringType, map[string]attr.Value{
				"one": types.StringValue("GET"),
				"two": types.StringNull(),
			}),
			expected: map[string]*string{
				"one": aws.String("GET"),
				"two": nil,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkStringMap(t.Context(), test.input)

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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkStringValueMap(t.Context(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestExpandFrameworkStringValueListMap(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	type testCase struct {
		input    basetypes.MapValuable
		expected map[string][]string
	}
	tests := map[string]testCase{
		"null": {
			input:    types.MapNull(types.ListType{ElemType: types.StringType}),
			expected: nil,
		},
		"unknown": {
			input:    types.MapUnknown(types.ListType{ElemType: types.StringType}),
			expected: nil,
		},
		"two elements": {
			input: types.MapValueMust(types.ListType{ElemType: types.StringType}, map[string]attr.Value{
				"one": types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("HEAD"),
				}),
				"two": types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("GET"),
					types.StringValue("PUT"),
				}),
			}),
			expected: map[string][]string{
				"one": {"HEAD"},
				"two": {"GET", "PUT"},
			},
		},
		"zero elements": {
			input:    types.MapValueMust(types.ListType{ElemType: types.StringType}, map[string]attr.Value{}),
			expected: map[string][]string{},
		},
		"two elements MapOfListOfString": {
			input: fwtypes.NewMapValueOfMust[fwtypes.ListOfString](ctx, map[string]attr.Value{
				"one": flex.FlattenFrameworkStringValueListOfString(ctx, []string{"HEAD"}),
				"two": flex.FlattenFrameworkStringValueListOfString(ctx, []string{"GET", "PUT"}),
			}),
			expected: map[string][]string{
				"one": {"HEAD"},
				"two": {"GET", "PUT"},
			},
		},
		"invalid element type": {
			input: types.MapValueMust(types.BoolType, map[string]attr.Value{
				"one": types.BoolValue(true),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.ExpandFrameworkStringValueListMap(ctx, test.input)

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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringValueMap(t.Context(), test.input)

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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringValueMapLegacy(t.Context(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
