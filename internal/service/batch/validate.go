// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"fmt"

	"github.com/YakDriver/regexache"
)

func validName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9A-Za-z]{1}[0-9A-Za-z_-]{0,127}$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q (%q) must be up to 128 letters (uppercase and lowercase), numbers, underscores and dashes, and must start with an alphanumeric.", k, v))
	}
	return
}

func validPrefix(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9A-Za-z]{1}[0-9A-Za-z_-]{0,101}$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q (%q) must be up to 102 letters (uppercase and lowercase), numbers, underscores and dashes, and must start with an alphanumeric.", k, v))
	}
	return
}

func validShareIdentifier(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9A-Za-z]{0,254}[0-9A-Za-z*]?$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q (%q) must be limited to 255 alphanumeric characters, where the last character can be an asterisk (*).", k, v))
	}
	return
}
