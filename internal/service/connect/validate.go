// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"fmt"

	"github.com/YakDriver/regexache"
)

func validDeskPhoneNumber(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`\+[1-9]\d{1,14}$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q (%q) must be a valid phone number", k, v))
	}
	return
}

func validPhoneNumberPrefix(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`\+[0-9]{1,11}`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q (%q) must be a valid phone number prefix and contain + as part of the country code", k, v))
	}
	return
}
