// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
)

func validEndpointID(v any, k string) (ws []string, es []error) {
	val := v.(string)

	if len(val) > 255 {
		es = append(es, fmt.Errorf("%q must not be longer than 255 characters", k))
	}
	if !regexache.MustCompile("^[A-Za-z][0-9A-Za-z-]+$").MatchString(val) {
		es = append(es, fmt.Errorf("%q must start with a letter, only contain alphanumeric characters and hyphens", k))
	}
	if strings.Contains(val, "--") {
		es = append(es, fmt.Errorf("%q must not contain consecutive hyphens", k))
	}
	if strings.HasSuffix(val, "-") {
		es = append(es, fmt.Errorf("%q must not end in a hyphen", k))
	}

	return
}

func validReplicationInstanceID(v any, k string) (ws []string, es []error) {
	val := v.(string)

	if len(val) > 63 {
		es = append(es, fmt.Errorf("%q must not be longer than 63 characters", k))
	}
	if !regexache.MustCompile("^[A-Za-z][0-9A-Za-z-]+$").MatchString(val) {
		es = append(es, fmt.Errorf("%q must start with a letter, only contain alphanumeric characters and hyphens", k))
	}
	if strings.Contains(val, "--") {
		es = append(es, fmt.Errorf("%q must not contain consecutive hyphens", k))
	}
	if strings.HasSuffix(val, "-") {
		es = append(es, fmt.Errorf("%q must not end in a hyphen", k))
	}

	return
}

func validReplicationSubnetGroupID(v any, k string) (ws []string, es []error) {
	val := v.(string)

	if val == "default" {
		es = append(es, fmt.Errorf("%q must not be default", k))
	}
	if len(val) > 255 {
		es = append(es, fmt.Errorf("%q must not be longer than 255 characters", k))
	}
	if !regexache.MustCompile(`^[0-9A-Za-z_ .-]+$`).MatchString(val) {
		es = append(es, fmt.Errorf("%q must only contain alphanumeric characters, periods, spaces, underscores and hyphens", k))
	}

	return
}

func validReplicationTaskID(v any, k string) (ws []string, es []error) {
	val := v.(string)

	if len(val) > 255 {
		es = append(es, fmt.Errorf("%q must not be longer than 255 characters", k))
	}
	if !regexache.MustCompile("^[A-Za-z][0-9A-Za-z-]+$").MatchString(val) {
		es = append(es, fmt.Errorf("%q must start with a letter, only contain alphanumeric characters and hyphens", k))
	}
	if strings.Contains(val, "--") {
		es = append(es, fmt.Errorf("%q must not contain consecutive hyphens", k))
	}
	if strings.HasSuffix(val, "-") {
		es = append(es, fmt.Errorf("%q must not end in a hyphen", k))
	}

	return
}
