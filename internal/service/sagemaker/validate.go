// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
)

func validEnvironment(v any, k string) (ws []string, errors []error) {
	value := v.(map[string]any)
	for envK, envV := range value {
		if !regexache.MustCompile(`^[0-9A-Za-z_]+$`).MatchString(envK) {
			errors = append(errors, fmt.Errorf(
				"only alphanumeric characters and underscore allowed in %q: %q",
				k, envK))
		}
		if len(envK) > 1024 {
			errors = append(errors, fmt.Errorf(
				"%q cannot be longer than 1024 characters: %q", k, envK))
		}
		if len(envV.(string)) > 1024 {
			errors = append(errors, fmt.Errorf(
				"%q cannot be longer than 1024 characters: %q", k, envV.(string)))
		}
		if regexache.MustCompile(`^[0-9]`).MatchString(envK) {
			errors = append(errors, fmt.Errorf(
				"%q cannot begin with a digit: %q", k, envK))
		}
	}
	return
}

func validImage(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`[\S]+`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"no whitespace allowed in %q: %q",
			k, value))
	}
	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 255 characters: %q", k, value))
	}
	return
}

func validModelDataURL(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^(https|s3)://([^/]+)/?(.*)$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q must be a valid path: %q",
			k, value))
	}
	if len(value) > 1024 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 1024 characters: %q", k, value))
	}
	if !regexache.MustCompile(`^(https|s3)://`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q must be a path that starts with either s3 or https: %q", k, value))
	}
	return
}

func validName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and hyphens allowed in %q: %q",
			k, value))
	}
	if len(value) > 63 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 63 characters: %q", k, value))
	}
	if regexache.MustCompile(`^-`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot begin with a hyphen: %q", k, value))
	}
	return
}

func validPrefix(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and hyphens allowed in %q: %q",
			k, value))
	}
	maxLength := 63 - id.UniqueIDSuffixLength
	if len(value) > maxLength {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than %d characters: %q", k, maxLength, value))
	}
	if regexache.MustCompile(`^-`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot begin with a hyphen: %q", k, value))
	}
	return
}
