// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegionalSingleton(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, response *resource.ImportStateResponse) {
	accountIDPath := path.Root(names.AttrAccountID)
	regionPath := path.Root(names.AttrRegion)

	var regionVal string
	if regionVal = request.ID; regionVal != "" {
		var region types.String
		response.Diagnostics.Append(response.State.GetAttribute(ctx, regionPath, &region)...)
		if response.Diagnostics.HasError() {
			return
		}

		if !region.IsNull() {
			if region.ValueString() != request.ID {
				response.Diagnostics.AddError(
					InvalidResourceImportIDValue,
					fmt.Sprintf("The region passed for import, %q, does not match the region %q in the ID", region.ValueString(), regionVal),
				)
				return
			}
		}
	} else if identity := request.Identity; identity != nil {
		response.Diagnostics.Append(validateAccountID(ctx, identity, client.AccountID(ctx))...)
		if response.Diagnostics.HasError() {
			return
		}

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
	for _, attr := range identitySpec.IdentityDuplicateAttrs {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(attr), regionVal)...)
	}

	accountID := client.AccountID(ctx)

	if identity := response.Identity; identity != nil {
		response.Diagnostics.Append(identity.SetAttribute(ctx, accountIDPath, accountID)...)
		response.Diagnostics.Append(identity.SetAttribute(ctx, regionPath, regionVal)...)
	}
}

func validateAccountID(ctx context.Context, identity *tfsdk.ResourceIdentity, expected string) (diags diag.Diagnostics) {
	accountIDPath := path.Root(names.AttrAccountID)
	var accountIDAttr types.String
	diags.Append(identity.GetAttribute(ctx, accountIDPath, &accountIDAttr)...)
	if diags.HasError() {
		return
	}
	if !accountIDAttr.IsNull() {
		if accountIDAttr.ValueString() != expected {
			diags.AddAttributeError(
				accountIDPath,
				InvalidResourceImportIDValue,
				fmt.Sprintf("Provider configured with Account ID %q cannot be used to import resources from account %q", expected, accountIDAttr.ValueString()),
			)
			return
		}
	}
	return
}
