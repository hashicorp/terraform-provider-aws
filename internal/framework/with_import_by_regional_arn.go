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

// WithImportByRegionalARN is intended to be embedded in resources which import state via the "arn" attribute.
// See https://developer.hashicorp.com/terraform/plugin/framework/resources/import.
type WithImportByRegionalARN struct {
	importByARN
}

func (w *WithImportByRegionalARN) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	arnARN := w.importState(ctx, request, response)
	if response.Diagnostics.HasError() {
		return
	}

	if request.ID != "" {
		var region types.String
		response.Diagnostics.Append(response.State.GetAttribute(ctx, path.Root(names.AttrRegion), &region)...)
		if response.Diagnostics.HasError() {
			return
		}

		if !region.IsNull() {
			if region.ValueString() != arnARN.Region {
				response.Diagnostics.AddError(
					"Invalid Resource Import ID Value",
					fmt.Sprintf("The region passed for import, %q, does not match the region %q in the ARN %q", region.ValueString(), arnARN.Region, request.ID),
				)
				return
			}
		}
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrRegion), arnARN.Region)...)
	if response.Diagnostics.HasError() {
		return
	}
}
