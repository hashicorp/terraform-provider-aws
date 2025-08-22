// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	directoryIDRegex = regexache.MustCompile(`^d-[0-9a-f]{10}$`)

	domain                = regexache.MustCompile(`^([0-9A-Za-z]+[\\.-])+([0-9A-Za-z])+$`)
	domainWithTrailingDot = regexache.MustCompile(`^([0-9A-Za-z]+[\\.-])+([0-9A-Za-z])+[.]?$`)
)

var directoryIDValidator validator.String = stringvalidator.RegexMatches(directoryIDRegex, "must be a valid Directory Service Directory ID")

var domainValidator = validation.StringMatch(domain, "must be a fully qualified domain name and cannot end with a trailing period")

var domainWithTrailingDotValidator validator.String = stringvalidator.RegexMatches(domainWithTrailingDot, "must be a fully qualified domain name and may end with a trailing period")

var trustPasswordValidator validator.String = stringvalidator.RegexMatches(regexache.MustCompile(`^(\p{L}|\p{Nd}|\p{P}| )+$`), "can contain upper- and lower-case letters, numbers, and punctuation characters")
