// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	lazystringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator/lazy"
)

func regexMatches(pattern string) {
	// ruleid: use-lazystringvalidator-regex-matches
	lazystringvalidator.RegexMatches(pattern, "must match pattern")

	// ruleid: use-lazystringvalidator-regex-matches
	lazystringvalidator.RegexMatches(`^[a-z]+$`, "must contain lowercase letters")

	// ok: use-lazystringvalidator-regex-matches
	lazystringvalidator.RegexMatches(pattern, "must match pattern")
}
