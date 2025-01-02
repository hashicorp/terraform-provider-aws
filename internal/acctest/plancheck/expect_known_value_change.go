// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plancheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

type expectKnownValueChangeCheck struct {
	base               Base
	attributePath      tfjsonpath.Path
	oldValue, newValue knownvalue.Check
}

func (e expectKnownValueChangeCheck) CheckPlan(ctx context.Context, request plancheck.CheckPlanRequest, response *plancheck.CheckPlanResponse) {
	resource, ok := e.base.ResourceFromState(request, response)
	if !ok {
		return
	}

	old, err := tfjsonpath.Traverse(resource.Change.Before, e.attributePath)
	if err != nil {
		response.Error = err

		return
	}

	if err := e.oldValue.CheckValue(old); err != nil {
		response.Error = fmt.Errorf("checking old value for attribute at path: %s.%s, err: %s", resource.Address, e.attributePath.String(), err)

		return
	}

	new, err := tfjsonpath.Traverse(resource.Change.After, e.attributePath)
	if err != nil {
		response.Error = err

		return
	}

	if err := e.newValue.CheckValue(new); err != nil {
		response.Error = fmt.Errorf("checking new value for attribute at path: %s.%s, err: %s", resource.Address, e.attributePath.String(), err)

		return
	}
}

func ExpectKnownValueChange(resourceAddress string, attributePath tfjsonpath.Path, oldValue, newValue knownvalue.Check) plancheck.PlanCheck {
	return expectKnownValueChangeCheck{
		base:          NewBase(resourceAddress),
		attributePath: attributePath,
		oldValue:      oldValue,
		newValue:      newValue,
	}
}
