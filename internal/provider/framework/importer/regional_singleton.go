// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegionalSingleton(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, response *resource.ImportStateResponse) {
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
					"Invalid Resource Import ID Value",
					fmt.Sprintf("The region passed for import, %q, does not match the region %q in the ID", region.ValueString(), regionVal),
				)
				return
			}
		}
	} else if identity := request.Identity; identity != nil {
		response.Diagnostics.Append(identity.GetAttribute(ctx, regionPath, &regionVal)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, regionPath, regionVal)...)
	for _, attr := range identitySpec.IdentityDuplicateAttrs {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(attr), regionVal)...)
	}

	accountID := client.AccountID(ctx)

	if identity := response.Identity; identity != nil {
		response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(names.AttrAccountID), accountID)...)
		response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(names.AttrRegion), regionVal)...)
	}
}
