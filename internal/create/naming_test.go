// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package create

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
)

func strPtr(str string) *string {
	return &str
}

func nameWithSuffix(name string, namePrefix string, nameSuffix string) string {
	return NewNameGenerator(WithConfiguredName(name), WithConfiguredPrefix(namePrefix), WithSuffix(nameSuffix)).Generate()
}

func TestName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName         string
		configuredName   string
		configuredPrefix string
		expectedRegexp   *regexp.Regexp
	}{
		{
			testName:       "no configured name or prefix",
			expectedRegexp: regexache.MustCompile(fmt.Sprintf("^terraform-[[:xdigit:]]{%d}$", id.UniqueIDSuffixLength)),
		},
		{
			testName:       "configured name only",
			configuredName: "testing",
			expectedRegexp: regexache.MustCompile(`^testing$`),
		},
		{
			testName:         "configured prefix only",
			configuredPrefix: "pfx-",
			expectedRegexp:   regexache.MustCompile(fmt.Sprintf("^pfx-[[:xdigit:]]{%d}$", id.UniqueIDSuffixLength)),
		},
		{
			testName:         "configured name and prefix",
			configuredName:   "testing",
			configuredPrefix: "pfx-",
			expectedRegexp:   regexache.MustCompile(`^testing$`),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			got := Name(testCase.configuredName, testCase.configuredPrefix)

			if !testCase.expectedRegexp.MatchString(got) {
				t.Errorf("Name(%q, %q) = %v, does not match %s", testCase.configuredName, testCase.configuredPrefix, got, testCase.expectedRegexp)
			}
		})
	}
}

func TestNameWithSuffix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName         string
		configuredName   string
		configuredPrefix string
		suffix           string
		expectedRegexp   *regexp.Regexp
	}{
		{
			testName:       "no configured name or prefix, no suffix",
			expectedRegexp: regexache.MustCompile(fmt.Sprintf("^terraform-[[:xdigit:]]{%d}$", id.UniqueIDSuffixLength)),
		},
		{
			testName:       "configured name only, no suffix",
			configuredName: "testing",
			expectedRegexp: regexache.MustCompile(`^testing$`),
		},
		{
			testName:         "configured prefix only, no suffix",
			configuredPrefix: "pfx-",
			expectedRegexp:   regexache.MustCompile(fmt.Sprintf("^pfx-[[:xdigit:]]{%d}$", id.UniqueIDSuffixLength)),
		},
		{
			testName:         "configured name and prefix, no suffix",
			configuredName:   "testing",
			configuredPrefix: "pfx-",
			expectedRegexp:   regexache.MustCompile(`^testing$`),
		},
		{
			testName:       "no configured name or prefix, with suffix",
			expectedRegexp: regexache.MustCompile(fmt.Sprintf("^terraform-[[:xdigit:]]{%d}-sfx$", id.UniqueIDSuffixLength)),
			suffix:         "-sfx",
		},
		{
			testName:       "configured name only, with suffix",
			configuredName: "testing",
			expectedRegexp: regexache.MustCompile(`^testing$`),
			suffix:         "-sfx",
		},
		{
			testName:         "configured prefix only, with suffix",
			configuredPrefix: "pfx-",
			expectedRegexp:   regexache.MustCompile(fmt.Sprintf("^pfx-[[:xdigit:]]{%d}-sfx$", id.UniqueIDSuffixLength)),
			suffix:           "-sfx",
		},
		{
			testName:         "configured name and prefix, with suffix",
			configuredName:   "testing",
			configuredPrefix: "pfx-",
			expectedRegexp:   regexache.MustCompile(`^testing$`),
			suffix:           "-sfx",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			got := nameWithSuffix(testCase.configuredName, testCase.configuredPrefix, testCase.suffix)

			if !testCase.expectedRegexp.MatchString(got) {
				t.Errorf("NameWithSuffix(%q, %q, %q) = %v, does not match %s", testCase.configuredName, testCase.configuredPrefix, testCase.suffix, got, testCase.expectedRegexp)
			}
		})
	}
}

func TestNameWithDefaultPrefix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName         string
		configuredName   string
		configuredPrefix string
		expectedRegexp   *regexp.Regexp
	}{
		{
			testName:       "no configured name or prefix",
			expectedRegexp: regexache.MustCompile(fmt.Sprintf("^def-[[:xdigit:]]{%d}$", id.UniqueIDSuffixLength)),
		},
		{
			testName:       "configured name only",
			configuredName: "testing",
			expectedRegexp: regexache.MustCompile(`^testing$`),
		},
		{
			testName:         "configured prefix only",
			configuredPrefix: "pfx-",
			expectedRegexp:   regexache.MustCompile(fmt.Sprintf("^pfx-[[:xdigit:]]{%d}$", id.UniqueIDSuffixLength)),
		},
		{
			testName:         "configured name and prefix",
			configuredName:   "testing",
			configuredPrefix: "pfx-",
			expectedRegexp:   regexache.MustCompile(`^testing$`),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			got := NewNameGenerator(WithConfiguredName(testCase.configuredName), WithConfiguredPrefix(testCase.configuredPrefix), WithDefaultPrefix("def-")).Generate()

			if !testCase.expectedRegexp.MatchString(got) {
				t.Errorf("NameWithDefaultPrefix(%q, %q) = %v, does not match %s", testCase.configuredName, testCase.configuredPrefix, got, testCase.expectedRegexp)
			}
		})
	}
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

		for i := range 10 {
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

		for i := range 10 {
			prefix := "test-"
			input := nameWithSuffix("", prefix, "suffix")
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
