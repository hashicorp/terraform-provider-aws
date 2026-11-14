// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	lazyvalidation "github.com/hashicorp/terraform-provider-aws/internal/sdkv2/validation/lazy"
)

func stringMatchDiag(pattern string) {
	// ruleid: use-lazyvalidation-string-match-diag
	validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(pattern), "must match pattern"))

	// ruleid: use-lazyvalidation-string-match-diag
	validation.ToDiagFunc(validation.StringMatch(
		regexache.MustCompile(`^[a-z]+$`),
		"must contain lowercase letters",
	))

	// ruleid: use-lazyvalidation-string-match
	validation.StringMatch(regexache.MustCompile(pattern), "must match pattern")

	// ruleid: use-lazyvalidation-string-match
	validation.StringMatch(
		regexache.MustCompile(`^[0-9]+$`),
		"must contain digits",
	)

	// ok: use-lazyvalidation-string-match-diag
	validation.ToDiagFunc(validation.StringLenBetween(1, 128))

	// ok: use-lazyvalidation-string-match-diag
	lazyvalidation.StringMatchDiag(pattern, "must match pattern")

	// ok: use-lazyvalidation-string-match
	lazyvalidation.StringMatch(pattern, "must match pattern")
}

func stringMatchFunctionRegex() {
	// ruleid: use-lazyvalidation-string-match-function-regex
	validation.StringMatch(patternRegex(), "must match pattern")

	// ruleid: use-lazyvalidation-string-match-function-regex
	validation.StringMatch(patternRegexWithMessage())
}

func stringDoesNotMatch(pattern string) {
	// ruleid: use-lazyvalidation-string-does-not-match-diag
	validation.ToDiagFunc(validation.StringDoesNotMatch(regexache.MustCompile(pattern), "must not match pattern"))

	// ruleid: use-lazyvalidation-string-does-not-match-diag
	validation.ToDiagFunc(validation.StringDoesNotMatch(
		regexache.MustCompile(`^aws/`),
		"must not begin with aws/",
	))

	// ruleid: use-lazyvalidation-string-does-not-match
	validation.StringDoesNotMatch(regexache.MustCompile(pattern), "must not match pattern")

	// ruleid: use-lazyvalidation-string-does-not-match
	validation.StringDoesNotMatch(
		regexache.MustCompile(`^aws/`),
		"must not begin with aws/",
	)

	// ok: use-lazyvalidation-string-does-not-match-diag
	validation.ToDiagFunc(validation.StringLenBetween(1, 128))

	// ok: use-lazyvalidation-string-does-not-match-diag
	lazyvalidation.StringDoesNotMatchDiag(pattern, "must not match pattern")

	// ok: use-lazyvalidation-string-does-not-match
	lazyvalidation.StringDoesNotMatch(pattern, "must not match pattern")
}

func stringDoesNotMatchFunctionRegex() {
	// ruleid: use-lazyvalidation-string-does-not-match-function-regex
	validation.StringDoesNotMatch(patternRegex(), "must not match pattern")

	// ruleid: use-lazyvalidation-string-does-not-match-function-regex
	validation.StringDoesNotMatch(patternRegexWithMessage())
}

func patternRegex() *regexp.Regexp {
	return regexache.MustCompile(`^[a-z]+$`)
}

func patternRegexWithMessage() (*regexp.Regexp, string) {
	return regexache.MustCompile(`^[a-z]+$`), "must match pattern"
}
