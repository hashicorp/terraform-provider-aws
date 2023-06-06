package storagegateway

import (
	"fmt"
	"regexp"
	"strconv"
)

func valid4ByteASN(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	asn, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q (%q) must be a 64-bit integer", k, v))
		return
	}

	if asn < 0 || asn > 4294967295 {
		errors = append(errors, fmt.Errorf("%q (%q) must be in the range 0 to 4294967295", k, v))
	}
	return
}

func validLinuxFileMode(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-7]{4}$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only valid linux mode is allowed in %q", k))
	}
	return
}
