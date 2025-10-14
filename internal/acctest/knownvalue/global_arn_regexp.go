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
)

var _ knownvalue.Check = globalARNRegexp{}

type globalARNRegexp struct {
	service        string
	resourceRegexp *regexp.Regexp
}

// CheckValue determines whether the passed value is of type string, and
// contains a matching sequence of bytes.
func (v globalARNRegexp) CheckValue(other any) error {
	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for GlobalARNRegexp check, got: %T", other)
	}

	re, err := regexp.Compile(v.buildARNString())
	if err != nil {
		return fmt.Errorf("unable to compile ARN regexp (%s): %w", v.buildARNString(), err)
	}

	if !re.MatchString(otherVal) {
		return fmt.Errorf("expected regex match %s for GlobalARNRegexp check, got: %s", re.String(), otherVal)
	}

	return nil
}

// String returns the string representation of the value.
func (v globalARNRegexp) String() string {
	return v.buildARNString()
}

func (v globalARNRegexp) buildARNString() string {
	return `^` + arn.ARN{
		AccountID: acctest.AccountID(context.Background()),
		Partition: acctest.Partition(),
		Region:    "",
		Service:   v.service,
		Resource:  v.resourceRegexp.String(),
	}.String() + `$`
}

func GlobalARNRegexp(service string, resource *regexp.Regexp) knownvalue.Check {
	return globalARNRegexp{
		service:        service,
		resourceRegexp: resource,
	}
}
