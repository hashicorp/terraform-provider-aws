// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegionalSingleParameterized(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, response *resource.ImportStateResponse) {
	regionPath := path.Root(names.AttrRegion)
	attrPath := path.Root(identitySpec.IdentityAttribute)

	var (
		parameterVal string
		regionVal    string
	)
	if parameterVal = request.ID; parameterVal != "" {
		var regionAttr types.String
		response.Diagnostics.Append(response.State.GetAttribute(ctx, regionPath, &regionAttr)...)
		if response.Diagnostics.HasError() {
			return
		}
		regionVal = regionAttr.ValueString()
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

		var parameterAttr types.String
		response.Diagnostics.Append(identity.GetAttribute(ctx, attrPath, &parameterAttr)...)
		if response.Diagnostics.HasError() {
			return
		}
		parameterVal = parameterAttr.ValueString()
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, regionPath, regionVal)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, attrPath, parameterVal)...)

	accountID := client.AccountID(ctx)

	if identity := response.Identity; identity != nil {
		response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(names.AttrAccountID), accountID)...)
		response.Diagnostics.Append(identity.SetAttribute(ctx, regionPath, regionVal)...)
		response.Diagnostics.Append(identity.SetAttribute(ctx, attrPath, parameterVal)...)
	}
}

func GlobalSingleParameterized(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, response *resource.ImportStateResponse) {
	attrPath := path.Root(identitySpec.IdentityAttribute)

	parameterVal := request.ID

	if identity := request.Identity; request.ID == "" && identity != nil {
		response.Diagnostics.Append(validateAccountID(ctx, identity, client.AccountID(ctx))...)
		if response.Diagnostics.HasError() {
			return
		}

		var parameterAttr types.String
		response.Diagnostics.Append(identity.GetAttribute(ctx, attrPath, &parameterAttr)...)
		if response.Diagnostics.HasError() {
			return
		}
		parameterVal = parameterAttr.ValueString()
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, attrPath, parameterVal)...)

	accountID := client.AccountID(ctx)

	if identity := response.Identity; identity != nil {
		response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(names.AttrAccountID), accountID)...)
		response.Diagnostics.Append(identity.SetAttribute(ctx, attrPath, parameterVal)...)
	}
}
