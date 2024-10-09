// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

type expectNoValueCheck struct {
	base          Base
	attributePath tfjsonpath.Path
}

func (e expectNoValueCheck) CheckState(ctx context.Context, request statecheck.CheckStateRequest, response *statecheck.CheckStateResponse) {
	resource, ok := e.base.ResourceFromState(request, response)
	if !ok {
		return
	}

	if _, err := tfjsonpath.Traverse(resource.AttributeValues, e.attributePath); err == nil {
		response.Error = fmt.Errorf("value for attribute at path: %s.%s exists", resource.Address, e.attributePath.String())

		return
	}
}

func ExpectNoValue(resourceAddress string, attributePath tfjsonpath.Path) statecheck.StateCheck {
	return expectNoValueCheck{
		base:          NewBase(resourceAddress),
		attributePath: attributePath,
	}
}
