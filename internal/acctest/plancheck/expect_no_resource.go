// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package plancheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

type expectNoResourceCheck struct {
	resourceAddress string
}

func (e expectNoResourceCheck) CheckPlan(ctx context.Context, request plancheck.CheckPlanRequest, response *plancheck.CheckPlanResponse) {
	for _, r := range request.Plan.ResourceChanges {
		if e.resourceAddress == r.Address {
			response.Error = fmt.Errorf("%s - Resource found in plan", e.resourceAddress)

			return
		}
	}
}

func ExpectNoResource(resourceAddress string) plancheck.PlanCheck {
	return expectNoResourceCheck{
		resourceAddress: resourceAddress,
	}
}
