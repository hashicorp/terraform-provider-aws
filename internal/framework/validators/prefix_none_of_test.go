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

func TestPrefixNoneOfValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in                 types.String
		prefixNoneOfValues []string
		expectError        bool
	}

	testCases := map[string]testCase{
		"simple-match": {
			in: types.StringValue("prefix"),
			prefixNoneOfValues: []string{
				"pre",
				"first",
				"1st",
			},
			expectError: true,
		},
		"simple-mismatch-case-insensitive": {
			in: types.StringValue("prefix"),
			prefixNoneOfValues: []string{
				"PRE",
				"first",
				"1st",
			},
		},
		"simple-mismatch": {
			in: types.StringValue("prefix"),
			prefixNoneOfValues: []string{
				"pri",
				"first",
				"1st",
			},
		},
		"skip-validation-on-null": {
			in: types.StringNull(),
			prefixNoneOfValues: []string{
				"pre",
				"first",
				"1st",
			},
		},
		"skip-validation-on-unknown": {
			in: types.StringUnknown(),
			prefixNoneOfValues: []string{
				"pre",
				"first",
				"1st",
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
			validators.PrefixNoneOf(test.prefixNoneOfValues...).ValidateString(context.TODO(), req, &res)

			if !res.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if res.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", res.Diagnostics)
			}
		})
	}
}

func TestPrefixNoneOfValidator_Description(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in       []string
		expected string
	}

	testCases := map[string]testCase{
		"quoted-once": {
			in:       []string{"foo", "bar", "baz"},
			expected: `value must begin with none of: ["foo" "bar" "baz"]`,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := validators.PrefixNoneOf(test.in...)

			got := v.MarkdownDescription(context.Background())

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
