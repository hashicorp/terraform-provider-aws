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

func RegionalSingleParameterized(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, importSpec *inttypes.FrameworkImport, response *resource.ImportStateResponse) {
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

func GlobalSingleParameterized(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, importSpec *inttypes.FrameworkImport, response *resource.ImportStateResponse) {
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

func RegionalMultipleParameterized(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, importSpec *inttypes.FrameworkImport, response *resource.ImportStateResponse) {
	regionPath := path.Root(names.AttrRegion)
	var regionVal string

	if request.ID != "" {
		id, parts, err := importSpec.ImportID.Parse(request.ID)
		if err != nil {
			response.Diagnostics.Append(InvalidResourceImportIDError(
				"could not be parsed.\n\n" +
					fmt.Sprintf("Value: %q\nError: %s", request.ID, err),
			))
			return
		}

		var regionAttr types.String
		response.Diagnostics.Append(response.State.GetAttribute(ctx, regionPath, &regionAttr)...)
		if response.Diagnostics.HasError() {
			return
		}
		regionVal = regionAttr.ValueString()

		for attr, val := range parts {
			attrPath := path.Root(attr)
			response.Diagnostics.Append(response.State.SetAttribute(ctx, attrPath, val)...)

			if identity := response.Identity; identity != nil {
				response.Diagnostics.Append(identity.SetAttribute(ctx, attrPath, val)...)
			}
		}

		if importSpec.SetIDAttr {
			response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), id)...)
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
		response.Diagnostics.Append(response.State.SetAttribute(ctx, regionPath, regionVal)...)

		for _, attr := range identitySpec.Attributes {
			switch attr.Name {
			case names.AttrAccountID, names.AttrRegion:
				// Do nothing

			default:
				attrPath := path.Root(attr.Name)
				var parameterAttr types.String
				response.Diagnostics.Append(identity.GetAttribute(ctx, attrPath, &parameterAttr)...)
				if response.Diagnostics.HasError() {
					return
				}
				parameterVal := parameterAttr.ValueString()

				response.Diagnostics.Append(response.State.SetAttribute(ctx, attrPath, parameterVal)...)

				if identity := response.Identity; identity != nil {
					response.Diagnostics.Append(identity.SetAttribute(ctx, attrPath, parameterVal)...)
				}
			}
		}

		if importSpec.SetIDAttr {
			if idCreator, ok := importSpec.ImportID.(inttypes.FrameworkImportIDCreator); ok {
				response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), idCreator.Create(ctx, response.State))...)
			} else {
				response.Diagnostics.AddError(
					"Unexpected Error",
					"An unexpected error occurred while importing a resource. "+
						"This is always an error in the provider. "+
						"Please report the following to the provider developer:\n\n"+
						"Import ID handler does not implement Creator, but needs to set \"id\" attribute.",
				)
				return
			}
		}
	}

	if identity := response.Identity; identity != nil {
		response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(names.AttrAccountID), client.AccountID(ctx))...)
		response.Diagnostics.Append(identity.SetAttribute(ctx, regionPath, regionVal)...)
	}
}

func GlobalMultipleParameterized(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, importSpec *inttypes.FrameworkImport, response *resource.ImportStateResponse) {
	if request.ID != "" {
		id, parts, err := importSpec.ImportID.Parse(request.ID)
		if err != nil {
			response.Diagnostics.Append(InvalidResourceImportIDError(
				"could not be parsed.\n\n" +
					fmt.Sprintf("Value: %q\nError: %s", request.ID, err),
			))
			return
		}

		for attr, val := range parts {
			attrPath := path.Root(attr)
			response.Diagnostics.Append(response.State.SetAttribute(ctx, attrPath, val)...)

			if identity := response.Identity; identity != nil {
				response.Diagnostics.Append(identity.SetAttribute(ctx, attrPath, val)...)
			}
		}

		if importSpec.SetIDAttr {
			response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), id)...)
		}
	} else if identity := request.Identity; identity != nil {
		response.Diagnostics.Append(validateAccountID(ctx, identity, client.AccountID(ctx))...)
		if response.Diagnostics.HasError() {
			return
		}

		for _, attr := range identitySpec.Attributes {
			switch attr.Name {
			case names.AttrAccountID, names.AttrRegion:
				// Do nothing

			default:
				attrPath := path.Root(attr.Name)
				var parameterAttr types.String
				response.Diagnostics.Append(identity.GetAttribute(ctx, attrPath, &parameterAttr)...)
				if response.Diagnostics.HasError() {
					return
				}
				parameterVal := parameterAttr.ValueString()

				response.Diagnostics.Append(response.State.SetAttribute(ctx, attrPath, parameterVal)...)

				if identity := response.Identity; identity != nil {
					response.Diagnostics.Append(identity.SetAttribute(ctx, attrPath, parameterVal)...)
				}
			}
		}

		if importSpec.SetIDAttr {
			if idCreator, ok := importSpec.ImportID.(inttypes.FrameworkImportIDCreator); ok {
				response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), idCreator.Create(ctx, response.State))...)
			} else {
				response.Diagnostics.AddError(
					"Unexpected Error",
					"An unexpected error occurred while importing a resource. "+
						"This is always an error in the provider. "+
						"Please report the following to the provider developer:\n\n"+
						"Import ID handler does not implement Creator, but needs to set \"id\" attribute.",
				)
				return
			}
		}
	}

	if identity := response.Identity; identity != nil {
		response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(names.AttrAccountID), client.AccountID(ctx))...)
	}
}
