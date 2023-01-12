package transfer

import (
	"fmt"
	"regexp"
)

func validServerID(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	// https://docs.aws.amazon.com/transfer/latest/userguide/API_CreateUser.html
	pattern := `^s-([0-9a-f]{17})$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q isn't a valid transfer server id (only lowercase alphanumeric characters are allowed): %q",
			k, value))
	}

	return
}

func validUserName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	// https://docs.aws.amazon.com/transfer/latest/userguide/API_CreateUser.html
	if !regexp.MustCompile(`^[\w][\w@.-]{2,99}$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("Invalid %q: must be between 3 and 100 alphanumeric characters, special characters, hyphens, or underscores. However, it cannot begin with a hyphen, period, or at sign", k))
	}
	return
}
