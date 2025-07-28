// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
)

type Base struct {
	resourceAddress string
}

func NewBase(resourceAddress string) Base {
	return Base{
		resourceAddress: resourceAddress,
	}
}

func (b Base) ResourceFromState(req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) (*tfjson.StateResource, bool) {
	var resource *tfjson.StateResource

	if req.State == nil {
		resp.Error = fmt.Errorf("state is nil")

		return nil, false
	}

	if req.State.Values == nil {
		resp.Error = fmt.Errorf("state does not contain any state values")

		return nil, false
	}

	if req.State.Values.RootModule == nil {
		resp.Error = fmt.Errorf("state does not contain a root module")

		return nil, false
	}

	for _, r := range req.State.Values.RootModule.Resources {
		if b.resourceAddress == r.Address {
			resource = r

			break
		}
	}

	if resource == nil {
		resp.Error = fmt.Errorf("%s - Resource not found in state", b.resourceAddress)

		return nil, false
	}

	return resource, true
}

func (b Base) ResourceAddress() string {
	return b.resourceAddress
}
