package elasticache

import (
	"fmt"
	"regexp"
)

const (
	versionStringRegexpInternalPattern = `[[:digit:]]+(.[[:digit:]]+){2}`
	versionStringRegexpPattern         = "^" + versionStringRegexpInternalPattern + "$"
)

var versionStringRegexp = regexp.MustCompile(versionStringRegexpPattern)

func validReplicationGroupAuthToken(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if (len(value) < 16) || (len(value) > 128) {
		errors = append(errors, fmt.Errorf(
			"%q must contain from 16 to 128 alphanumeric characters or symbols (excluding @, \", and /)", k))
	}
	if !regexp.MustCompile(`^[^@"\/]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters or symbols (excluding @, \", and /) allowed in %q", k))
	}
	return
}

func validVersionString(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if !versionStringRegexp.MatchString(value) {
		errors = append(errors, fmt.Errorf("%s: must be a version string matching x.y.z", k))
	}

	return
}
