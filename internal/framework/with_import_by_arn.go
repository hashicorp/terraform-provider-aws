// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// WithImportByARN is intended to be embedded in resources which import state via the "arn" attribute.
// See https://developer.hashicorp.com/terraform/plugin/framework/resources/import.
type WithImportByARN struct{}

func (w *WithImportByARN) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	arnARN, err := arn.Parse(request.ID)
	if err != nil {
		response.Diagnostics.AddError(
			"Invalid Resource Import ID Value",
			"The import ID could not be parsed as an ARN.\n\n"+
				fmt.Sprintf("Value: %q\nError: %s", request.ID, err),
		)
		return
	}

	var region types.String
	response.Diagnostics.Append(response.State.GetAttribute(ctx, path.Root("region"), &region)...)
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
	} else {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("region"), arnARN.Region)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), request, response)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), request.ID)...)
}
