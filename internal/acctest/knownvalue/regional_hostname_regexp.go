// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package knownvalue

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var _ knownvalue.Check = regionalHostnameRegexp{}

type regionalHostnameRegexp struct {
	check          string
	region         string
	service        string
	hostnameRegexp *regexp.Regexp
}

// CheckValue determines whether the passed value is of type string, and
// contains a matching sequence of bytes.
func (v regionalHostnameRegexp) CheckValue(other any) error {
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
func (v regionalHostnameRegexp) String() string {
	return v.buildHostnameString()
}

func (v regionalHostnameRegexp) buildHostnameString() string {
	return fmt.Sprintf(`^%s\.%s\.%s\.%s$`, v.hostnameRegexp.String(), v.service, v.region, names.PartitionForRegion(v.region).DNSSuffix())
}

func RegionalHostnameRegexp(service string, hostname *regexp.Regexp) knownvalue.Check {
	return regionalHostnameRegexp{
		check:          "RegionalHostnameRegexp",
		region:         acctest.Region(),
		service:        service,
		hostnameRegexp: hostname,
	}
}
