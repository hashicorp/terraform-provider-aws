// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"fmt"

	"github.com/YakDriver/regexache"
)

func validateS3MultiRegionAccessPointName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) < 3 || len(value) > 50 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be less than 3 or longer than 50 characters", k))
	}
	if regexache.MustCompile(`_|[A-Z]|\.`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"cannot contain underscores, uppercase letters, or periods. %q", k))
	}
	if regexache.MustCompile(`^-`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot begin with a hyphen", k))
	}
	if regexache.MustCompile(`-$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot end with a hyphen", k))
	}
	return
}
