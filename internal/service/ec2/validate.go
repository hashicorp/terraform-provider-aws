package ec2

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func validSecurityGroupRuleDescription(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 255 characters: %q", k, value))
	}

	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_IpRange.html. Note that
	// "" is an allowable description value.
	pattern := `^[A-Za-z0-9 \.\_\-\:\/\(\)\#\,\@\[\]\+\=\&\;\{\}\!\$\*]*$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't comply with restrictions (%q): %q",
			k, pattern, value))
	}
	return
}

// validNestedExactlyOneOf is called on the map representing a nested schema element
// Once ExactlyOneOf is supported for nested elements, this should be deprecated.
func validNestedExactlyOneOf(m map[string]interface{}, valid []string) error {
	specified := make([]string, 0)
	for _, k := range valid {
		if v, ok := m[k].(string); ok && v != "" {
			specified = append(specified, k)
		}
	}

	if len(specified) == 0 {
		return fmt.Errorf("one of `%s` must be specified", strings.Join(valid, ", "))
	}
	if len(specified) > 1 {
		return fmt.Errorf("only one of `%s` can be specified, but `%s` were specified.", strings.Join(valid, ", "), strings.Join(specified, ", "))
	}
	return nil
}

func validAmazonSideASN(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	// http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateVpnGateway.html
	asn, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q (%q) must be a 64-bit integer", k, v))
		return
	}

	// https://github.com/hashicorp/terraform-provider-aws/issues/5263
	isLegacyAsn := func(a int64) bool {
		return a == 7224 || a == 9059 || a == 10124 || a == 17493
	}

	if !isLegacyAsn(asn) && ((asn < 64512) || (asn > 65534 && asn < 4200000000) || (asn > 4294967294)) {
		errors = append(errors, fmt.Errorf("%q (%q) must be 7224, 9059, 10124 or 17493 or in the range 64512 to 65534 or 4200000000 to 4294967294", k, v))
	}
	return
}

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
