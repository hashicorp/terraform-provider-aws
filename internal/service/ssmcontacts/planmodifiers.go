// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts

import (
	"cmp"
	"context"
	"slices"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.List = shiftCoveragesModifier{}

func shiftCoveragesPlanModifier() planmodifier.List {
	return &shiftCoveragesModifier{}
}

type shiftCoveragesModifier struct{}

func (s shiftCoveragesModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
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

	slices.SortFunc(plan, func(a, b shiftCoveragesData) int {
		return cmp.Compare(a.MapBlockKey.ValueString(), b.MapBlockKey.ValueString())
	})

	slices.SortFunc(state, func(a, b shiftCoveragesData) int {
		return cmp.Compare(a.MapBlockKey.ValueString(), b.MapBlockKey.ValueString())
	})

	if gocmp.Equal(plan, state) {
		resp.PlanValue = req.StateValue
	}
}

func (s shiftCoveragesModifier) Description(_ context.Context) string {
	return "Suppress diff for shift_coverages"
}

func (s shiftCoveragesModifier) MarkdownDescription(ctx context.Context) string {
	return s.Description(ctx)
}
