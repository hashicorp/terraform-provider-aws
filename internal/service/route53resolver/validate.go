package route53resolver

import (
	"fmt"
	"regexp"
)

func validResolverName(v interface{}, k string) (ws []string, errors []error) {
	// Type: String
	// Length Constraints: Maximum length of 64.
	// Pattern: (?!^[0-9]+$)([a-zA-Z0-9-_' ']+)
	value := v.(string)

	// re2 doesn't support negative lookaheads so check for single numeric character explicitly.
	if regexp.MustCompile(`^[0-9]$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q cannot be a single digit", k))
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9-_' ']+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters, '-', '_' and ' ' are allowed in %q", k))
	}
	if len(value) > 64 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than 64 characters", k))
	}

	return
}
