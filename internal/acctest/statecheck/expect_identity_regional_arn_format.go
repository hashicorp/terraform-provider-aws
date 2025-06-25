// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
)

var _ statecheck.StateCheck = expectIdentityRegionalARNFormatCheck{}

type expectIdentityRegionalARNFormatCheck struct {
	base         Base
	arnService   string
	arnFormat    string
	checkFactory func(service string, arn string) knownvalue.Check
}

func (e expectIdentityRegionalARNFormatCheck) CheckState(ctx context.Context, request statecheck.CheckStateRequest, response *statecheck.CheckStateResponse) {
	resource, ok := e.base.ResourceFromState(request, response)
	if !ok {
		return
	}

	if resource.IdentitySchemaVersion == nil || len(resource.IdentityValues) == 0 {
		response.Error = fmt.Errorf("%s - Identity not found in state. Either the resource does not support identity or the Terraform version running the test does not support identity. (must be v1.12+)", e.base.ResourceAddress())
		return
	}

	if len(resource.IdentityValues) > 1 {
		deltaMsg := createDeltaString(resource.IdentityValues, map[string]bool{"arn": true}, "actual identity has extra attribute(s): ")

		response.Error = fmt.Errorf("%s - Expected %d attribute(s) in the actual identity object, got %d attribute(s): %s", e.base.ResourceAddress(), 1, len(resource.IdentityValues), deltaMsg)
		return
	}

	attrPath := tfjsonpath.New("arn")
	value, err := tfjsonpath.Traverse(resource.AttributeValues, attrPath)
	if err != nil {
		response.Error = err
		return
	}

	arnString, err := populateFromResourceState(e.arnFormat, resource)
	if err != nil {
		response.Error = err
		return
	}

	knownCheck := e.checkFactory(e.arnService, arnString)
	if err = knownCheck.CheckValue(value); err != nil {
		response.Error = fmt.Errorf("checking value for attribute at path: %s.%s, err: %s", e.base.ResourceAddress(), attrPath, err)
		return
	}
}

func ExpectIdentityRegionalARNFormat(resourceAddress string, arnService, arnFormat string) statecheck.StateCheck {
	return expectIdentityRegionalARNFormatCheck{
		base:       NewBase(resourceAddress),
		arnService: arnService,
		arnFormat:  arnFormat,
		checkFactory: func(service string, arn string) knownvalue.Check {
			return tfknownvalue.RegionalARNExact(service, arn)
		},
	}
}

func ExpectIdentityRegionalARNAlternateRegionFormat(resourceAddress string, arnService, arnFormat string) statecheck.StateCheck {
	return expectIdentityRegionalARNFormatCheck{
		base:       NewBase(resourceAddress),
		arnService: arnService,
		arnFormat:  arnFormat,
		checkFactory: func(service string, arn string) knownvalue.Check {
			return tfknownvalue.RegionalARNAlternateRegionExact(service, arn)
		},
	}
}

// createDeltaString prints the map keys that are present in mapA and not present in mapB
func createDeltaString[T any, V any](mapA map[string]T, mapB map[string]V, msgPrefix string) string {
	deltaMsg := ""

	deltaMap := make(map[string]T, len(mapA))
	maps.Copy(deltaMap, mapA)
	for key := range mapB {
		delete(deltaMap, key)
	}

	deltaKeys := slices.Sorted(maps.Keys(deltaMap))

	for i, k := range deltaKeys {
		if i == 0 {
			deltaMsg += msgPrefix
		} else {
			deltaMsg += ", "
		}
		deltaMsg += fmt.Sprintf("%q", k)
	}

	return deltaMsg
}
