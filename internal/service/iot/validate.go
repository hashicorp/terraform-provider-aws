// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/YakDriver/regexache"
)

func validThingTypeDescription(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 2028 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 2028 characters", k))
	}
	if !regexache.MustCompile(`[\\p{Graph}\\x20]*`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			`%q must match pattern [\p{Graph}\x20]*`, k))
	}
	return
}

func validThingTypeName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`[0-9A-Za-z_:-]+`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters, colons, underscores and hyphens allowed in %q", k))
	}
	return
}

func validThingTypeSearchableAttribute(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 128 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 128 characters", k))
	}
	if !regexache.MustCompile(`[0-9A-Za-z_.,@/:#-]+`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters, underscores, dots, commas, arobases, slashes, colons, hashes and hyphens allowed in %q", k))
	}
	return
}

func validTopicRuleCloudWatchAlarmStateValue(v any, s string) ([]string, []error) {
	switch v.(string) {
	case
		"OK",
		"ALARM",
		"INSUFFICIENT_DATA":
		return nil, nil
	}

	return nil, []error{fmt.Errorf("State must be one of OK, ALARM, or INSUFFICIENT_DATA")}
}

func validTopicRuleElasticsearchEndpoint(v any, k string) (ws []string, errors []error) {
	value := v.(string)

	// https://docs.aws.amazon.com/iot/latest/apireference/API_ElasticsearchAction.html
	if !regexache.MustCompile(`https?://.*`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q should be an URL: %q",
			k, value))
	}
	return
}

func validTopicRuleFirehoseSeparator(v any, s string) ([]string, []error) {
	switch v.(string) {
	case
		",",
		"\t",
		"\n",
		"\r\n":
		return nil, nil
	}

	return nil, []error{fmt.Errorf(`Separator must be one of ',' (comma), '\t' (tab) '\n' (newline) or '\r\n' (Windows newline)`)}
}

func validTopicRuleName(v any, s string) ([]string, []error) {
	name := v.(string)
	if len(name) < 1 || len(name) > 128 {
		return nil, []error{fmt.Errorf("Name must between 1 and 128 characters long")}
	}

	matched, err := regexp.MatchReader("^[0-9A-Za-z_]+$", strings.NewReader(name))

	if err != nil {
		return nil, []error{err}
	}

	if !matched {
		return nil, []error{fmt.Errorf("Name must match the pattern ^[0-9A-Za-z_]+$")}
	}

	return nil, nil
}
