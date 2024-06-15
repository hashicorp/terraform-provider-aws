// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plancheck

import (
	"context"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

type expectKnownChange struct {
	resourceAddress string
	attributePath   tfjsonpath.Path
	knownvalueFrom  knownvalue.Check
	knownvalueTo    knownvalue.Check
}

func (e expectKnownChange) CheckPlan(ctx context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	var rc *tfjson.ResourceChange

	for _, resourceChange := range req.Plan.ResourceChanges {
		if resourceChange.Address == e.resourceAddress {
			rc = resourceChange
			break
		}
	}

	if rc == nil {
		resp.Error = fmt.Errorf("%s - Resource not found in plan", e.resourceAddress)
		return
	}

	before, err := tfjsonpath.Traverse(rc.Change.Before, e.attributePath)

	if err != nil {
		resp.Error = err
		return
	}

	if err := e.knownvalueFrom.CheckValue(before); err != nil {
		resp.Error = err
		return
	}

	after, err := tfjsonpath.Traverse(rc.Change.After, e.attributePath)

	if err != nil {
		resp.Error = err
		return
	}

	if err := e.knownvalueTo.CheckValue(after); err != nil {
		resp.Error = err
		return
	}
}

func ExpectKnownChange(resourceAddress string, attributePath tfjsonpath.Path, knownvalueFrom, knownvalueTo knownvalue.Check) plancheck.PlanCheck {
	return expectKnownChange{
		resourceAddress: resourceAddress,
		attributePath:   attributePath,
		knownvalueFrom:  knownvalueFrom,
		knownvalueTo:    knownvalueTo,
	}
}
