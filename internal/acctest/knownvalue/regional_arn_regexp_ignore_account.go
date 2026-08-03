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

const (
	accountIDRegexpPattern = `\d{12}`
)

var _ knownvalue.Check = regionalARNRegexpIgnoreAccount{}

type regionalARNRegexpIgnoreAccount struct {
	region         string
	service        string
	resourceRegexp *regexp.Regexp
}

// CheckValue determines whether the passed value is of type string, and
// contains a matching sequence of bytes.
func (v regionalARNRegexpIgnoreAccount) CheckValue(other any) error {
	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for RegionalARNRegexpIgnoreAccount check, got: %T", other)
	}

	re, err := regexp.Compile(v.buildARNString())
	if err != nil {
		return fmt.Errorf("unable to compile ARN regexp (%s): %w", v.buildARNString(), err)
	}

	if !re.MatchString(otherVal) {
		return fmt.Errorf("expected regex match %s for RegionalARNRegexpIgnoreAccount check, got: %s", re.String(), otherVal)
	}

	return nil
}

// String returns the string representation of the value.
func (v regionalARNRegexpIgnoreAccount) String() string {
	return v.buildARNString()
}

func (v regionalARNRegexpIgnoreAccount) buildARNString() string {
	return `^` + arn.ARN{
		AccountID: accountIDRegexpPattern,
		Partition: names.PartitionForRegion(v.region).ID(),
		Region:    v.region,
		Service:   v.service,
		Resource:  v.resourceRegexp.String(),
	}.String() + `$`
}

func RegionalARNRegexpIgnoreAccount(service string, resource *regexp.Regexp) knownvalue.Check {
	return regionalARNRegexpIgnoreAccount{
		region:         acctest.Region(),
		service:        service,
		resourceRegexp: resource,
	}
}
