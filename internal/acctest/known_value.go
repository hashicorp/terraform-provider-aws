// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

var _ knownvalue.Check = globalARNCheck{}

type globalARNCheck struct {
	arnService  string
	arnResource string
}

func (v globalARNCheck) CheckValue(other any) error {
	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for GlobalARN check, got: %T", other)
	}

	arnValue := globalARNValue(context.Background(), v.arnService, v.arnResource)

	if otherVal != arnValue {
		return fmt.Errorf("expected value %s for GlobalARN check, got: %s", arnValue, otherVal)
	}

	return nil
}

// String returns the string representation of the value.
func (v globalARNCheck) String() string {
	return globalARNValue(context.Background(), v.arnService, v.arnResource)
}

func GlobalARN(arnService, arnResource string) globalARNCheck {
	return globalARNCheck{
		arnService:  arnService,
		arnResource: arnResource,
	}
}
