package neptune

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validEngine() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"neptune",
	}, false)
}

func validEventSubscriptionName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and hyphens allowed in %q", k))
	}
	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than 255 characters", k))
	}
	return
}

func validEventSubscriptionNamePrefix(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9A-Za-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and hyphens allowed in %q", k))
	}
	prefixMaxLength := 255 - resource.UniqueIDSuffixLength
	if len(value) > prefixMaxLength {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than %d characters", k, prefixMaxLength))
	}
	return
}

func validIdentifier(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9a-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters and hyphens allowed in %q", k))
	}
	if !regexp.MustCompile(`^[a-z]`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"first character of %q must be a letter", k))
	}
	if regexp.MustCompile(`--`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot contain two consecutive hyphens", k))
	}
	if regexp.MustCompile(`-$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot end with a hyphen", k))
	}
	return
}

func validIdentifierPrefix(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9a-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters and hyphens allowed in %q", k))
	}
	if !regexp.MustCompile(`^[a-z]`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"first character of %q must be a letter", k))
	}
	if regexp.MustCompile(`--`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot contain two consecutive hyphens", k))
	}
	return
}

func validParamGroupName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9a-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters and hyphens allowed in %q", k))
	}
	if !regexp.MustCompile(`^[a-z]`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"first character of %q must be a letter", k))
	}
	if regexp.MustCompile(`--`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot contain two consecutive hyphens", k))
	}
	if regexp.MustCompile(`-$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot end with a hyphen", k))
	}
	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than 255 characters", k))
	}
	return
}

func validParamGroupNamePrefix(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9a-z-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters and hyphens allowed in %q", k))
	}
	if !regexp.MustCompile(`^[a-z]`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"first character of %q must be a letter", k))
	}
	if regexp.MustCompile(`--`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot contain two consecutive hyphens", k))
	}
	prefixMaxLength := 255 - resource.UniqueIDSuffixLength
	if len(value) > prefixMaxLength {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than %d characters", k, prefixMaxLength))
	}
	return
}

func validSubnetGroupName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[ .0-9a-z-_]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters, hyphens, underscores, periods, and spaces allowed in %q", k))
	}
	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 255 characters", k))
	}
	if value == "default" {
		errors = append(errors, fmt.Errorf(
			"%q is not allowed as %q", "Default", k))
	}
	return
}

func validSubnetGroupNamePrefix(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[ .0-9a-z-_]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only lowercase alphanumeric characters, hyphens, underscores, periods, and spaces allowed in %q", k))
	}
	prefixMaxLength := 255 - resource.UniqueIDSuffixLength
	if len(value) > prefixMaxLength {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than %d characters", k, prefixMaxLength))
	}
	return
}
