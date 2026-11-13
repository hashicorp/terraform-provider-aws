// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lazyvalidation

import (
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
)

func TestValidationStringMatch(t *testing.T) {
	t.Parallel()

	runTestCases(t, []testCase{
		{
			val: "foobar",
			f:   StringMatch(".*foo.*", ""),
		},
		{
			val:         "bar",
			f:           StringMatch(".*foo.*", ""),
			expectedErr: regexache.MustCompile(`^expected value of test_property to match regular expression ` + regexp.QuoteMeta(`".*foo.*"`) + `, got "bar"$`),
		},
		{
			val:         "bar",
			f:           StringMatch(".*foo.*", "value must contain foo"),
			expectedErr: regexache.MustCompile(`^invalid value for test_property \(value must contain foo\)$`),
		},
		// Wrong input type
		{
			val:         42,
			f:           StringMatch(".*foo.*", ""),
			expectedErr: regexache.MustCompile(`^expected type of test_property to be string$`),
		},
		// Invalid regexp
		{
			val:         "bar",
			f:           StringMatch("(", ""),
			expectedErr: regexache.MustCompile(`^failed to compile regex pattern "\(": ` + regexp.QuoteMeta(func() string { _, err := regexp.Compile(`(`); return err.Error() }()) + `$`),
		},
	})
}

func TestValidationStringMatchDiag(t *testing.T) {
	t.Parallel()

	runDiagTestCases(t, []diagTestCase{
		{
			val: "foobar",
			f:   StringMatchDiag(".*foo.*", ""),
		},
		{
			val:                "bar",
			f:                  StringMatchDiag(".*foo.*", ""),
			expectedDiagDetail: regexache.MustCompile(`^Expected value to match regular expression ` + regexp.QuoteMeta(`".*foo.*"`) + `, got "bar"$`),
		},
		{
			val:                "bar",
			f:                  StringMatchDiag(".*foo.*", "value must contain foo"),
			expectedDiagDetail: regexache.MustCompile(`^Value is invalid: value must contain foo$`),
		},
		// Wrong input type
		{
			val:                42,
			f:                  StringMatchDiag(".*foo.*", ""),
			expectedDiagDetail: regexache.MustCompile(`^Expected type to be string$`),
		},
		// Invalid regexp
		{
			val:                "bar",
			f:                  StringMatchDiag("(", ""),
			expectedDiagDetail: regexache.MustCompile(`(?s)An unexpected error occurred while validating the attribute value\..+The regex pattern "\(" failed to compile: ` + regexp.QuoteMeta(func() string { _, err := regexp.Compile(`(`); return err.Error() }())),
		},
	})
}

func TestValidationStringDoesNotMatch(t *testing.T) {
	t.Parallel()

	runTestCases(t, []testCase{
		{
			val: "foobar",
			f:   StringDoesNotMatch(".*baz.*", ""),
		},
		{
			val:         "bar",
			f:           StringDoesNotMatch(".*bar.*", ""),
			expectedErr: regexache.MustCompile(`^expected value of test_property to not match regular expression ` + regexp.QuoteMeta(`".*bar.*"`) + `, got "bar"$`),
		},
		{
			val:         "bar",
			f:           StringDoesNotMatch(".*bar.*", "value must not contain foo"),
			expectedErr: regexache.MustCompile(`^invalid value for test_property \(value must not contain foo\)$`),
		},
		// Wrong input type
		{
			val:         42,
			f:           StringDoesNotMatch(".*bar.*", ""),
			expectedErr: regexache.MustCompile(`^expected type of test_property to be string$`),
		},
		// Invalid regexp
		{
			val:         "bar",
			f:           StringDoesNotMatch("(", ""),
			expectedErr: regexache.MustCompile(`^failed to compile regex pattern "\(": ` + regexp.QuoteMeta(func() string { _, err := regexp.Compile(`(`); return err.Error() }()) + `$`),
		},
	})
}

func TestValidationStringDoesNotMatchDiag(t *testing.T) {
	t.Parallel()

	runDiagTestCases(t, []diagTestCase{
		{
			val: "foobar",
			f:   StringDoesNotMatchDiag(".*baz.*", ""),
		},
		{
			val:                "bar",
			f:                  StringDoesNotMatchDiag(".*bar.*", ""),
			expectedDiagDetail: regexache.MustCompile(`^Expected value to not match regular expression ` + regexp.QuoteMeta(`".*bar.*"`) + `, got "bar"$`),
		},
		{
			val:                "bar",
			f:                  StringDoesNotMatchDiag(".*bar.*", "value must not contain foo"),
			expectedDiagDetail: regexache.MustCompile(`^Value is invalid: value must not contain foo$`),
		},
		// Wrong input type
		{
			val:                42,
			f:                  StringDoesNotMatchDiag(".*bar.*", ""),
			expectedDiagDetail: regexache.MustCompile(`^Expected type to be string$`),
		},
		// Invalid regexp
		{
			val:                "bar",
			f:                  StringDoesNotMatchDiag("(", ""),
			expectedDiagDetail: regexache.MustCompile(`(?s)An unexpected error occurred while validating the attribute value\..+The regex pattern "\(" failed to compile: ` + regexp.QuoteMeta(func() string { _, err := regexp.Compile(`(`); return err.Error() }())),
		},
	})
}

func TestRegexacheCompile(t *testing.T) {
	t.Parallel()

	t.Run("valid pattern", func(t *testing.T) {
		t.Parallel()

		re, err := regexacheCompile(`^foo[0-9]+$`)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if re == nil {
			t.Fatal("expected non-nil regexp")
		}
		if !re.MatchString("foo123") {
			t.Errorf("expected %q to match foo123", re)
		}
		if re.MatchString("bar") {
			t.Errorf("expected %q not to match bar", re)
		}
	})

	t.Run("invalid pattern returns stdlib error", func(t *testing.T) {
		t.Parallel()

		const pattern = `(`

		re, err := regexacheCompile(pattern)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if re != nil {
			t.Errorf("expected nil regexp, got %q", re)
		}

		_, wantErr := regexp.Compile(pattern)
		if err.Error() != wantErr.Error() {
			t.Errorf("expected error %q, got %q", wantErr, err)
		}
	})

	t.Run("repeated calls return same compiled regexp", func(t *testing.T) {
		t.Parallel()

		if !regexache.IsCachingEnabled() {
			t.Skip("regexache caching is disabled")
		}

		const pattern = `^[a-z]+$`

		first, err := regexacheCompile(pattern)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		second, err := regexacheCompile(pattern)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if first != second {
			t.Errorf("expected the same *regexp.Regexp for repeated calls, got %p and %p", first, second)
		}
		if !first.MatchString("abc") {
			t.Errorf("expected %q to match abc", first)
		}
	})
}
