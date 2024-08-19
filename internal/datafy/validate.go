package datafy

import (
	"fmt"
	"golang.org/x/mod/semver"
)

func ValidAgentVersion(v interface{}, k string) (warning []string, errors []error) {
	value := v.(string)
	if value == "latest" || semver.IsValid(v.(string)) {
		return
	}

	errors = append(errors, fmt.Errorf("%q is not a valid semantic version", k))
	return
}
