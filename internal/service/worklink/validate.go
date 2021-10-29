package worklink

import (
	"fmt"
	"regexp"
)

func validFleetName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[a-z0-9](?:[a-z0-9\-]{0,46}[a-z0-9])?$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters are allowed in %q", k))
	}
	if len(value) < 1 {
		errors = append(errors, fmt.Errorf("%q cannot be shorter than 1 character", k))
	} else if len(value) > 48 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 48 characters", k))
	}

	return
}
