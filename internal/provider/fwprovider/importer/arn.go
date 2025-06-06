// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TODO: Testing is currently handled in `internal/framework`

func GlobalARN(ctx context.Context, request resource.ImportStateRequest, identitySpec *inttypes.Identity, response *resource.ImportStateResponse) {
	importByARN(ctx, request, identitySpec, response)
}

func RegionalARN(ctx context.Context, request resource.ImportStateRequest, identitySpec *inttypes.Identity, response *resource.ImportStateResponse) {
	arnARN := importByARN(ctx, request, identitySpec, response)
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

func importByARN(ctx context.Context, request resource.ImportStateRequest, identitySpec *inttypes.Identity, response *resource.ImportStateResponse) arn.ARN {
	var (
		arnARN arn.ARN
		arnVal string
	)
	if arnVal = request.ID; arnVal != "" {
		var err error
		arnARN, err = arn.Parse(arnVal)
		if err != nil {
			response.Diagnostics.AddError(
				"Invalid Resource Import ID Value",
				"The import ID could not be parsed as an ARN.\n\n"+
					fmt.Sprintf("Value: %q\nError: %s", arnVal, err),
			)
			return arn.ARN{}
		}
	} else if identity := request.Identity; identity != nil {
		arnPath := path.Root(identitySpec.IdentityAttribute)
		identity.GetAttribute(ctx, arnPath, &arnVal)

		var err error
		arnARN, err = arn.Parse(arnVal)
		if err != nil {
			response.Diagnostics.AddAttributeError(
				arnPath,
				"Invalid Import Attribute Value",
				fmt.Sprintf("Import attribute %q is not a valid ARN, got: %s", arnPath, arnVal),
			)
			return arn.ARN{}
		}
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(identitySpec.IdentityAttribute), arnVal)...)
	for _, attr := range identitySpec.IdentityDuplicateAttrs {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(attr), arnVal)...)
	}

	if identity := response.Identity; identity != nil {
		response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(identitySpec.IdentityAttribute), arnVal)...)
	}

	return arnARN
}
