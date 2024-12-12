// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plancheck

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

type Base struct {
	resourceAddress string
}

func NewBase(resourceAddress string) Base {
	return Base{
		resourceAddress: resourceAddress,
	}
}

func (b Base) ResourceFromState(req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) (*tfjson.ResourceChange, bool) {
	var resource *tfjson.ResourceChange

	if req.Plan == nil {
		resp.Error = fmt.Errorf("plan is nil")

		return nil, false
	}

	for _, r := range req.Plan.ResourceChanges {
		if b.resourceAddress == r.Address {
			resource = r

			break
		}
	}

	if resource == nil {
		resp.Error = fmt.Errorf("%s - Resource not found in plan", b.resourceAddress)

		return nil, false
	}

	return resource, true
}

func (b Base) ResourceAddress() string {
	return b.resourceAddress
}
