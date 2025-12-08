// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var _ statecheck.StateCheck = expectAttributeFormatCheck{}

type expectAttributeFormatCheck struct {
	base          Base
	attributePath tfjsonpath.Path
	format        string
}

func (e expectAttributeFormatCheck) CheckState(ctx context.Context, request statecheck.CheckStateRequest, response *statecheck.CheckStateResponse) {
	resource, ok := e.base.ResourceFromState(request, response)
	if !ok {
		return
	}

	value, err := tfjsonpath.Traverse(resource.AttributeValues, e.attributePath)
	if err != nil {
		response.Error = err
		return
	}

	otherVal, ok := value.(string)
	if !ok {
		response.Error = fmt.Errorf("expected string value for ExpectAttributeFormat check, got: %T", value)
		return
	}

	expectedValue, err := populateFromResourceState(e.format, resource)
	if err != nil {
		response.Error = err
		return
	}

	if otherVal != expectedValue {
		response.Error = fmt.Errorf("expected value %s for ExpectAttributeFormat check, got: %s", expectedValue, otherVal)
		return
	}
}

func ExpectAttributeFormat(resourceAddress string, attributePath tfjsonpath.Path, format string) statecheck.StateCheck {
	return expectAttributeFormatCheck{
		base:          NewBase(resourceAddress),
		attributePath: attributePath,
		format:        format,
	}
}
