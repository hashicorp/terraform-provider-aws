// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"fmt"
	"strconv"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
)

func validName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) == 0 {
		return // short-circuit
	}
	if len(value) > 32 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 32 characters: %q", k, value))
	}
	if !regexache.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and hyphens allowed in %q: %q",
			k, value))
	}
	if regexache.MustCompile(`^-`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot begin with a hyphen: %q", k, value))
	}
	if regexache.MustCompile(`-$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot end with a hyphen: %q", k, value))
	}
	if regexache.MustCompile(`^internal-`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			`%q cannot begin with "internal-": %q`, k, value))
	}
	return
}

func validNamePrefix(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and hyphens allowed in %q: %q",
			k, value))
	}
	if len(value) > 6 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 6 characters: %q", k, value))
	}
	if regexache.MustCompile(`^-`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot begin with a hyphen: %q", k, value))
	}
	if regexache.MustCompile(`^internal-`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			`%q cannot begin with "internal-": %q`, k, value))
	}
	return
}

func validTargetGroupName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 32 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 32 characters", k))
	}
	if !regexache.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and hyphens allowed in %q", k))
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

func validTargetGroupNamePrefix(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	prefixMaxLength := 32 - id.UniqueIDSuffixLength
	if len(value) > prefixMaxLength {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than %d characters", k, prefixMaxLength))
	}
	if !regexache.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and hyphens allowed in %q", k))
	}
	if regexache.MustCompile(`^-`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot begin with a hyphen", k))
	}
	return
}

func validTargetGroupHealthInput(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if value != "off" {
		_, err := strconv.Atoi(value)
		if err != nil {
			errors = append(errors, fmt.Errorf(
				"%q must be an integer or 'off'", k))
		}
	}
	return
}

func validTargetGroupHealthPercentageInput(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if value != "off" {
		intValue, err := strconv.Atoi(value)
		if err != nil || intValue < 1 || intValue > 100 {
			errors = append(errors, fmt.Errorf(
				"%q must be an integer between 0 and 100 or 'off'", k))
		}
	}
	return
}
