// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package stringvalidator

import (
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

var (
	ContainsOnlyLowerCaseLettersNumbersHyphens     = stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase letters, numbers, or hyphens")
	ContainsOnlyLowerCaseLettersNumbersUnderscores = stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-z_]+$`), "must contain only lowercase letters, numbers, or underscores")

	StartsWithLetterOrNumber = stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-z]`), "must start with a letter or number")
	EndsWithLetterOrNumber   = stringvalidator.RegexMatches(regexache.MustCompile(`[0-9a-z]$`), "must end with a letter or number")
)
