// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"fmt"
	"regexp"
	"strings"
)

func validSecurityGroupRuleDescription(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 255 characters: %q", k, value))
	}

	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_IpRange.html. Note that
	// "" is an allowable description value.
	pattern := `^[A-Za-z0-9 \.\_\-\:\/\(\)\#\,\@\[\]\+\=\&\;\{\}\!\$\*]*$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't comply with restrictions (%q): %q",
			k, pattern, value))
	}
	return
}

// validNestedExactlyOneOf is called on the map representing a nested schema element
// Once ExactlyOneOf is supported for nested elements, this should be deprecated.
func validNestedExactlyOneOf(m map[string]interface{}, valid []string) error {
	specified := make([]string, 0)
	for _, k := range valid {
		if v, ok := m[k].(string); ok && v != "" {
			specified = append(specified, k)
		}
	}

	if len(specified) == 0 {
		return fmt.Errorf("one of `%s` must be specified", strings.Join(valid, ", "))
	}
	if len(specified) > 1 {
		return fmt.Errorf("only one of `%s` can be specified, but `%s` were specified.", strings.Join(valid, ", "), strings.Join(specified, ", "))
	}
	return nil
}
