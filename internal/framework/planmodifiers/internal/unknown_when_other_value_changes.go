// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package internal

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type planModifierRequest struct {
	Plan           tfsdk.Plan
	PlanValue      attr.Value
	State          tfsdk.State
	StateValue     attr.Value
	Path           path.Path
	PathExpression path.Expression
}

type planModifierResponse struct {
	Diagnostics diag.Diagnostics
}

func UnknownWhenOtherValueChanges(path path.Path) unknownWhenOtherValueChanges {
	return unknownWhenOtherValueChanges{
		path: path,
	}
}

var (
	_ planmodifier.List   = (*unknownWhenOtherValueChanges)(nil)
	_ planmodifier.String = (*unknownWhenOtherValueChanges)(nil)
)

type unknownWhenOtherValueChanges struct {
	path path.Path
}

func (m unknownWhenOtherValueChanges) Description(context.Context) string {
	return fmt.Sprintf("Uses state for unknown when the value at %[1]q does not change.", m.path)
}

func (m unknownWhenOtherValueChanges) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m unknownWhenOtherValueChanges) PlanModifyList(ctx context.Context, request planmodifier.ListRequest, response *planmodifier.ListResponse) {
	if request.StateValue.IsNull() {
		return
	}

	planModifyRequest := planModifierRequest{
		Plan:           request.Plan,
		PlanValue:      request.PlanValue,
		State:          request.State,
		StateValue:     request.StateValue,
		Path:           request.Path,
		PathExpression: request.PathExpression,
	}
	var planModifyResponse planModifierResponse

	hasChange := m.hasChangeAt(ctx, m.path, planModifyRequest, &planModifyResponse)
	response.Diagnostics.Append(planModifyResponse.Diagnostics...)
	if response.Diagnostics.HasError() {
		return
	}

	if !hasChange {
		response.PlanValue = request.StateValue
	}
}

func (m unknownWhenOtherValueChanges) PlanModifyString(ctx context.Context, request planmodifier.StringRequest, response *planmodifier.StringResponse) {
	if request.StateValue.IsNull() {
		return
	}

	planModifyRequest := planModifierRequest{
		Plan:           request.Plan,
		PlanValue:      request.PlanValue,
		State:          request.State,
		StateValue:     request.StateValue,
		Path:           request.Path,
		PathExpression: request.PathExpression,
	}
	var planModifyResponse planModifierResponse

	hasChange := m.hasChangeAt(ctx, m.path, planModifyRequest, &planModifyResponse)
	response.Diagnostics.Append(planModifyResponse.Diagnostics...)
	if response.Diagnostics.HasError() {
		return
	}

	if !hasChange {
		response.PlanValue = request.StateValue
	}
}

func (m unknownWhenOtherValueChanges) hasChangeAt(ctx context.Context, path path.Path, request planModifierRequest, response *planModifierResponse) bool {
	var planValue, stateValue attr.Value
	response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path, &planValue)...)
	response.Diagnostics.Append(request.State.GetAttribute(ctx, path, &stateValue)...)
	if response.Diagnostics.HasError() {
		return false
	}

	return !planValue.Equal(stateValue)
}
