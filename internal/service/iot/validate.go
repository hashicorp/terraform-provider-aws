package iot

import (
	"fmt"
	"regexp"
	"strings"
)

func validThingTypeDescription(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 2028 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 2028 characters", k))
	}
	if !regexp.MustCompile(`[\\p{Graph}\\x20]*`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			`%q must match pattern [\p{Graph}\x20]*`, k))
	}
	return
}

func validThingTypeName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`[a-zA-Z0-9:_-]+`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters, colons, underscores and hyphens allowed in %q", k))
	}
	return
}

func validThingTypeSearchableAttribute(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 128 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 128 characters", k))
	}
	if !regexp.MustCompile(`[a-zA-Z0-9_.,@/:#-]+`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters, underscores, dots, commas, arobases, slashes, colons, hashes and hyphens allowed in %q", k))
	}
	return
}

func validTopicRuleCloudWatchAlarmStateValue(v interface{}, s string) ([]string, []error) {
	switch v.(string) {
	case
		"OK",
		"ALARM",
		"INSUFFICIENT_DATA":
		return nil, nil
	}

	return nil, []error{fmt.Errorf("State must be one of OK, ALARM, or INSUFFICIENT_DATA")}
}

func validTopicRuleElasticSearchEndpoint(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	// https://docs.aws.amazon.com/iot/latest/apireference/API_ElasticsearchAction.html
	if !regexp.MustCompile(`https?://.*`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q should be an URL: %q",
			k, value))
	}
	return
}

func validTopicRuleFirehoseSeparator(v interface{}, s string) ([]string, []error) {
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

func validTopicRuleName(v interface{}, s string) ([]string, []error) {
	name := v.(string)
	if len(name) < 1 || len(name) > 128 {
		return nil, []error{fmt.Errorf("Name must between 1 and 128 characters long")}
	}

	matched, err := regexp.MatchReader("^[a-zA-Z0-9_]+$", strings.NewReader(name))

	if err != nil {
		return nil, []error{err}
	}

	if !matched {
		return nil, []error{fmt.Errorf("Name must match the pattern ^[a-zA-Z0-9_]+$")}
	}

	return nil, nil
}
