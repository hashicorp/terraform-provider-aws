// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"fmt"

	"github.com/YakDriver/regexache"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

var (
	httpsOrS3URIRegexp = regexache.MustCompile(`^(https|s3)://([^/]+)/?(.*)$`)
	validHTTPSOrS3URI  = validation.All(
		validation.StringMatch(httpsOrS3URIRegexp, "must be HTTPS or Amazon S3 URI"),
		validation.StringLenBetween(0, 1024),
	)
)

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
	maxLength := 63 - sdkid.UniqueIDSuffixLength
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
