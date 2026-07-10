// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package listplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// ErrorIfSingleBlockRemoved returns a plan modifier that errors if a previously
// configured single block is removed.
func ErrorIfSingleBlockRemoved() planmodifier.List {
	return errorIfSingleBlockRemoved{}
}

type errorIfSingleBlockRemoved struct{}

func (m errorIfSingleBlockRemoved) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m errorIfSingleBlockRemoved) MarkdownDescription(context.Context) string {
	return "Disallow removing previously configured block."
}

func (m errorIfSingleBlockRemoved) PlanModifyList(ctx context.Context, request planmodifier.ListRequest, response *planmodifier.ListResponse) {
	// Skip create or destroy.
	if request.State.Raw.IsNull() || request.Plan.Raw.IsNull() {
		return
	}

	// Do nothing if there is not a known planned value.
	if !request.PlanValue.IsUnknown() {
		return
	}

	if len(request.StateValue.Elements()) == 1 && len(request.PlanValue.Elements()) == 0 {
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			request.Path,
			"Invalid Block Removal",
			"Removing the previously configured block is not allowed. Re-add the block or recreate the resource manually if you truly intend to remove it.",
		))
	}

}
