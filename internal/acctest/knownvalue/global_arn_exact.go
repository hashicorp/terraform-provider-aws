// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package knownvalue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

var _ knownvalue.Check = globalARNExact{}

type globalARNExact struct {
	service  string
	resource string
}

// CheckValue determines whether the passed value is of type string, and
// contains a matching sequence of bytes.
func (v globalARNExact) CheckValue(other any) error {
	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for GlobalARNExact check, got: %T", other)
	}

	if otherVal != v.buildARNString() {
		return fmt.Errorf("expected value %s for GlobalARNExact check, got: %s", v.buildARNString(), otherVal)
	}

	return nil
}

// String returns the string representation of the value.
func (v globalARNExact) String() string {
	return v.buildARNString()
}

func (v globalARNExact) buildARNString() string {
	return arn.ARN{
		AccountID: acctest.AccountID(context.Background()),
		Partition: acctest.Partition(),
		Region:    "",
		Service:   v.service,
		Resource:  v.resource,
	}.String()
}

func GlobalARNExact(service, resource string) knownvalue.Check {
	return globalARNExact{
		service:  service,
		resource: resource,
	}
}
