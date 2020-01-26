package validation

import (
	"fmt"
	"time"
)

// ValidateRFC3339TimeString is a ValidateFunc that ensures a string parses
// as time.RFC3339 format
func ValidateRFC3339TimeString(v interface{}, k string) (ws []string, errors []error) {
	if _, err := time.Parse(time.RFC3339, v.(string)); err != nil {
		errors = append(errors, fmt.Errorf("%q: invalid RFC3339 timestamp", k))
	}
	return
}
