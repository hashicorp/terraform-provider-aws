// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator

import lazystringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator/lazy"

var (
	ContainsOnlyLowerCaseLettersNumbersHyphens     = lazystringvalidator.RegexMatches(`^[0-9a-z-]+$`, "must contain only lowercase letters, numbers, or hyphens")
	ContainsOnlyLowerCaseLettersNumbersUnderscores = lazystringvalidator.RegexMatches(`^[0-9a-z_]+$`, "must contain only lowercase letters, numbers, or underscores")

	StartsWithLetterOrNumber = lazystringvalidator.RegexMatches(`^[0-9a-z]`, "must start with a letter or number")
	EndsWithLetterOrNumber   = lazystringvalidator.RegexMatches(`[0-9a-z]$`, "must end with a letter or number")
)
