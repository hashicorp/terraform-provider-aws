// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

var _ statecheck.StateCheck = expectGlobalARNFormatCheck{}

type expectGlobalARNFormatCheck struct {
	base          Base
	attributePath tfjsonpath.Path
	arnService    string
	arnFormat     string
}

func (e expectGlobalARNFormatCheck) CheckState(ctx context.Context, request statecheck.CheckStateRequest, response *statecheck.CheckStateResponse) {
	resource, ok := e.base.ResourceFromState(request, response)
	if !ok {
		return
	}

	value, err := tfjsonpath.Traverse(resource.AttributeValues, e.attributePath)
	if err != nil {
		response.Error = err
		return
	}

	arnString, err := populateFromResourceState(e.arnFormat, resource)
	if err != nil {
		response.Error = err
		return
	}

	knownCheck := acctest.GlobalARN(e.arnService, arnString)
	if err = knownCheck.CheckValue(value); err != nil { //nolint:contextcheck // knownCheck implements an interface
		response.Error = fmt.Errorf("checking value for attribute at path: %s.%s, err: %s", e.base.ResourceAddress(), e.attributePath, err)
		return
	}
}

func ExpectGlobalARNFormat(resourceAddress string, attributePath tfjsonpath.Path, arnService, arnFormat string) statecheck.StateCheck {
	return expectGlobalARNFormatCheck{
		base:          NewBase(resourceAddress),
		attributePath: attributePath,
		arnService:    arnService,
		arnFormat:     arnFormat,
	}
}
