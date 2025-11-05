// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

type stateValue struct {
	resourceAddress string
	attributePath   tfjsonpath.Path
	value           *string
}

func StateValue() stateValue {
	return stateValue{}
}

// GetStateValue sets the resource address and attribute path to check and stores the state value.
// Calls to GetStateValue occur before any TestStep is run.
func (v *stateValue) GetStateValue(resourceAddress string, attributePath tfjsonpath.Path) statecheck.StateCheck {
	v.resourceAddress = resourceAddress
	v.attributePath = attributePath

	return newStateValueStateChecker(v)
}

// Value checks the stored state value against the provided value.
// Calls to Value occur before any TestStep is run.
func (v *stateValue) Value() knownvalue.Check {
	return newStateValueKnownValueChecker(v)
}

type stateValueStateChecker struct {
	base       Base
	stateValue *stateValue
}

func newStateValueStateChecker(stateValue *stateValue) stateValueStateChecker {
	return stateValueStateChecker{
		base:       NewBase(stateValue.resourceAddress),
		stateValue: stateValue,
	}
}

func (vc stateValueStateChecker) CheckState(ctx context.Context, request statecheck.CheckStateRequest, response *statecheck.CheckStateResponse) {
	resource, ok := vc.base.ResourceFromState(request, response)
	if !ok {
		return
	}

	value, err := tfjsonpath.Traverse(resource.AttributeValues, vc.stateValue.attributePath)
	if err != nil {
		response.Error = err
		return
	}

	stringVal, ok := value.(string)
	if !ok {
		response.Error = fmt.Errorf("expected string value for StateValue check, got: %T", value)
		return
	}

	vc.stateValue.value = &stringVal
}

type stateValueKnownValueChecker struct {
	stateValue *stateValue
}

func newStateValueKnownValueChecker(stateValue *stateValue) stateValueKnownValueChecker {
	return stateValueKnownValueChecker{
		stateValue: stateValue,
	}
}

func (vc stateValueKnownValueChecker) CheckValue(other any) error {
	if vc.stateValue.value == nil {
		return fmt.Errorf("state value has not been set")
	}

	otherVal, ok := other.(string)

	if !ok {
		return fmt.Errorf("expected string value for StateValue check, got: %T", other)
	}

	if otherVal != *vc.stateValue.value {
		return fmt.Errorf("expected value %s for StateValue check, got: %s", *vc.stateValue.value, otherVal)
	}

	return nil
}

func (vc stateValueKnownValueChecker) String() string {
	if vc.stateValue.value == nil {
		return "error: state value has not been set"
	}
	return fmt.Sprintf("%s (from state: %q %q)", *vc.stateValue.value, vc.stateValue.resourceAddress, vc.stateValue.attributePath.String())
}
