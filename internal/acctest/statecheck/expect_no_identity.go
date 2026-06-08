// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

var _ statecheck.StateCheck = expectNoIdentity{}

type expectNoIdentity struct {
	base Base
}

func (e expectNoIdentity) CheckState(ctx context.Context, request statecheck.CheckStateRequest, response *statecheck.CheckStateResponse) {
	resource, ok := e.base.ResourceFromState(request, response)
	if !ok {
		return
	}

	if resource.IdentitySchemaVersion != nil || len(resource.IdentityValues) > 0 {
		response.Error = fmt.Errorf("%s - Identity found in state, and was not expected.", e.base.resourceAddress)
	}
}

func ExpectNoIdentity(resourceAddress string) statecheck.StateCheck {
	return expectNoIdentity{
		base: NewBase(resourceAddress),
	}
}
