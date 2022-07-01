package secretsmanager

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func validSecretName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9A-Za-z/_+=.@-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and /_+=.@- special characters are allowed in %q", k))
	}
	if len(value) > 512 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than 512 characters", k))
	}
	return
}

func validSecretNamePrefix(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9A-Za-z/_+=.@-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters and /_+=.@- special characters are allowed in %q", k))
	}
	prefixMaxLength := 512 - resource.UniqueIDSuffixLength
	if len(value) > prefixMaxLength {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than %d characters", k, prefixMaxLength))
	}
	return
}
