// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

import (
	"fmt"

	"github.com/YakDriver/regexache"
)

func validResolverName(v any, k string) (ws []string, errors []error) {
	// Type: String
	// Length Constraints: Maximum length of 64.
	// Pattern: (?!^[0-9]+$)([0-9A-Za-z-_' ']+)
	value := v.(string)

	// re2 doesn't support negative lookaheads so check for single numeric character explicitly.
	if regexache.MustCompile(`^[0-9]$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot be a single digit", k))
	}
	if !regexache.MustCompile(`^[0-9A-Za-z_' '-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters, '-', '_' and ' ' are allowed in %q", k))
	}
	if len(value) > 64 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than 64 characters", k))
	}

	return
}
