package eks

import (
	"fmt"
	"regexp"
)

func validClusterName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) < 1 || len(value) > 100 {
		errors = append(errors, fmt.Errorf(
			"%q length must be between 1-100 characters: %q", k, value))
	}

	// https://docs.aws.amazon.com/eks/latest/APIReference/API_CreateCluster.html#API_CreateCluster_RequestSyntax
	pattern := `^[0-9A-Za-z][A-Za-z0-9\-_]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't comply with restrictions (%q): %q",
			k, pattern, value))
	}

	return
}
