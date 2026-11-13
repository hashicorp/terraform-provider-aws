// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func TestSetDifference(t *testing.T) {
	t.Parallel()

	type testCase struct {
		a, b       types.Set
		expected   types.Set
		wantErrors int
	}
	tests := map[string]testCase{
		"a unknown b non-empty": {
			a: types.SetUnknown(types.StringType),
			b: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("x"),
			}),
			expected:   types.SetUnknown(types.StringType),
			wantErrors: 1,
		},
		"a non-empty b unknown": {
			a: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("x"),
			}),
			b:          types.SetUnknown(types.StringType),
			expected:   types.SetUnknown(types.StringType),
			wantErrors: 1,
		},
		"mismatched element types": {
			a: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("x"),
			}),
			b: types.SetValueMust(types.Int32Type, []attr.Value{
				types.Int32Value(2),
			}),
			expected:   types.SetUnknown(types.StringType),
			wantErrors: 1,
		},
		"both null treated as empty": {
			a:        types.SetNull(types.StringType),
			b:        types.SetNull(types.StringType),
			expected: types.SetValueMust(types.StringType, []attr.Value{}),
		},
		"a null b non-empty": {
			a: types.SetNull(types.StringType),
			b: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("x"),
			}),
			expected: types.SetValueMust(types.StringType, []attr.Value{}),
		},
		"a non-empty b null": {
			a: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("x"),
				types.StringValue("y"),
			}),
			b:        types.SetNull(types.StringType),
			expected: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("x"), types.StringValue("y")}),
		},
		"disjoint sets": {
			a: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("b"),
			}),
			b: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("c"),
				types.StringValue("d"),
			}),
			expected: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("b"),
			}),
		},
		"a subset of b": {
			a: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("b"),
			}),
			b: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("b"),
				types.StringValue("c"),
			}),
			expected: types.SetValueMust(types.StringType, []attr.Value{}),
		},
		"partial overlap": {
			a: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("b"),
				types.StringValue("c"),
			}),
			b: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("b"),
				types.StringValue("d"),
			}),
			expected: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("c"),
			}),
		},
		"equal sets": {
			a: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("b"),
			}),
			b: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("b"),
			}),
			expected: types.SetValueMust(types.StringType, []attr.Value{}),
		},
		"both empty": {
			a:        types.SetValueMust(types.StringType, []attr.Value{}),
			b:        types.SetValueMust(types.StringType, []attr.Value{}),
			expected: types.SetValueMust(types.StringType, []attr.Value{}),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := flex.SetDifference(context.Background(), test.a, test.b)

			if got := diags.ErrorsCount(); got != test.wantErrors {
				t.Errorf("expected %d error(s), got %d: %s", test.wantErrors, got, diags)
			}

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
		expected inttypes.Set[string]
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.FlattenFrameworkStringValueSetLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
