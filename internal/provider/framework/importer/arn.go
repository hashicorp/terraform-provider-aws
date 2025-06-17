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

func GlobalARN(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, response *resource.ImportStateResponse) {
	importByARN(ctx, client, request, identitySpec, response)
}

func RegionalARN(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, response *resource.ImportStateResponse) {
	arnARN := importByARN(ctx, client, request, identitySpec, response)
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
					"Invalid Resource Import Region Value",
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

func RegionalARNWithGlobalFormat(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, response *resource.ImportStateResponse) {
	importByARN(ctx, client, request, identitySpec, response)
	if response.Diagnostics.HasError() {
		return
	}

	regionPath := path.Root(names.AttrRegion)

	var regionVal string
	if request.ID != "" {
		response.Diagnostics.Append(response.State.GetAttribute(ctx, regionPath, &regionVal)...)
		if response.Diagnostics.HasError() {
			return
		}
	} else if identity := request.Identity; identity != nil {
		response.Diagnostics.Append(identity.GetAttribute(ctx, regionPath, &regionVal)...)
		if response.Diagnostics.HasError() {
			return
		}

		if regionVal == "" {
			response.Diagnostics.Append(response.State.GetAttribute(ctx, regionPath, &regionVal)...)
			if response.Diagnostics.HasError() {
				return
			}
		}
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrRegion), regionVal)...)
	if response.Diagnostics.HasError() {
		return
	}

	if identity := response.Identity; identity != nil {
		response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(names.AttrRegion), regionVal)...)
	}
}

func importByARN(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, response *resource.ImportStateResponse) arn.ARN {
	var (
		arnARN arn.ARN
		arnVal string
	)
	if arnVal = request.ID; arnVal != "" {
		var err error
		arnARN, err = arn.Parse(arnVal)
		if err != nil {
			response.Diagnostics.Append(InvalidResourceImportIDError(
				"could not be parsed as an ARN.\n\n" +
					fmt.Sprintf("Value: %q\nError: %s", arnVal, err),
			))
			return arn.ARN{}
		}
	} else if identity := request.Identity; identity != nil {
		arnPath := path.Root(identitySpec.IdentityAttribute)
		identity.GetAttribute(ctx, arnPath, &arnVal)

		var err error
		arnARN, err = arn.Parse(arnVal)
		if err != nil {
			response.Diagnostics.Append(InvalidIdentityAttributeError(
				arnPath,
				"could not be parsed as an ARN.\n\n"+
					fmt.Sprintf("Value: %q\nError: %s", arnVal, err),
			))
			return arn.ARN{}
		}
	}

	accountID := client.AccountID(ctx)
	if arnARN.AccountID != accountID {
		if request.ID != "" {
			response.Diagnostics.Append(InvalidResourceImportIDError(
				fmt.Sprintf("contains an Account ID %q which does not match the provider's %q.\n\nValue: %q", arnARN.AccountID, accountID, arnVal),
			))
		} else {
			arnPath := path.Root(identitySpec.IdentityAttribute)
			response.Diagnostics.Append(InvalidIdentityAttributeError(
				arnPath,
				fmt.Sprintf("contains an Account ID %q which does not match the provider's %q.\n\nValue: %q", arnARN.AccountID, accountID, arnVal),
			))
		}
		return arn.ARN{}
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
