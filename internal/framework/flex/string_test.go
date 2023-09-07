// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
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
			input:    fwtypes.ARNValue(errs.Must(arn.Parse("arn:aws:iam::123456789012:user/David"))),
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

			got := flex.ARNStringFromFramework(context.Background(), test.input)

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
			expected: fwtypes.ARNValue(errs.Must(arn.Parse("arn:aws:iam::123456789012:user/David"))),
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

			var diags diag.Diagnostics
			got := flex.StringToFrameworkARN(context.Background(), test.input, &diags)

			if err := fwdiag.DiagnosticsError(diags); err != nil {
				t.Errorf("err %q", err)
			} else if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
