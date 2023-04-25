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

func TestExpandFrameworkStringList(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.List
		expected []*string
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
			expected: []*string{aws.String("GET"), aws.String("HEAD")},
		},
		"zero elements": {
			input:    types.ListValueMust(types.StringType, []attr.Value{}),
			expected: []*string{},
		},
		"invalid element type": {
			input: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(42),
			}),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := ExpandFrameworkStringList(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

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
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := ExpandFrameworkStringValueList(context.Background(), test.input)

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
		expected Set[string]
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
			expected: types.ListNull(types.StringType),
		},
		"nil array": {
			input:    nil,
			expected: types.ListNull(types.StringType),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := FlattenFrameworkStringList(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkStringListLegacy(t *testing.T) {
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
			t.Parallel()

			got := FlattenFrameworkStringListLegacy(context.Background(), test.input)

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
			expected: types.ListNull(types.StringType),
		},
		"nil array": {
			input:    nil,
			expected: types.ListNull(types.StringType),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := FlattenFrameworkStringValueList(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFlattenFrameworkStringValueListLegacy(t *testing.T) {
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
			t.Parallel()

			got := FlattenFrameworkStringValueListLegacy(context.Background(), test.input)

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

			got := FlattenFrameworkStringValueSet(context.Background(), test.input)

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

			got := FlattenFrameworkStringValueSetLegacy(context.Background(), test.input)

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

			got := FlattenFrameworkStringValueMapLegacy(context.Background(), test.input)

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
			t.Parallel()

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
			t.Parallel()

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
			t.Parallel()

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
			t.Parallel()

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
			t.Parallel()

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
			t.Parallel()

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
			t.Parallel()

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
			t.Parallel()

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
			t.Parallel()

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
			t.Parallel()

			got := StringToFrameworkWithTransform(context.Background(), test.input, strings.ToLower)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStringValueToFramework(t *testing.T) {
	t.Parallel()

	// AWS enums use custom types with an underlying string type
	type custom string

	type testCase struct {
		input    custom
		expected types.String
	}
	tests := map[string]testCase{
		"valid": {
			input:    "TEST",
			expected: types.StringValue("TEST"),
		},
		"empty": {
			input:    "",
			expected: types.StringNull(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := StringValueToFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStringValueToFrameworkLegacy(t *testing.T) {
	t.Parallel()

	// AWS enums use custom types with an underlying string type
	type custom string

	type testCase struct {
		input    custom
		expected types.String
	}
	tests := map[string]testCase{
		"valid": {
			input:    "TEST",
			expected: types.StringValue("TEST"),
		},
		"empty": {
			input:    "",
			expected: types.StringValue(""),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := StringValueToFrameworkLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFloat64ToFramework(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *float64
		expected types.Float64
	}
	tests := map[string]testCase{
		"valid float64": {
			input:    aws.Float64(42.1),
			expected: types.Float64Value(42.1),
		},
		"zero float64": {
			input:    aws.Float64(0),
			expected: types.Float64Value(0),
		},
		"nil float64": {
			input:    nil,
			expected: types.Float64Null(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := Float64ToFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFloat64ToFrameworkLegacy(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *float64
		expected types.Float64
	}
	tests := map[string]testCase{
		"valid int64": {
			input:    aws.Float64(42.1),
			expected: types.Float64Value(42.1),
		},
		"zero int64": {
			input:    aws.Float64(0),
			expected: types.Float64Value(0),
		},
		"nil int64": {
			input:    nil,
			expected: types.Float64Value(0),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := Float64ToFrameworkLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestSet_Difference_strings(t *testing.T) {
	t.Parallel()

	type testCase struct {
		original Set[string]
		new      Set[string]
		expected Set[string]
	}
	tests := map[string]testCase{
		"nil": {
			original: nil,
			new:      nil,
			expected: nil,
		},
		"equal": {
			original: Set[string]{"one"},
			new:      Set[string]{"one"},
			expected: nil,
		},
		"difference": {
			original: Set[string]{"one", "two", "four"},
			new:      Set[string]{"one", "two", "three"},
			expected: Set[string]{"four"},
		},
		"difference_remove": {
			original: Set[string]{"one", "two"},
			new:      Set[string]{"one"},
			expected: Set[string]{"two"},
		},
		"difference_add": {
			original: Set[string]{"one"},
			new:      Set[string]{"one", "two"},
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.original.Difference(test.new)
			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
