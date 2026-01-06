// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package stringplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/privatestate"
)

// RequiresReplaceWO returns a plan modifier that forces resource replacement
// if a write-only value changes.
func RequiresReplaceWO(privateStateKey string) planmodifier.String {
	return requiresReplaceWO{
		privateStateKey: privateStateKey,
	}
}

type requiresReplaceWO struct {
	privateStateKey string
}

func (m requiresReplaceWO) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m requiresReplaceWO) MarkdownDescription(context.Context) string {
	return "If the value of this write-only attribute changes, Terraform will destroy and recreate the resource."
}

func (m requiresReplaceWO) PlanModifyString(ctx context.Context, request planmodifier.StringRequest, response *planmodifier.StringResponse) {
	newValue := request.ConfigValue
	newValueExists := !newValue.IsNull()

	woStore := privatestate.NewWriteOnlyValueStore(request.Private, m.privateStateKey)
	oldValueExists, diags := woStore.HasValue(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !newValueExists {
		if oldValueExists {
			response.RequiresReplace = true
		}
		return
	}

	if !oldValueExists {
		response.RequiresReplace = true
		return
	}

	equal, diags := woStore.EqualValue(ctx, newValue)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if !equal {
		response.RequiresReplace = true
	}
}
