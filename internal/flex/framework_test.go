package flex

import (
	"context"
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
			input: types.Set{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
			expected: []*string{aws.String("GET"), aws.String("HEAD")},
		},
		"zero elements": {
			input:    types.Set{ElemType: types.StringType, Elems: []attr.Value{}},
			expected: []*string{},
		},
		"invalid element type": {
			input: types.Set{ElemType: types.Int64Type, Elems: []attr.Value{
				types.Int64{Value: 42},
			}},
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
			input: types.Set{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
			expected: []string{"GET", "HEAD"},
		},
		"zero elements": {
			input:    types.Set{ElemType: types.StringType, Elems: []attr.Value{}},
			expected: []string{},
		},
		"invalid element type": {
			input: types.Set{ElemType: types.Int64Type, Elems: []attr.Value{
				types.Int64{Value: 42},
			}},
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
			input: types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{
				"one": types.String{Value: "GET"},
				"two": types.String{Value: "HEAD"},
			}},
			expected: map[string]string{
				"one": "GET",
				"two": "HEAD",
			},
		},
		"zero elements": {
			input:    types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{}},
			expected: map[string]string{},
		},
		"invalid element type": {
			input: types.Map{ElemType: types.BoolType, Elems: map[string]attr.Value{
				"one": types.Bool{Value: true},
				"two": types.Bool{Value: false},
			}},
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
			expected: types.List{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
		},
		"zero elements": {
			input:    []*string{},
			expected: types.List{ElemType: types.StringType, Elems: []attr.Value{}},
		},
		"nil array": {
			input:    nil,
			expected: types.List{ElemType: types.StringType, Elems: []attr.Value{}},
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
			expected: types.List{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
		},
		"zero elements": {
			input:    []string{},
			expected: types.List{ElemType: types.StringType, Elems: []attr.Value{}},
		},
		"nil array": {
			input:    nil,
			expected: types.List{ElemType: types.StringType, Elems: []attr.Value{}},
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
			expected: types.Set{ElemType: types.StringType, Elems: []attr.Value{
				types.String{Value: "GET"},
				types.String{Value: "HEAD"},
			}},
		},
		"zero elements": {
			input:    []string{},
			expected: types.Set{ElemType: types.StringType, Elems: []attr.Value{}},
		},
		"nil array": {
			input:    nil,
			expected: types.Set{ElemType: types.StringType, Elems: []attr.Value{}},
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
			expected: types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{
				"one": types.String{Value: "GET"},
				"two": types.String{Value: "HEAD"},
			}},
		},
		"zero elements": {
			input:    map[string]string{},
			expected: types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{}},
		},
		"nil map": {
			input:    nil,
			expected: types.Map{ElemType: types.StringType, Elems: map[string]attr.Value{}},
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
