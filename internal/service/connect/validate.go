package connect

import (
	"fmt"
	"regexp"
)

func validDeskPhoneNumber(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`\+[1-9]\d{1,14}$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q (%q) must be a valid phone number", k, v))
	}
	return
}
