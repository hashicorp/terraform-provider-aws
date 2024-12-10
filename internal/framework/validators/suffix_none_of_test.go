// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

func TestSuffixNoneOfValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in                 types.String
		suffixNoneOfValues []string
		expectError        bool
	}

	testCases := map[string]testCase{
		"simple-match": {
			in: types.StringValue("suffix"),
			suffixNoneOfValues: []string{
				"fix",
				"last",
			},
			expectError: true,
		},
		"simple-mismatch-case-insensitive": {
			in: types.StringValue("suffix"),
			suffixNoneOfValues: []string{
				"FIX",
				"last",
			},
		},
		"simple-mismatch": {
			in: types.StringValue("suffix"),
			suffixNoneOfValues: []string{
				"fax",
				"last",
			},
		},
		"skip-validation-on-null": {
			in: types.StringNull(),
			suffixNoneOfValues: []string{
				"fix",
				"last",
			},
		},
		"skip-validation-on-unknown": {
			in: types.StringUnknown(),
			suffixNoneOfValues: []string{
				"fix",
				"last",
			},
		},
	}

	for name, test := range testCases {
		t.Run(fmt.Sprintf("ValidateString - %s", name), func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{
				ConfigValue: test.in,
			}
			res := validator.StringResponse{}
			validators.SuffixNoneOf(test.suffixNoneOfValues...).ValidateString(context.TODO(), req, &res)

			if !res.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if res.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", res.Diagnostics)
			}
		})
	}
}

func TestSuffixNoneOfValidator_Description(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in       []string
		expected string
	}

	testCases := map[string]testCase{
		"quoted-once": {
			in:       []string{"foo", "bar", "baz"},
			expected: `value must end with none of: ["foo" "bar" "baz"]`,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := validators.SuffixNoneOf(test.in...)

			got := v.MarkdownDescription(context.Background())

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
