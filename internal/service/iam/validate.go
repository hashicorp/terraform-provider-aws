package iam

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

func validRolePolicyName(v interface{}, k string) (ws []string, errors []error) {
	// https://github.com/boto/botocore/blob/2485f5c/botocore/data/iam/2010-05-08/service-2.json#L8291-L8296
	value := v.(string)
	if len(value) > 128 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 128 characters", k))
	}
	if !regexp.MustCompile(`^[\w+=,.@-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(`%q must match [\w+=,.@-]`, k))
	}
	return
}

func validRolePolicyNamePrefix(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 100 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 100 characters", k))
	}
	if !regexp.MustCompile(`^[\w+=,.@-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(`%q must match [\w+=,.@-]`, k))
	}
	return
}

func validAccountAlias(v interface{}, k string) (ws []string, es []error) {
	val := v.(string)

	if (len(val) < 3) || (len(val) > 63) {
		es = append(es, fmt.Errorf("%q must contain from 3 to 63 alphanumeric characters or hyphens", k))
	}
	if !regexp.MustCompile("^[a-z0-9][a-z0-9-]+$").MatchString(val) {
		es = append(es, fmt.Errorf("%q must start with an alphanumeric character and only contain lowercase alphanumeric characters and hyphens", k))
	}
	if strings.Contains(val, "--") {
		es = append(es, fmt.Errorf("%q must not contain consecutive hyphens", k))
	}
	if strings.HasSuffix(val, "-") {
		es = append(es, fmt.Errorf("%q must not end in a hyphen", k))
	}
	return
}

func validOpenIDURL(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	u, err := url.Parse(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q has to be a valid URL", k))
		return
	}
	if u.Scheme != "https" {
		errors = append(errors, fmt.Errorf("%q has to use HTTPS scheme (i.e. begin with https://)", k))
	}
	if len(u.Query()) > 0 {
		errors = append(errors, fmt.Errorf("%q cannot contain query parameters per the OIDC standard", k))
	}
	return
}
