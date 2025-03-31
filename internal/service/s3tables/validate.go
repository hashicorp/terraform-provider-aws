// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables

import (
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

var (
	stringMustContainLowerCaseLettersNumbersHypens      = stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase letters, numbers, or hyphens")
	stringMustContainLowerCaseLettersNumbersUnderscores = stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-z_]+$`), "must contain only lowercase letters, numbers, or underscores")

	stringMustStartWithLetterOrNumber = stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-z]`), "must start with a letter or number")
	stringMustEndWithLetterOrNumber   = stringvalidator.RegexMatches(regexache.MustCompile(`[0-9a-z]$`), "must end with a letter or number")
)
