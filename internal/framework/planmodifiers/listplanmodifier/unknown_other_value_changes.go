// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package listplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func UnknownWhenOtherValueChanges(path path.Path) planmodifier.List {
	return unknownWhenOtherValueChanges{
		path: path,
	}
}

type unknownWhenOtherValueChanges struct {
	path path.Path
}

func (m unknownWhenOtherValueChanges) Description(_ context.Context) string {
	return fmt.Sprintf("Uses state for unknown when the value at %[1]q does not change.", m.path)
}

func (m unknownWhenOtherValueChanges) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m unknownWhenOtherValueChanges) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.StateValue.IsNull() {
		return
	}

	var otherPlanValue, otherStateValue attr.Value
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, m.path, &otherPlanValue)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, m.path, &otherStateValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !otherPlanValue.Equal(otherStateValue) {
		return
	}

	resp.PlanValue = req.StateValue
}
