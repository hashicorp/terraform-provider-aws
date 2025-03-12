// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var _ knownvalue.Check = regionalARNExact{}

type regionalARNExact struct {
	check    string
	region   string
	service  string
	resource string
}

// CheckValue determines whether the passed value is of type string, and
// contains a matching sequence of bytes.
func (v regionalARNExact) CheckValue(other any) error {
	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for %s check, got: %T", v.check, other)
	}

	if otherVal != v.buildARNString() {
		return fmt.Errorf("expected value %s for %s check, got: %s", v.buildARNString(), v.check, otherVal)
	}

	return nil
}

// String returns the string representation of the value.
func (v regionalARNExact) String() string {
	return v.buildARNString()
}

func (v regionalARNExact) buildARNString() string {
	return arn.ARN{
		AccountID: acctest.AccountID(context.Background()),
		Partition: names.PartitionForRegion(v.region).ID(),
		Region:    v.region,
		Service:   v.service,
		Resource:  v.resource,
	}.String()
}

func RegionalARNExact(service, resource string) knownvalue.Check {
	return regionalARNExact{
		check:    "RegionalARNExact",
		region:   acctest.Region(),
		service:  service,
		resource: resource,
	}
}

func RegionalARNAlternateRegionExact(service, resource string) knownvalue.Check {
	return regionalARNExact{
		check:    "RegionalARNAlternateRegionExact",
		region:   acctest.AlternateRegion(),
		service:  service,
		resource: resource,
	}
}

func RegionalARNThirdRegionExact(service, resource string) knownvalue.Check {
	return regionalARNExact{
		check:    "RegionalARNThirdRegionExact",
		region:   acctest.ThirdRegion(),
		service:  service,
		resource: resource,
	}
}
