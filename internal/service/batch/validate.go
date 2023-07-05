// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"fmt"
	"regexp"
)

func validName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9a-zA-Z]{1}[0-9a-zA-Z_\-]{0,127}$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q (%q) must be up to 128 letters (uppercase and lowercase), numbers, underscores and dashes, and must start with an alphanumeric.", k, v))
	}
	return
}

func validPrefix(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9a-zA-Z]{1}[0-9a-zA-Z_\-]{0,101}$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q (%q) must be up to 102 letters (uppercase and lowercase), numbers, underscores and dashes, and must start with an alphanumeric.", k, v))
	}
	return
}

func validShareIdentifier(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[a-zA-Z0-9]{0,254}[a-zA-Z0-9*]?$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q (%q) must be limited to 255 alphanumeric characters, where the last character can be an asterisk (*).", k, v))
	}
	return
}
