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

var _ statecheck.StateCheck = expectRegionalARNFormatCheck{}

type expectRegionalARNFormatCheck struct {
	base          Base
	attributePath tfjsonpath.Path
	arnService    string
	arnFormat     string
	checkFactory  func(service string, arn string) knownvalue.Check
}

func (e expectRegionalARNFormatCheck) CheckState(ctx context.Context, request statecheck.CheckStateRequest, response *statecheck.CheckStateResponse) {
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

	knownCheck := e.checkFactory(e.arnService, arnString)
	if err = knownCheck.CheckValue(value); err != nil {
		response.Error = fmt.Errorf("checking value for attribute at path: %s.%s, err: %s", e.base.ResourceAddress(), e.attributePath, err)
		return
	}
}

func ExpectRegionalARNFormat(resourceAddress string, attributePath tfjsonpath.Path, arnService, arnFormat string) statecheck.StateCheck {
	return expectRegionalARNFormatCheck{
		base:          NewBase(resourceAddress),
		attributePath: attributePath,
		arnService:    arnService,
		arnFormat:     arnFormat,
		checkFactory: func(service string, arn string) knownvalue.Check {
			return tfknownvalue.RegionalARNExact(service, arn)
		},
	}
}

func ExpectRegionalARNAlternateRegionFormat(resourceAddress string, attributePath tfjsonpath.Path, arnService, arnFormat string) statecheck.StateCheck {
	return expectRegionalARNFormatCheck{
		base:          NewBase(resourceAddress),
		attributePath: attributePath,
		arnService:    arnService,
		arnFormat:     arnFormat,
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
