package ssm

import (
	"fmt"
	"regexp"
)

func validName(v interface{}, k string) (ws []string, errors []error) {
	// http://docs.aws.amazon.com/systems-manager/latest/APIReference/API_CreateDocument.html#EC2-CreateDocument-request-Name
	value := v.(string)

	if !regexp.MustCompile(`^[a-zA-Z0-9_\-.]{3,128}$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			`Only alphanumeric characters, hyphens, dots & underscores allowed in %q: %q (Must satisfy regular expression pattern: ^[a-zA-Z0-9_\-.]{3,128}$)`,
			k, value))
	}

	return
}
