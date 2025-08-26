// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"fmt"

	"github.com/YakDriver/regexache"
)

func validReplicationGroupAuthToken(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if (len(value) < 16) || (len(value) > 128) {
		errors = append(errors, fmt.Errorf(
			"%q must contain from 16 to 128 alphanumeric characters or symbols (excluding @, \", and /)", k))
	}
	if !regexache.MustCompile(`^[^@"\/]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters or symbols (excluding @, \", and /) allowed in %q", k))
	}
	return
}
