// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package setplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// RequiresReplaceIfElementsDeleted returns a plan modifier that requires resource replacement if:
//   - The resource is planned for update.
//   - The plan and state values are not equal.
//   - The state value contains elements that are not present in the plan value.
var RequiresReplaceIfElementsDeleted planmodifier.Set = setplanmodifier.RequiresReplaceIf(func(ctx context.Context, request planmodifier.SetRequest, response *setplanmodifier.RequiresReplaceIfFuncResponse) {
	// Do nothing if there is an unknown planned value.
	if request.PlanValue.IsUnknown() {
		return
	}

	diff, diags := fwflex.SetDifference(ctx, request.StateValue, request.PlanValue)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0 {
		response.RequiresReplace = true
	}
}, requiresReplaceIfElementsDeletedDescription, requiresReplaceIfElementsDeletedDescription)

const requiresReplaceIfElementsDeletedDescription = `If the previous value of this attribute contains elements that are not present in the planned value, Terraform will destroy and recreate the resource.`
