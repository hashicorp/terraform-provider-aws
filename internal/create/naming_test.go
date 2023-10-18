// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package create

import (
	"testing"
)

func strPtr(str string) *string {
	return &str
}

func TestHasResourceUniqueIDPlusAdditionalSuffix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Input    string
		Expected bool
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: false,
		},
		{
			TestName: "incorrect suffix",
			Input:    "test-123",
			Expected: false,
		},
		{
			TestName: "missing additional suffix with numbers",
			Input:    "test-20060102150405000000000001",
			Expected: false,
		},
		{
			TestName: "correct suffix with numbers",
			Input:    "test-20060102150405000000000001suffix",
			Expected: true,
		},
		{
			TestName: "missing additional suffix with hex",
			Input:    "test-200601021504050000000000a1",
			Expected: false,
		},
		{
			TestName: "correct suffix with hex",
			Input:    "test-200601021504050000000000a1suffix",
			Expected: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got := hasResourceUniqueIDPlusAdditionalSuffix(testCase.Input, "suffix")

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}

func TestNamePrefixFromName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Input    string
		Expected *string
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: nil,
		},
		{
			TestName: "incorrect suffix",
			Input:    "test-123",
			Expected: nil,
		},
		{
			TestName: "prefix without hyphen, correct suffix",
			Input:    "test20060102150405000000000001",
			Expected: strPtr("test"),
		},
		{
			TestName: "prefix with hyphen, correct suffix",
			Input:    "test-20060102150405000000000001",
			Expected: strPtr("test-"),
		},
		{
			TestName: "prefix with hyphen, correct suffix with hex",
			Input:    "test-200601021504050000000000f1",
			Expected: strPtr("test-"),
		},
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17017
		{
			TestName: "terraform prefix, correct suffix",
			Input:    "terraform-20060102150405000000000001",
			Expected: strPtr("terraform-"),
		},
		{
			TestName: "KMS alias prefix",
			Input:    "alias/20210723150229087000000002",
			Expected: strPtr("alias/"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			expected := testCase.Expected
			got := NamePrefixFromName(testCase.Input)

			if expected == nil && got != nil {
				t.Errorf("got %s, expected nil", *got)
			}

			if expected != nil && got == nil {
				t.Errorf("got nil, expected %s", *expected)
			}

			if expected != nil && got != nil && *expected != *got {
				t.Errorf("got %s, expected %s", *got, *expected)
			}
		})
	}

	t.Run("extracting prefix from generated name", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < 10; i++ {
			prefix := "test-"
			input := Name("", prefix)
			got := NamePrefixFromName(input)

			if got == nil {
				t.Errorf("run%d: got nil, expected %s for input %s", i, prefix, input)
			}

			if got != nil && prefix != *got {
				t.Errorf("run%d: got %s, expected %s for input %s", i, *got, prefix, input)
			}
		}
	})
}

func TestNamePrefixFromNameWithSuffix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Input    string
		Expected *string
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: nil,
		},
		{
			TestName: "incorrect suffix",
			Input:    "test-123",
			Expected: nil,
		},
		{
			TestName: "prefix without hyphen, missing additional suffix",
			Input:    "test20060102150405000000000001",
			Expected: nil,
		},
		{
			TestName: "prefix without hyphen, correct suffix",
			Input:    "test20060102150405000000000001suffix",
			Expected: strPtr("test"),
		},
		{
			TestName: "prefix with hyphen, missing additional suffix",
			Input:    "test-20060102150405000000000001",
			Expected: nil,
		},
		{
			TestName: "prefix with hyphen, correct suffix",
			Input:    "test-20060102150405000000000001suffix",
			Expected: strPtr("test-"),
		},
		{
			TestName: "prefix with hyphen, missing additional suffix with hex",
			Input:    "test-200601021504050000000000f1",
			Expected: nil,
		},
		{
			TestName: "prefix with hyphen, correct suffix with hex",
			Input:    "test-200601021504050000000000f1suffix",
			Expected: strPtr("test-"),
		},
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17017
		{
			TestName: "terraform prefix, missing additional suffix",
			Input:    "terraform-20060102150405000000000001",
			Expected: nil,
		},
		{
			TestName: "terraform prefix, correct suffix",
			Input:    "terraform-20060102150405000000000001suffix",
			Expected: strPtr("terraform-"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			expected := testCase.Expected
			got := NamePrefixFromNameWithSuffix(testCase.Input, "suffix")

			if expected == nil && got != nil {
				t.Errorf("got %s, expected nil", *got)
			}

			if expected != nil && got == nil {
				t.Errorf("got nil, expected %s", *expected)
			}

			if expected != nil && got != nil && *expected != *got {
				t.Errorf("got %s, expected %s", *got, *expected)
			}
		})
	}

	t.Run("extracting prefix from generated name", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < 10; i++ {
			prefix := "test-"
			input := NameWithSuffix("", prefix, "suffix")
			got := NamePrefixFromNameWithSuffix(input, "suffix")

			if got == nil {
				t.Errorf("run%d: got nil, expected %s for input %s", i, prefix, input)
			}

			if got != nil && prefix != *got {
				t.Errorf("run%d: got %s, expected %s for input %s", i, *got, prefix, input)
			}
		}
	})
}
