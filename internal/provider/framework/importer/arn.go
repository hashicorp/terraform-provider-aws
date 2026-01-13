// Copyright IBM Corp. 2014, 2026
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

func ARN(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, _ *inttypes.FrameworkImport, response *resource.ImportStateResponse) {
	arnARN := importByARN(ctx, client, request, identitySpec, response)

	if !identitySpec.IsGlobalResource {
		if identitySpec.IsGlobalARNFormat {
			setRegionFromStateOrIdentity(ctx, client, request, response)
		} else {
			setRegionFromARN(ctx, request, arnARN, response)
		}
	}
}

func setRegionFromARN(ctx context.Context, request resource.ImportStateRequest, arnARN arn.ARN, response *resource.ImportStateResponse) {
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

func setRegionFromStateOrIdentity(ctx context.Context, client AWSClient, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	regionPath := path.Root(names.AttrRegion)

	var regionVal string
	if request.ID != "" {
		var regionAttr types.String
		response.Diagnostics.Append(response.State.GetAttribute(ctx, regionPath, &regionAttr)...)
		if response.Diagnostics.HasError() {
			return
		}
		regionVal = regionAttr.ValueString()
	} else if identity := request.Identity; identity != nil {
		var regionAttr types.String
		response.Diagnostics.Append(identity.GetAttribute(ctx, regionPath, &regionAttr)...)
		if response.Diagnostics.HasError() {
			return
		}

		if !regionAttr.IsNull() {
			regionVal = regionAttr.ValueString()
		} else {
			regionVal = client.Region(ctx)
		}
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, regionPath, regionVal)...)
	if response.Diagnostics.HasError() {
		return
	}

	if identity := response.Identity; identity != nil {
		response.Diagnostics.Append(identity.SetAttribute(ctx, regionPath, regionVal)...)
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
