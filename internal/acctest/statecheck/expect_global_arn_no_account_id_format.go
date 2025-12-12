// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
)

var _ statecheck.StateCheck = expectGlobalARNNoAccountIDFormatCheck{}

type expectGlobalARNNoAccountIDFormatCheck struct {
	base          Base
	attributePath tfjsonpath.Path
	arnService    string
	arnFormat     string
}

func (e expectGlobalARNNoAccountIDFormatCheck) CheckState(ctx context.Context, request statecheck.CheckStateRequest, response *statecheck.CheckStateResponse) {
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

	knownCheck := tfknownvalue.GlobalARNNoAccountIDExact(e.arnService, arnString)
	if err = knownCheck.CheckValue(value); err != nil { //nolint:contextcheck // knownCheck implements an interface
		response.Error = fmt.Errorf("checking value for attribute at path: %s.%s, err: %w", e.base.ResourceAddress(), e.attributePath, err)
		return
	}
}

func ExpectGlobalARNNoAccountIDFormat(resourceAddress string, attributePath tfjsonpath.Path, arnService, arnFormat string) statecheck.StateCheck {
	return expectGlobalARNNoAccountIDFormatCheck{
		base:          NewBase(resourceAddress),
		attributePath: attributePath,
		arnService:    arnService,
		arnFormat:     arnFormat,
	}
}
