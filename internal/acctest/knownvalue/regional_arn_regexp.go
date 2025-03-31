// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var _ knownvalue.Check = regionalARNRegexp{}

type regionalARNRegexp struct {
	check          string
	region         string
	service        string
	resourceRegexp *regexp.Regexp
}

// CheckValue determines whether the passed value is of type string, and
// contains a matching sequence of bytes.
func (v regionalARNRegexp) CheckValue(other any) error {
	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for %s check, got: %T", v.check, other)
	}

	re, err := regexp.Compile(v.buildARNString())
	if err != nil {
		return fmt.Errorf("unable to compile ARN regexp (%s): %w", v.buildARNString(), err)
	}

	if !re.MatchString(otherVal) {
		return fmt.Errorf("expected regex match %s for %s check, got: %s", re.String(), v.check, otherVal)
	}

	return nil
}

// String returns the string representation of the value.
func (v regionalARNRegexp) String() string {
	return v.buildARNString()
}

func (v regionalARNRegexp) buildARNString() string {
	return `^` + arn.ARN{
		AccountID: acctest.AccountID(context.Background()),
		Partition: names.PartitionForRegion(v.region).ID(),
		Region:    v.region,
		Service:   v.service,
		Resource:  v.resourceRegexp.String(),
	}.String() + `$`
}

func RegionalARNRegexp(service string, resource *regexp.Regexp) knownvalue.Check {
	return regionalARNRegexp{
		check:          "RegionalARNRegexp",
		region:         acctest.Region(),
		service:        service,
		resourceRegexp: resource,
	}
}

func RegionalARNAlternateRegionRegexp(service string, resource *regexp.Regexp) knownvalue.Check {
	return regionalARNRegexp{
		check:          "RegionalARNAlternateRegionRegexp",
		region:         acctest.AlternateRegion(),
		service:        service,
		resourceRegexp: resource,
	}
}

func RegionalARNThirdRegionRegexp(service string, resource *regexp.Regexp) knownvalue.Check {
	return regionalARNRegexp{
		check:          "RegionalARNThirdRegionRegexp",
		region:         acctest.ThirdRegion(),
		service:        service,
		resourceRegexp: resource,
	}
}
