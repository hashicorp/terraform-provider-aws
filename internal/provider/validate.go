package provider

import (
	"fmt"
	"time"
)

// ValidAssumeRoleDuration validates a string can be parsed as a valid time.Duration
// and is within a minimum of 15 minutes and maximum of 12 hours
func ValidAssumeRoleDuration(v interface{}, k string) (ws []string, errors []error) {
	duration, err := time.ParseDuration(v.(string))

	if err != nil {
		errors = append(errors, fmt.Errorf("%q cannot be parsed as a duration: %w", k, err))
		return
	}

	if duration.Minutes() < 15 || duration.Hours() > 12 {
		errors = append(errors, fmt.Errorf("duration %q must be between 15 minutes (15m) and 12 hours (12h), inclusive", k))
	}

	return
}
