// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package knownvalue

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var _ knownvalue.Check = regionalARNRegexpNoAccountID{}

type regionalARNRegexpNoAccountID struct {
	region         string
	service        string
	resourceRegexp *regexp.Regexp
}

// CheckValue determines whether the passed value is of type string, and
// contains a matching sequence of bytes.
func (v regionalARNRegexpNoAccountID) CheckValue(other any) error {
	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for RegionalARNRegexpNoAccountID check, got: %T", other)
	}

	re, err := regexp.Compile(v.buildARNString())
	if err != nil {
		return fmt.Errorf("unable to compile ARN regexp (%s): %w", v.buildARNString(), err)
	}

	if !re.MatchString(otherVal) {
		return fmt.Errorf("expected regex match %s for RegionalARNRegexpNoAccountID check, got: %s", re.String(), otherVal)
	}

	return nil
}

// String returns the string representation of the value.
func (v regionalARNRegexpNoAccountID) String() string {
	return v.buildARNString()
}

func (v regionalARNRegexpNoAccountID) buildARNString() string {
	return `^` + arn.ARN{
		AccountID: "",
		Partition: names.PartitionForRegion(v.region).ID(),
		Region:    v.region,
		Service:   v.service,
		Resource:  v.resourceRegexp.String(),
	}.String() + `$`
}

func RegionalARNRegexpNoAccountID(service string, resource *regexp.Regexp) knownvalue.Check {
	return regionalARNRegexpNoAccountID{
		region:         acctest.Region(),
		service:        service,
		resourceRegexp: resource,
	}
}
