// Copyright IBM Corp. 2014, 2026
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

func SingleParameterized(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, importSpec *inttypes.FrameworkImport, response *resource.ImportStateResponse) {
	attr := identitySpec.Attributes[len(identitySpec.Attributes)-1]
	identityPath := path.Root(attr.Name())
	resourcePath := path.Root(attr.ResourceAttributeName())

	parameterVal := request.ID

	if identity := request.Identity; request.ID == "" && identity != nil {
		response.Diagnostics.Append(validateAccountID(ctx, identity, client.AccountID(ctx))...)
		if response.Diagnostics.HasError() {
			return
		}

		var parameterAttr types.String
		response.Diagnostics.Append(identity.GetAttribute(ctx, identityPath, &parameterAttr)...)
		if response.Diagnostics.HasError() {
			return
		}
		parameterVal = parameterAttr.ValueString()
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, resourcePath, parameterVal)...)

	if identity := response.Identity; identity != nil {
		response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(names.AttrAccountID), client.AccountID(ctx))...)
		response.Diagnostics.Append(identity.SetAttribute(ctx, identityPath, parameterVal)...)
	}

	if !identitySpec.IsGlobalResource {
		setRegionFromStateOrIdentity(ctx, client, request, response)
	}
}

func MultipleParameterized(ctx context.Context, client AWSClient, request resource.ImportStateRequest, identitySpec *inttypes.Identity, importSpec *inttypes.FrameworkImport, response *resource.ImportStateResponse) {
	if request.ID != "" {
		id, parts, err := importSpec.ImportID.Parse(request.ID)
		if err != nil {
			response.Diagnostics.Append(InvalidResourceImportIDError(
				"could not be parsed.\n\n" +
					fmt.Sprintf("Value: %q\nError: %s", request.ID, err),
			))
			return
		}

		for resourceAttr, val := range parts {
			resourcePath := path.Root(resourceAttr)
			response.Diagnostics.Append(response.State.SetAttribute(ctx, resourcePath, val)...)

			if identity := response.Identity; identity != nil {
				var identityAttr string
				for _, attr := range identitySpec.Attributes {
					if attr.ResourceAttributeName() == resourceAttr {
						identityAttr = attr.Name()
						break
					}
				}
				if identityAttr == "" {
					response.Diagnostics.AddError(
						"Unexpected Error",
						"An unexpected error occurred while importing a resource. "+
							"This is always an error in the provider. "+
							"Please report the following to the provider developer:\n\n"+
							fmt.Sprintf("No Resource Identity mapping found for attribute %q.", resourceAttr),
					)
					return
				}
				identityPath := path.Root(identityAttr)
				response.Diagnostics.Append(identity.SetAttribute(ctx, identityPath, val)...)
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
			switch attr.Name() {
			case names.AttrAccountID, names.AttrRegion:
				// Do nothing

			default:
				identityPath := path.Root(attr.Name())
				resourcePath := path.Root(attr.ResourceAttributeName())

				var parameterAttr types.String
				response.Diagnostics.Append(identity.GetAttribute(ctx, identityPath, &parameterAttr)...)
				if response.Diagnostics.HasError() {
					return
				}
				parameterVal := parameterAttr.ValueString()

				response.Diagnostics.Append(response.State.SetAttribute(ctx, resourcePath, parameterVal)...)

				if identity := response.Identity; identity != nil {
					response.Diagnostics.Append(identity.SetAttribute(ctx, identityPath, parameterVal)...)
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

	if !identitySpec.IsGlobalResource {
		setRegionFromStateOrIdentity(ctx, client, request, response)
	}
}
