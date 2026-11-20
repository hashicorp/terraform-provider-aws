// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	lazystringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator/lazy"
	lazyvalidation "github.com/hashicorp/terraform-provider-aws/internal/sdkv2/validation/lazy"
)

var directoryIDValidator validator.String = lazystringvalidator.RegexMatches(`^d-[0-9a-f]{10}$`, "must be a valid Directory Service Directory ID")

var domainValidator = lazyvalidation.StringMatch(`^([0-9A-Za-z]+[\\.-])+([0-9A-Za-z])+$`, "must be a fully qualified domain name and cannot end with a trailing period")

var domainWithTrailingDotValidator validator.String = lazystringvalidator.RegexMatches(`^([0-9A-Za-z]+[\\.-])+([0-9A-Za-z])+[.]?$`, "must be a fully qualified domain name and may end with a trailing period")

var trustPasswordValidator validator.String = lazystringvalidator.RegexMatches(`^(\p{L}|\p{Nd}|\p{P}| )+$`, "can contain upper- and lower-case letters, numbers, and punctuation characters")
