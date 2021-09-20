package sfn

import (
	"fmt"
	"regexp"
)

func validStateMachineName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 80 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 80 characters", k))
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9-_]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q must be composed with only these characters [a-zA-Z0-9-_]: %v", k, value))
	}
	return
}
