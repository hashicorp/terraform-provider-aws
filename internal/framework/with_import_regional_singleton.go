// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// WithImportRegionalSingleton is intended to be embedded in resources which import state via the "region" attribute.
// See https://developer.hashicorp.com/terraform/plugin/framework/resources/import.
type WithImportRegionalSingleton struct{}

func (w *WithImportRegionalSingleton) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	var region types.String
	response.Diagnostics.Append(response.State.GetAttribute(ctx, path.Root("region"), &region)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !region.IsNull() {
		if region.ValueString() != request.ID {
			response.Diagnostics.AddError(
				"Invalid Resource Import ID Value",
				fmt.Sprintf("The region passed for import, %q, does not match the region %q in the ID", region.ValueString(), request.ID),
			)
			return
		}
	} else {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("region"), request.ID)...)
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), request.ID)...)
}
