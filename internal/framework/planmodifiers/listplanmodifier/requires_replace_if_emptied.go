// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package listplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// RequiresReplaceIfEmptied returns a plan modifier requires resource replacement if:
//   - The resource is planned for update.
//   - The plan and state values are not equal.
//   - The state value is a non-empty list.
//   - The plan value is an empty list.
var RequiresReplaceIfEmptied planmodifier.List = listplanmodifier.RequiresReplaceIf(func(ctx context.Context, request planmodifier.ListRequest, response *listplanmodifier.RequiresReplaceIfFuncResponse) {
	// Do nothing if there is an unknown planned value.
	if request.PlanValue.IsUnknown() {
		return
	}

	if request.StateValue.Length(basetypes.CollectionLengthOptions{UnhandledNullAsZero: true}) > 0 && request.PlanValue.Length(basetypes.CollectionLengthOptions{UnhandledNullAsZero: true}) == 0 {
		response.RequiresReplace = true
	}
}, requiresReplaceIfEmptiedDescription, requiresReplaceIfEmptiedDescription)

const requiresReplaceIfEmptiedDescription = `If the value of this attribute changes from a non-empty list to an empty list, Terraform will destroy and recreate the resource.`
