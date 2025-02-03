// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tags

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

func newMapValueMust(elements map[string]attr.Value) Map {
	return fwdiag.Must(NewMapValue(elements))
}

func TestTagMapEquality(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val1, val2     Map
		equals         bool
		semanticEquals bool
	}
	tests := map[string]testCase{
		"equal": {
			val1: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
					"key2": types.StringValue("value2"),
				},
			),
			val2: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
					"key2": types.StringValue("value2"),
				},
			),
			equals:         true,
			semanticEquals: true,
		},

		"not equal": {
			val1: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
					"key2": types.StringValue("value2"),
				},
			),
			val2: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue("other1"),
					"key2": types.StringValue("value2"),
				},
			),
			equals:         false,
			semanticEquals: false,
		},

		"missing-set": {
			val1: newMapValueMust(
				map[string]attr.Value{
					"key2": types.StringValue("value2"),
				},
			),
			val2: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
					"key2": types.StringValue("value2"),
				},
			),
			equals:         false,
			semanticEquals: false,
		},

		"set-missing": {
			val1: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
					"key2": types.StringValue("value2"),
				},
			),
			val2: newMapValueMust(
				map[string]attr.Value{
					"key2": types.StringValue("value2"),
				},
			),
			equals:         false,
			semanticEquals: false,
		},

		"null-set": {
			val1: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringNull(),
					"key2": types.StringValue("value2"),
				},
			),
			val2: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
					"key2": types.StringValue("value2"),
				},
			),
			equals:         false,
			semanticEquals: false,
		},

		"set-null": {
			val1: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
					"key2": types.StringValue("value2"),
				},
			),
			val2: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringNull(),
					"key2": types.StringValue("value2"),
				},
			),
			equals:         false,
			semanticEquals: false,
		},

		"null-null": {
			val1: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringNull(),
					"key2": types.StringValue("value2"),
				},
			),
			val2: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringNull(),
					"key2": types.StringValue("value2"),
				},
			),
			equals:         true,
			semanticEquals: true,
		},

		"null-unset": {
			val1: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringNull(),
					"key2": types.StringValue("value2"),
				},
			),
			val2: newMapValueMust(
				map[string]attr.Value{
					"key2": types.StringValue("value2"),
				},
			),
			equals:         false,
			semanticEquals: false,
		},

		"unset-null": {
			val1: newMapValueMust(
				map[string]attr.Value{
					"key2": types.StringValue("value2"),
				},
			),
			val2: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringNull(),
					"key2": types.StringValue("value2"),
				},
			),
			equals:         false,
			semanticEquals: false,
		},

		"null-empty": {
			val1: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringNull(),
					"key2": types.StringValue("value2"),
				},
			),
			val2: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue(""),
					"key2": types.StringValue("value2"),
				},
			),
			equals:         false,
			semanticEquals: true,
		},

		"empty-null": {
			val1: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue(""),
					"key2": types.StringValue("value2"),
				},
			),
			val2: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringNull(),
					"key2": types.StringValue("value2"),
				},
			),
			equals:         false,
			semanticEquals: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			equals := test.val1.Equal(test.val2)

			if got, want := equals, test.equals; got != want {
				t.Errorf("Equal(%q, %q) = %v, want %v", test.val1, test.val2, got, want)
			}

			equals, _ = test.val1.MapSemanticEquals(ctx, test.val2)

			if got, want := equals, test.semanticEquals; got != want {
				t.Errorf("MapSemanticEquals(%q, %q) = %v, want %v", test.val1, test.val2, got, want)
			}
		})
	}
}

func TestTagMapIsWhollyKnown(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val        Map
		fullyKnown bool
	}
	tests := map[string]testCase{
		"map unknown": {
			val:        NewMapValueUnknown(),
			fullyKnown: false,
		},
		"map null": {
			val:        NewMapValueNull(),
			fullyKnown: true,
		},
		"map empty": {
			val:        newMapValueMust(map[string]attr.Value{}),
			fullyKnown: true,
		},
		"map single unknown element": {
			val: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringUnknown(),
				},
			),
			fullyKnown: false,
		},
		"map single null element": {
			val: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringNull(),
				},
			),
			fullyKnown: true,
		},
		"map single element": {
			val: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
				},
			),
			fullyKnown: true,
		},
		"map multiple elements, one unknown": {
			val: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
					"key2": types.StringUnknown(),
					"key3": types.StringValue("value3"),
				},
			),
			fullyKnown: false,
		},
		"map multiple elements, one null": {
			val: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
					"key2": types.StringNull(),
					"key3": types.StringValue("value3"),
				},
			),
			fullyKnown: true,
		},
		"map multiple elements": {
			val: newMapValueMust(
				map[string]attr.Value{
					"key1": types.StringValue("value1"),
					"key2": types.StringValue("value2"),
					"key3": types.StringValue("value3"),
				},
			),
			fullyKnown: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got, want := test.val.IsWhollyKnown(), test.fullyKnown; got != want {
				t.Errorf("IsFullyKnown(%q) = %v, want %v", test.val, got, want)
			}
		})
	}
}
