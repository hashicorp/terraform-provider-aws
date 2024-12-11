// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"fmt"

	"github.com/YakDriver/regexache"
)

func validLinuxFileMode(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-7]{4}$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only valid linux mode is allowed in %q", k))
	}
	return
}
