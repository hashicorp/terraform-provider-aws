// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

var _ statecheck.StateCheck = expectHasIdentity{}

type expectHasIdentity struct {
	base Base
}

func (e expectHasIdentity) CheckState(ctx context.Context, request statecheck.CheckStateRequest, response *statecheck.CheckStateResponse) {
	resource, ok := e.base.ResourceFromState(request, response)
	if !ok {
		return
	}

	if resource.IdentitySchemaVersion == nil || len(resource.IdentityValues) == 0 {
		response.Error = fmt.Errorf("%s - Identity not found in state. Either the resource does not support identity or the Terraform version running the test does not support identity. (must be v1.12+)", e.base.resourceAddress)
	}
}

func ExpectHasIdentity(resourceAddress string) statecheck.StateCheck {
	return expectHasIdentity{
		base: NewBase(resourceAddress),
	}
}
