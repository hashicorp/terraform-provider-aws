// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts

import (
	"context"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.List = shiftCoveragesPlanModifier{}

func ShiftCoveragesPlanModifier() planmodifier.List {
	return &shiftCoveragesPlanModifier{}
}

type shiftCoveragesPlanModifier struct{}

func (s shiftCoveragesPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.PlanValue.IsNull() {
		return
	}

	if req.PlanValue.IsUnknown() {
		return
	}

	if req.ConfigValue.IsUnknown() {
		return
	}

	var plan, state []shiftCoveragesData
	resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &plan, false)...)
	resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &state, false)...)

	req.PlanValue.ElementsAs(ctx, &plan, false)
	if resp.Diagnostics.HasError() {
		return
	}

	sort.Slice(plan, func(i, j int) bool {
		return plan[i].MapBlockKey.ValueString() < plan[j].MapBlockKey.ValueString()
	})

	sort.Slice(state, func(i, j int) bool {
		return state[i].MapBlockKey.ValueString() < state[j].MapBlockKey.ValueString()
	})

	isEqual := cmp.Diff(plan, state) == ""

	if isEqual {
		resp.PlanValue = req.StateValue
	}
}

func (s shiftCoveragesPlanModifier) Description(_ context.Context) string {
	return "Suppress diff for shift_coverages"
}

func (s shiftCoveragesPlanModifier) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}
