// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
)

func TestContainsOnlyLowerCaseLettersNumbersHyphens(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in          types.String
		expectError bool
	}

	testCases := map[string]testCase{
		"valid": {
			in:          types.StringValue("valid-string-123"),
			expectError: false,
		},
		"invalid-uppercase": {
			in:          types.StringValue("Invalid-String"),
			expectError: true,
		},
		"invalid-special-char": {
			in:          types.StringValue("invalid@string"),
			expectError: true,
		},
		"empty": {
			in:          types.StringValue(""),
			expectError: true,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				ConfigValue: test.in,
			}
			res := validator.StringResponse{}
			tfstringvalidator.ContainsOnlyLowerCaseLettersNumbersHyphens.ValidateString(t.Context(), req, &res)

			if !res.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if res.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", res.Diagnostics)
			}
		})
	}
}

func TestContainsOnlyLowerCaseLettersNumbersUnderscores(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in          types.String
		expectError bool
	}

	testCases := map[string]testCase{
		"valid": {
			in:          types.StringValue("valid_string_123"),
			expectError: false,
		},
		"invalid-uppercase": {
			in:          types.StringValue("Invalid_String"),
			expectError: true,
		},
		"invalid-special-char": {
			in:          types.StringValue("invalid-string"),
			expectError: true,
		},
		"empty": {
			in:          types.StringValue(""),
			expectError: true,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				ConfigValue: test.in,
			}
			res := validator.StringResponse{}
			tfstringvalidator.ContainsOnlyLowerCaseLettersNumbersUnderscores.ValidateString(t.Context(), req, &res)

			if !res.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if res.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", res.Diagnostics)
			}
		})
	}
}

func TestStartsWithLetterOrNumber(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in          types.String
		expectError bool
	}

	testCases := map[string]testCase{
		"valid-starts-with-letter": {
			in:          types.StringValue("valid123"),
			expectError: false,
		},
		"valid-starts-with-number": {
			in:          types.StringValue("123valid"),
			expectError: false,
		},
		"invalid-special-char": {
			in:          types.StringValue("-invalid"),
			expectError: true,
		},
		"empty": {
			in:          types.StringValue(""),
			expectError: true,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				ConfigValue: test.in,
			}
			res := validator.StringResponse{}
			tfstringvalidator.StartsWithLetterOrNumber.ValidateString(t.Context(), req, &res)

			if !res.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if res.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", res.Diagnostics)
			}
		})
	}
}

func TestEndsWithLetterOrNumber(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in          types.String
		expectError bool
	}

	testCases := map[string]testCase{
		"valid-ends-with-letter": {
			in:          types.StringValue("123valid"),
			expectError: false,
		},
		"valid-ends-with-number": {
			in:          types.StringValue("valid123"),
			expectError: false,
		},
		"invalid-special-char": {
			in:          types.StringValue("invalid-"),
			expectError: true,
		},
		"empty": {
			in:          types.StringValue(""),
			expectError: true,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				ConfigValue: test.in,
			}
			res := validator.StringResponse{}
			tfstringvalidator.EndsWithLetterOrNumber.ValidateString(t.Context(), req, &res)

			if !res.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if res.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", res.Diagnostics)
			}
		})
	}
}
