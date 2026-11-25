// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"regexp"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	lazystringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator/lazy"
	"github.com/hashicorp/terraform-provider-aws/internal/otherpkg"
)

func regexMatches(pattern string) {
	// ruleid: use-lazystringvalidator-regex-matches
	stringvalidator.RegexMatches(regexache.MustCompile(pattern), "must match pattern")

	// ruleid: use-lazystringvalidator-regex-matches
	stringvalidator.RegexMatches(
		regexache.MustCompile(`^[a-z]+$`),
		"must contain lowercase letters",
	)

	// ok: use-lazystringvalidator-regex-matches
	lazystringvalidator.RegexMatches(pattern, "must match pattern")
}

func regexMatchesFunctionRegex() {
	// ruleid: use-lazystringvalidator-regex-matches-precompiled-regex
	stringvalidator.RegexMatches(patternRegex(), "must match pattern")

	// ruleid: use-lazystringvalidator-regex-matches-precompiled-regex
	stringvalidator.RegexMatches(patternRegex(), "")

	// ok: use-lazystringvalidator-regex-matches-precompiled-regex
	lazystringvalidator.RegexMatches(`^[a-z]+$`, "must match pattern")
}

func regexMatchesVariableRegex() {
	localRegex := regexache.MustCompile(`^[a-z]+$`)

	// ruleid: use-lazystringvalidator-regex-matches-precompiled-regex
	stringvalidator.RegexMatches(localRegex, "must match pattern")

	// ruleid: use-lazystringvalidator-regex-matches-precompiled-regex
	stringvalidator.RegexMatches(packageRegex, "")

	// ruleid: use-lazystringvalidator-regex-matches-precompiled-regex
	stringvalidator.RegexMatches(otherpkg.Regex, "must match pattern")

	// ok: use-lazystringvalidator-regex-matches-precompiled-regex
	lazystringvalidator.RegexMatches(`^[a-z]+$`, "must match pattern")

	// ok: use-lazystringvalidator-regex-matches-precompiled-regex
	stringvalidator.LengthBetween(1, 64)
}

var packageRegex = regexache.MustCompile(`^[a-z]+$`)

func patternRegex() *regexp.Regexp {
	return regexache.MustCompile(`^[a-z]+$`)
}

func regexMatchesStdlibRegex() {
	// ruleid: use-lazystringvalidator-regex-matches-precompiled-regex
	stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z]+$`), "must match pattern")

	// A `regexache.MustCompile(...)` literal is reported by this rule, not by the one for other expressions.
	// ruleid: use-lazystringvalidator-regex-matches
	stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z]+$`), "must match pattern")
}
