// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
)

func validEventSubscriptionName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[A-Za-z][0-9A-Za-z-]*$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only ASCII letters, digits, and hyphens allowed in %q, and it must begin with a letter", k))
	}
	// When using name prefix, the generated suffix will always end with a number, so this check is not relevant
	if strings.HasSuffix(value, "-") {
		errors = append(errors, fmt.Errorf(
			"%q cannot end with a hyphen", k))
	}
	if strings.Contains(value, "--") {
		errors = append(errors, fmt.Errorf(
			"%q cannot contain two consecutive hyphens", k))
	}
	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 255 characters", k))
	}
	return
}

func validEventSubscriptionNamePrefix(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[A-Za-z][0-9A-Za-z-]*$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only ASCII letters, digits, and hyphens allowed in %q, and it must begin with a letter", k))
	}
	if strings.Contains(value, "--") {
		errors = append(errors, fmt.Errorf(
			"%q cannot contain two consecutive hyphens", k))
	}
	// When using name prefix, the limit must account for the generated suffix that is 26 characters long
	if len(value) > 229 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 229 characters", k))
	}
	return
}

func validOptionGroupName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[a-z]`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"first character of %q must be a letter", k))
	}
	if !regexache.MustCompile(`^[0-9a-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters and hyphens allowed in %q", k))
	}
	if regexache.MustCompile(`--`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot contain two consecutive hyphens", k))
	}
	if regexache.MustCompile(`-$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot end with a hyphen", k))
	}
	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than 255 characters", k))
	}
	return
}

func validOptionGroupNamePrefix(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[a-z]`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"first character of %q must be a letter", k))
	}
	if !regexache.MustCompile(`^[0-9a-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and hyphens allowed in %q", k))
	}
	if regexache.MustCompile(`--`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot contain two consecutive hyphens", k))
	}
	if len(value) > 229 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than 229 characters", k))
	}
	return
}

func validParamGroupName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9a-z.-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters, periods, and hyphens allowed in parameter group %q", k))
	}
	if !regexache.MustCompile(`^[a-z]`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"first character of parameter group %q must be a letter", k))
	}
	if regexache.MustCompile(`--`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"parameter group %q cannot contain two consecutive hyphens", k))
	}
	if regexache.MustCompile(`-$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"parameter group %q cannot end with a hyphen", k))
	}
	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"parameter group %q cannot be greater than 255 characters", k))
	}
	return
}

func validParamGroupNamePrefix(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9a-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters and hyphens allowed in parameter group %q", k))
	}
	if !regexache.MustCompile(`^[a-z]`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"first character of parameter group %q must be a letter", k))
	}
	if regexache.MustCompile(`--`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"parameter group %q cannot contain two consecutive hyphens", k))
	}
	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"parameter group %q cannot be greater than 226 characters", k))
	}
	return
}

func validSubnetGroupName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9a-z_ .-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters, hyphens, underscores, periods, and spaces allowed in %q", k))
	}
	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 255 characters", k))
	}
	if regexache.MustCompile(`(?i)^default$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q is not allowed as %q", "Default", k))
	}
	return
}

func validSubnetGroupNamePrefix(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9a-z_ .-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters, hyphens, underscores, periods, and spaces allowed in %q", k))
	}
	if len(value) > 229 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 229 characters", k))
	}
	return
}

func validIdentifier(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9a-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters and hyphens allowed in %q", k))
	}
	if !regexache.MustCompile(`^[a-z]`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"first character of %q must be a letter", k))
	}
	if regexache.MustCompile(`--`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot contain two consecutive hyphens", k))
	}
	if regexache.MustCompile(`-$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot end with a hyphen", k))
	}
	return
}

func validIdentifierPrefix(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9a-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters and hyphens allowed in %q", k))
	}
	if !regexache.MustCompile(`^[a-z]`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"first character of %q must be a letter", k))
	}
	if regexache.MustCompile(`--`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot contain two consecutive hyphens", k))
	}
	return
}
