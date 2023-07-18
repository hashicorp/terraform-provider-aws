// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"fmt"
	"regexp"
)

func validateS3MultiRegionAccessPointName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) < 3 || len(value) > 50 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be less than 3 or longer than 50 characters", k))
	}
	if regexp.MustCompile(`_|[A-Z]|\.`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"cannot contain underscores, uppercase letters, or periods. %q", k))
	}
	if regexp.MustCompile(`^-`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot begin with a hyphen", k))
	}
	if regexp.MustCompile(`-$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot end with a hyphen", k))
	}
	return
}
