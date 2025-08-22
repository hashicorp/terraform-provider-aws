// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"fmt"
	"reflect"

	"github.com/YakDriver/regexache"
)

func validMetricName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9A-Za-z]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"Only alphanumeric characters allowed in %q: %q",
			k, value))
	}
	return
}

func sliceContainsMap(l []any, m map[string]any) (int, bool) {
	for i, t := range l {
		if reflect.DeepEqual(m, t.(map[string]any)) {
			return i, true
		}
	}

	return -1, false
}
