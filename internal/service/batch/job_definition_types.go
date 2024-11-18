// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// This file is meant to be used as a plan modifier from the
// It provides plan modifiers for comparing JSON strings representing the properties
// Eventually this can be removed with #29817
// DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool
// DiffSuppressOnRefresh: true,
// ValidateFunc:          validJobContainerProperties,

type ecsStringPlanModifier struct{}

func (m *ecsStringPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.StateValue.IsNull() {
		return
	}

	ok, err := equivalentECSPropertiesJSON(req.PlanValue.ValueString(), req.StateValue.ValueString())
	if err == nil && ok {
		resp.PlanValue = req.StateValue
	}
}

func (m *ecsStringPlanModifier) Description(ctx context.Context) string {
	return "compares ecs properties by parsing and comparing their content"
}

func (m *ecsStringPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func ECSStringPlanModifier() planmodifier.String {
	return &ecsStringPlanModifier{}
}

type containerPropertiesStringPlanModifier struct{}

func (m *containerPropertiesStringPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.StateValue.IsNull() {
		return
	}

	ok, err := equivalentContainerPropertiesJSON(req.PlanValue.ValueString(), req.StateValue.ValueString())
	if err == nil && ok {
		// If the query specifications are equivalent, suppress the diff
		// by setting the plan value to value already in state.
		resp.PlanValue = req.StateValue
	}
}

func (m *containerPropertiesStringPlanModifier) Description(ctx context.Context) string {
	return "compares container properties by parsing and comparing their content"
}

func (m *containerPropertiesStringPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func ContainerPropertiesStringPlanModifier() planmodifier.String {
	return &containerPropertiesStringPlanModifier{}
}

type nodePropertiesStringPlanModifier struct{}

func (m *nodePropertiesStringPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.StateValue.IsNull() {
		return
	}

	ok, err := equivalentNodePropertiesJSON(req.PlanValue.ValueString(), req.StateValue.ValueString())
	if err == nil && ok {
		// If the query specifications are equivalent, suppress the diff
		// by setting the plan value to value already in state.
		resp.PlanValue = req.StateValue
	}
}

func (m *nodePropertiesStringPlanModifier) Description(ctx context.Context) string {
	return "compares node properties by parsing and comparing their content"
}

func (m *nodePropertiesStringPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func NodePropertiesStringPlanModifier() planmodifier.String {
	return &nodePropertiesStringPlanModifier{}
}
