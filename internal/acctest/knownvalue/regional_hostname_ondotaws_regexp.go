// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package knownvalue

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

var _ knownvalue.Check = regionalHostnameOnDotAWSRegexp{}

type regionalHostnameOnDotAWSRegexp struct {
	check          string
	region         string
	service        string
	hostnameRegexp *regexp.Regexp
}

// CheckValue determines whether the passed value is of type string, and
// contains a matching sequence of bytes.
func (v regionalHostnameOnDotAWSRegexp) CheckValue(other any) error {
	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for %s check, got: %T", v.check, other)
	}

	re, err := regexp.Compile(v.buildHostnameString())
	if err != nil {
		return fmt.Errorf("unable to compile hostname regexp (%s): %w", v.buildHostnameString(), err)
	}

	if !re.MatchString(otherVal) {
		return fmt.Errorf("expected regex match %s for %s check, got: %s", re.String(), v.check, otherVal)
	}

	return nil
}

// String returns the string representation of the value.
func (v regionalHostnameOnDotAWSRegexp) String() string {
	return v.buildHostnameString()
}

func (v regionalHostnameOnDotAWSRegexp) buildHostnameString() string {
	return fmt.Sprintf(`^%s\.%s\.%s\.%s$`, v.hostnameRegexp.String(), v.service, v.region, "on.aws")
}

func RegionalHostnameOnDotAWSRegexp(service string, hostname *regexp.Regexp) knownvalue.Check { // nosemgrep:ci.aws-in-func-name
	return regionalHostnameOnDotAWSRegexp{
		check:          "RegionalHostnameOnDotAWSRegexp",
		region:         acctest.Region(),
		service:        service,
		hostnameRegexp: hostname,
	}
}
