// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package knownvalue

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

var _ knownvalue.Check = globalARNNoAccountIDExact{}

type globalARNNoAccountIDExact struct {
	service  string
	resource string
}

// CheckValue determines whether the passed value is of type string, and
// contains a matching sequence of bytes.
func (v globalARNNoAccountIDExact) CheckValue(other any) error {
	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for GlobalARNNoAccountIDExact check, got: %T", other)
	}

	if otherVal != v.buildARNString() {
		return fmt.Errorf("expected value %s for GlobalARNNoAccountIDExact check, got: %s", v.buildARNString(), otherVal)
	}

	return nil
}

// String returns the string representation of the value.
func (v globalARNNoAccountIDExact) String() string {
	return v.buildARNString()
}

func (v globalARNNoAccountIDExact) buildARNString() string {
	return arn.ARN{
		AccountID: "",
		Partition: acctest.Partition(),
		Region:    "",
		Service:   v.service,
		Resource:  v.resource,
	}.String()
}

func GlobalARNNoAccountIDExact(service, resource string) knownvalue.Check {
	return globalARNNoAccountIDExact{
		service:  service,
		resource: resource,
	}
}
