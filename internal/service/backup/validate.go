// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"fmt"

	"github.com/YakDriver/regexache"
)

func validReportPlanName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[A-Za-z]{1}[0-9A-Za-z_]{0,255}$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q (%q) must be must be between 1 and 256 characters, starting with a letter, and consisting of letters, numbers, and underscores.", k, v))
	}
	return
}

// The pattern for framework and report plan name is the same but separate functions are used in the event that there are future differences
func validFrameworkName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[A-Za-z]{1}[0-9A-Za-z_]{0,255}$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q (%q) must be must be between 1 and 256 characters, starting with a letter, and consisting of letters, numbers, and underscores.", k, v))
	}
	return
}
