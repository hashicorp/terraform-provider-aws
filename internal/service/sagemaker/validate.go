package sagemaker

import (
	"fmt"
	"regexp"
)

func validEnvironment(v interface{}, k string) (ws []string, errors []error) {
	value := v.(map[string]interface{})
	for envK, envV := range value {
		if !regexp.MustCompile(`^[0-9A-Za-z_]+$`).MatchString(envK) {
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
		if regexp.MustCompile(`^[0-9]`).MatchString(envK) {
			errors = append(errors, fmt.Errorf(
				"%q cannot begin with a digit: %q", k, envK))
		}
	}
	return
}

func validImage(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`[\S]+`).MatchString(value) {
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

func validModelDataURL(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^(https|s3)://([^/]+)/?(.*)$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q must be a valid path: %q",
			k, value))
	}
	if len(value) > 1024 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 1024 characters: %q", k, value))
	}
	if !regexp.MustCompile(`^(https|s3)://`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q must be a path that starts with either s3 or https: %q", k, value))
	}
	return
}

func validName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and hyphens allowed in %q: %q",
			k, value))
	}
	if len(value) > 63 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 63 characters: %q", k, value))
	}
	if regexp.MustCompile(`^-`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot begin with a hyphen: %q", k, value))
	}
	return
}
