package cloudfront

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform/helper/resource"
)

var regionRegexp = regexp.MustCompile(`^[a-z]{2}(-[a-z]+)+-\d$`)

func validPublicKeyName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9A-Za-z_-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters, underscores and hyphens allowed in %q", k))
	}
	if len(value) > 128 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than 128 characters", k))
	}
	return
}

func validPublicKeyNamePrefix(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9A-Za-z_-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters, underscores and hyphens allowed in %q", k))
	}
	prefixMaxLength := 128 - resource.UniqueIDSuffixLength
	if len(value) > prefixMaxLength {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than %d characters", k, prefixMaxLength))
	}
	return
}
