// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package stringplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/privatestate"
)

func RequiresReplaceWO(privateStateKey string) planmodifier.String {
	description := "If the value of this write-only attribute changes, Terraform will destroy and recreate the resource."

	return stringplanmodifier.RequiresReplaceIf(func(ctx context.Context, request planmodifier.StringRequest, response *stringplanmodifier.RequiresReplaceIfFuncResponse) {
		newValue := request.ConfigValue
		newValueExists := !newValue.IsNull()

		woStore := privatestate.NewWriteOnlyValueStore(request.Private, privateStateKey)
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
	}, description, description)
}
