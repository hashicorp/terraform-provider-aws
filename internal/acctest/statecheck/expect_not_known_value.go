// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

type expectNotKnownValueCheck struct {
	base          Base
	attributePath tfjsonpath.Path
	notValue      knownvalue.Check
}

func (e expectNotKnownValueCheck) CheckState(ctx context.Context, request statecheck.CheckStateRequest, response *statecheck.CheckStateResponse) {
	resource, ok := e.base.ResourceFromState(request, response)
	if !ok {
		return
	}

	value, err := tfjsonpath.Traverse(resource.AttributeValues, e.attributePath)
	if err != nil {
		response.Error = err

		return
	}

	err = e.notValue.CheckValue(value)
	if err == nil {
		response.Error = fmt.Errorf("value for attribute at path: %s.%s is %v", resource.Address, e.attributePath.String(), value)

		return
	}
}

func ExpectNotKnownValue(resourceAddress string, attributePath tfjsonpath.Path, notValue knownvalue.Check) statecheck.StateCheck {
	return expectNotKnownValueCheck{
		base:          NewBase(resourceAddress),
		attributePath: attributePath,
		notValue:      notValue,
	}
}
