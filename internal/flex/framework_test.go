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

func TestBoolFromFramework(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Bool
		expected *bool
	}
	tests := map[string]testCase{
		"valid bool": {
			input:    types.BoolValue(true),
			expected: aws.Bool(true),
		},
		"null bool": {
			input:    types.BoolNull(),
			expected: nil,
		},
		"unknown bool": {
			input:    types.BoolUnknown(),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := BoolFromFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestInt64FromFramework(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.Int64
		expected *int64
	}
	tests := map[string]testCase{
		"valid int64": {
			input:    types.Int64Value(42),
			expected: aws.Int64(42),
		},
		"zero int64": {
			input:    types.Int64Value(0),
			expected: aws.Int64(0),
		},
		"null int64": {
			input:    types.Int64Null(),
			expected: nil,
		},
		"unknown int64": {
			input:    types.Int64Unknown(),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := Int64FromFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStringFromFramework(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.String
		expected *string
	}
	tests := map[string]testCase{
		"valid string": {
			input:    types.StringValue("TEST"),
			expected: aws.String("TEST"),
		},
		"empty string": {
			input:    types.StringValue(""),
			expected: aws.String(""),
		},
		"null string": {
			input:    types.StringNull(),
			expected: nil,
		},
		"unknown string": {
			input:    types.StringUnknown(),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := StringFromFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestBoolToFramework(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *bool
		expected types.Bool
	}
	tests := map[string]testCase{
		"valid bool": {
			input:    aws.Bool(true),
			expected: types.BoolValue(true),
		},
		"nil bool": {
			input:    nil,
			expected: types.BoolNull(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := BoolToFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestBoolToFrameworkLegacy(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *bool
		expected types.Bool
	}
	tests := map[string]testCase{
		"valid bool": {
			input:    aws.Bool(true),
			expected: types.BoolValue(true),
		},
		"nil bool": {
			input:    nil,
			expected: types.BoolValue(false),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := BoolToFrameworkLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestInt64ToFramework(t *testing.T) {
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
		"nil int64": {
			input:    nil,
			expected: types.Int64Null(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := Int64ToFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestInt64ToFrameworkLegacy(t *testing.T) {
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
		"nil int64": {
			input:    nil,
			expected: types.Int64Value(0),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := Int64ToFrameworkLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStringToFramework(t *testing.T) {
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
			got := StringToFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStringToFrameworkLegacy(t *testing.T) {
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
			expected: types.StringValue(""),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := StringToFrameworkLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStringToFrameworkWithTransform(t *testing.T) {
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
			got := StringToFrameworkWithTransform(context.Background(), test.input, strings.ToLower)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
