package flex

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpandFrameworkStringSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Set
		expected []*string
	}
	tests := map[string]testCase{
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
			got := ExpandFrameworkStringSet(context.Background(), test.input)

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
		expected []string
	}
	tests := map[string]testCase{
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
			got := ExpandFrameworkStringValueSet(context.Background(), test.input)

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
			got := ExpandFrameworkStringValueMap(context.Background(), test.input)

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
			got := FlattenFrameworkStringList(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkStringValueList(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []string
		expected types.List
	}
	tests := map[string]testCase{
		"two elements": {
			input: []string{"GET", "HEAD"},
			expected: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("GET"),
				types.StringValue("HEAD"),
			}),
		},
		"zero elements": {
			input:    []string{},
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
			got := FlattenFrameworkStringValueList(context.Background(), test.input)

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
			got := FlattenFrameworkStringValueSet(context.Background(), test.input)

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
			got := FlattenFrameworkStringValueMap(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestToFrameworkInt64Value(t *testing.T) {
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
		"nil string": {
			input:    nil,
			expected: types.Int64Null(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := ToFrameworkInt64Value(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestToFrameworkStringValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *string
		expected types.String
	}
	tests := map[string]testCase{
		"valid string": {
			input:    aws.String("TEST"),
			expected: types.StringValue("TEST"),
		},
		"empty string": {
			input:    aws.String(""),
			expected: types.StringValue(""),
		},
		"nil string": {
			input:    nil,
			expected: types.StringNull(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := ToFrameworkStringValue(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestToFrameworkStringValueWithTransform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *string
		expected types.String
	}
	tests := map[string]testCase{
		"valid string": {
			input:    aws.String("TEST"),
			expected: types.StringValue("test"),
		},
		"empty string": {
			input:    aws.String(""),
			expected: types.StringValue(""),
		},
		"nil string": {
			input:    nil,
			expected: types.StringNull(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := ToFrameworkStringValueWithTransform(context.Background(), test.input, strings.ToLower)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
