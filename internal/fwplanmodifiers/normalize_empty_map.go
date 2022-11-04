package fwplanmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type normalizeEmptyMap struct{}

// NormalizeEmptyMap return an AttributePlanModifier that normalizes a missing Map.
// Useful for resolving null vs. empty differences for resource tags.
func NormalizeEmptyMap() tfsdk.AttributePlanModifier {
	return normalizeEmptyMap{}
}

func (m normalizeEmptyMap) Description(context.Context) string {
	return "Resolve differences between null and empty maps"
}

func (m normalizeEmptyMap) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m normalizeEmptyMap) Modify(ctx context.Context, request tfsdk.ModifyAttributePlanRequest, response *tfsdk.ModifyAttributePlanResponse) {
	if request.AttributeState == nil {
		response.AttributePlan = request.AttributePlan

		return
	}

	// If the current value is semantically equivalent to the planned value
	// then return the current value, else return the planned value.

	var planned types.Map

	response.Diagnostics = append(response.Diagnostics, tfsdk.ValueAs(ctx, request.AttributePlan, &planned)...)

	if response.Diagnostics.HasError() {
		return
	}

	var current types.Map

	response.Diagnostics = append(response.Diagnostics, tfsdk.ValueAs(ctx, request.AttributeState, &current)...)

	if response.Diagnostics.HasError() {
		return
	}

	if planned.IsNull() && (current.IsNull() || len(current.Elems) == 0) ||
		(current.IsNull() && (planned.IsNull() || len(planned.Elems) == 0)) {
		response.AttributePlan = request.AttributeState

		return
	}

	response.AttributePlan = request.AttributePlan
}
