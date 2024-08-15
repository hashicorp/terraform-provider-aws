// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

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

			got := flex.StringFromFramework(context.Background(), test.input)

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

			got := flex.StringToFramework(context.Background(), test.input)

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

			got := flex.StringToFrameworkLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStringValueToFramework(t *testing.T) {
	t.Parallel()

	// AWS enums use custom types with an underlying string type.
	type custom string
	const (
		test custom = "TEST"
	)

	type testCase struct {
		input    custom
		expected types.String
	}
	tests := map[string]testCase{
		"valid": {
			input:    test,
			expected: types.StringValue("TEST"),
		},
		"empty": {
			input:    custom(""),
			expected: types.StringNull(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.StringValueToFramework(context.Background(), test.input)

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

			got := flex.StringValueToFrameworkLegacy(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestARNStringFromFramework(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    fwtypes.ARN
		expected *string
	}
	tests := map[string]testCase{
		"valid ARN": {
			input:    fwtypes.ARNValue("arn:aws:iam::123456789012:user/David"),
			expected: aws.String("arn:aws:iam::123456789012:user/David"),
		},
		"null ARN": {
			input:    fwtypes.ARNNull(),
			expected: nil,
		},
		"unknown ARN": {
			input:    fwtypes.ARNUnknown(),
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.StringFromFramework(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStringToFrameworkARN(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    *string
		expected fwtypes.ARN
	}
	tests := map[string]testCase{
		"valid ARN": {
			input:    aws.String("arn:aws:iam::123456789012:user/David"),
			expected: fwtypes.ARNValue("arn:aws:iam::123456789012:user/David"),
		},
		"null ARN": {
			input:    nil,
			expected: fwtypes.ARNNull(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.StringToFrameworkARN(context.Background(), test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestEmptyStringAsNull(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    types.String
		expected types.String
	}
	tests := map[string]testCase{
		"valid": {
			input:    types.StringValue("TEST"),
			expected: types.StringValue("TEST"),
		},
		"empty": {
			input:    types.StringValue(""),
			expected: types.StringNull(),
		},
		"null": {
			input:    types.StringNull(),
			expected: types.StringNull(),
		},
		"unknown": {
			input:    types.StringUnknown(),
			expected: types.StringUnknown(),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := flex.EmptyStringAsNull(test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
