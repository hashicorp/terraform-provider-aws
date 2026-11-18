// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"regexp"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/otherpkg"
	lazyvalidation "github.com/hashicorp/terraform-provider-aws/internal/sdkv2/validation/lazy"
)

func stringMatchDiag(pattern string) {
	// ruleid: use-lazyvalidation-string-match-diag
	lazyvalidation.StringMatchDiag(pattern, "must match pattern")

	// ruleid: use-lazyvalidation-string-match-diag
	lazyvalidation.StringMatchDiag(`^[a-z]+$`, "must contain lowercase letters")

	// ruleid: use-lazyvalidation-string-match
	lazyvalidation.StringMatch(pattern, "must match pattern")

	// ruleid: use-lazyvalidation-string-match
	lazyvalidation.StringMatch(`^[0-9]+$`, "must contain digits")

	// ok: use-lazyvalidation-string-match-diag
	validation.ToDiagFunc(validation.StringLenBetween(1, 128))

	// ok: use-lazyvalidation-string-match-diag
	lazyvalidation.StringMatchDiag(pattern, "must match pattern")

	// ok: use-lazyvalidation-string-match
	lazyvalidation.StringMatch(pattern, "must match pattern")
}

func stringMatchFunctionRegex() {
	// ruleid: use-lazyvalidation-string-match-precompiled-regex
	validation.StringMatch(patternRegex(), "must match pattern")

	// ruleid: use-lazyvalidation-string-match-precompiled-regex
	validation.StringMatch(patternRegexWithMessage())
}

func stringDoesNotMatch(pattern string) {
	// ruleid: use-lazyvalidation-string-does-not-match-diag
	lazyvalidation.StringDoesNotMatchDiag(pattern, "must not match pattern")

	// ruleid: use-lazyvalidation-string-does-not-match-diag
	lazyvalidation.StringDoesNotMatchDiag(`^aws/`, "must not begin with aws/")

	// ruleid: use-lazyvalidation-string-does-not-match
	lazyvalidation.StringDoesNotMatch(pattern, "must not match pattern")

	// ruleid: use-lazyvalidation-string-does-not-match
	lazyvalidation.StringDoesNotMatch(`^aws/`, "must not begin with aws/")

	// ok: use-lazyvalidation-string-does-not-match-diag
	validation.ToDiagFunc(validation.StringLenBetween(1, 128))

	// ok: use-lazyvalidation-string-does-not-match-diag
	lazyvalidation.StringDoesNotMatchDiag(pattern, "must not match pattern")

	// ok: use-lazyvalidation-string-does-not-match
	lazyvalidation.StringDoesNotMatch(pattern, "must not match pattern")
}

func stringDoesNotMatchFunctionRegex() {
	// ruleid: use-lazyvalidation-string-does-not-match-precompiled-regex
	validation.StringDoesNotMatch(patternRegex(), "must not match pattern")

	// ruleid: use-lazyvalidation-string-does-not-match-precompiled-regex
	validation.StringDoesNotMatch(patternRegexWithMessage())
}

func patternRegex() *regexp.Regexp {
	return regexache.MustCompile(`^[a-z]+$`)
}

func patternRegexWithMessage() (*regexp.Regexp, string) {
	return regexache.MustCompile(`^[a-z]+$`), "must match pattern"
}

func stringMatchVariableRegex() {
	localRegex := regexache.MustCompile(`^[a-z]+$`)

	// ruleid: use-lazyvalidation-string-match-precompiled-regex
	validation.StringMatch(localRegex, "must match pattern")

	// ruleid: use-lazyvalidation-string-match-precompiled-regex
	validation.StringMatch(packageRegex, "")

	// ruleid: use-lazyvalidation-string-match-precompiled-regex
	validation.StringMatch(otherpkg.Regex, "must match pattern")

	// ruleid: use-lazyvalidation-string-match-precompiled-regex
	validation.ToDiagFunc(validation.StringMatch(packageRegex, "must match pattern"))

	// ok: use-lazyvalidation-string-match-precompiled-regex
	lazyvalidation.StringMatch(`^[a-z]+$`, "must match pattern")

	// ok: use-lazyvalidation-string-match-precompiled-regex
	validation.StringLenBetween(1, 128)
}

func stringDoesNotMatchVariableRegex() {
	localRegex := regexache.MustCompile(`^aws/`)

	// ruleid: use-lazyvalidation-string-does-not-match-precompiled-regex
	validation.StringDoesNotMatch(localRegex, "must not match pattern")

	// ruleid: use-lazyvalidation-string-does-not-match-precompiled-regex
	validation.StringDoesNotMatch(packageRegex, "")

	// ruleid: use-lazyvalidation-string-does-not-match-precompiled-regex
	validation.StringDoesNotMatch(otherpkg.Regex, "must not match pattern")

	// ok: use-lazyvalidation-string-does-not-match-precompiled-regex
	lazyvalidation.StringDoesNotMatch(`^aws/`, "must not match pattern")

	// ok: use-lazyvalidation-string-does-not-match-precompiled-regex
	validation.StringLenBetween(1, 128)
}

var packageRegex = regexache.MustCompile(`^[a-z]+$`)

func stringMatchStdlibRegex() {
	// ruleid: use-lazyvalidation-string-match-precompiled-regex
	validation.StringMatch(regexp.MustCompile(`^[a-z]+$`), "must match pattern")

	// A `regexache.MustCompile(...)` literal is reported by this rule, not by the one for other expressions.
	// ruleid: use-lazyvalidation-string-match
	lazyvalidation.StringMatch(`^[a-z]+$`, "must match pattern")

	// ruleid: use-lazyvalidation-string-does-not-match-precompiled-regex
	validation.StringDoesNotMatch(regexp.MustCompile(`^aws/`), "must not match pattern")

	// A `regexache.MustCompile(...)` literal is reported by this rule, not by the one for other expressions.
	// ruleid: use-lazyvalidation-string-does-not-match
	lazyvalidation.StringDoesNotMatch(`^aws/`, "must not match pattern")
}
