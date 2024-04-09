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

func newMapValueMust(elements map[string]attr.Value) MapValue {
	return fwdiag.Must(NewMapValue(elements))
}

func TestTagMapEquality(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val1, val2     MapValue
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
			semanticEquals: true,
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
			semanticEquals: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
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
