// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"fmt"

	"github.com/YakDriver/regexache"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
)

func validSecretName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9A-Za-z/_+=.@-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and /_+=.@- special characters are allowed in %q", k))
	}
	if len(value) > 512 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than 512 characters", k))
	}
	return
}

func validSecretNamePrefix(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9A-Za-z/_+=.@-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and /_+=.@- special characters are allowed in %q", k))
	}
	prefixMaxLength := 512 - sdkid.UniqueIDSuffixLength
	if len(value) > prefixMaxLength {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than %d characters", k, prefixMaxLength))
	}
	return
}
