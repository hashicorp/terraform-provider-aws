// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// WithImportByParameterizedIdentity is intended to be embedded in global resources which import state via a parameterized Identity.
// See https://developer.hashicorp.com/terraform/plugin/framework/resources/import.
type WithImportByParameterizedIdentity struct {
	identity inttypes.Identity
}

func (w *WithImportByParameterizedIdentity) SetIdentitySpec(identity inttypes.Identity) {
	w.identity = identity
}

func (w *WithImportByParameterizedIdentity) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	client := importer.Client(ctx)
	if client == nil {
		response.Diagnostics.AddError(
			"Unexpected Error",
			"An unexpected error occurred while importing a resource. "+
				"This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+
				"Importer context was nil.",
		)
		return
	}

	if w.identity.IsSingleParameter {
		if w.identity.IsGlobalResource {
			importer.GlobalSingleParameterized(ctx, client, request, &w.identity, response)
		} else {
			importer.RegionalSingleParameterized(ctx, client, request, &w.identity, response)
		}

		return
	}

	if request.ID != "" {
		if w.identity.IDAttrShadowsAttr != names.AttrID {
			response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(w.identity.IDAttrShadowsAttr), request.ID)...)
		}

		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), request.ID)...)

		return
	}

	if identity := request.Identity; identity != nil {
		for _, attr := range w.identity.Attributes {
			switch attr.Name {
			case names.AttrAccountID:

			case names.AttrRegion:
				regionPath := path.Root(names.AttrRegion)
				var regionVal string
				response.Diagnostics.Append(identity.GetAttribute(ctx, regionPath, &regionVal)...)
				if response.Diagnostics.HasError() {
					return
				}

				response.Diagnostics.Append(response.State.SetAttribute(ctx, regionPath, regionVal)...)
				if response.Diagnostics.HasError() {
					return
				}

			default:
				attrPath := path.Root(attr.Name)
				var attrVal string
				response.Diagnostics.Append(identity.GetAttribute(ctx, attrPath, &attrVal)...)
				if response.Diagnostics.HasError() {
					return
				}

				response.Diagnostics.Append(response.State.SetAttribute(ctx, attrPath, attrVal)...)
				if response.Diagnostics.HasError() {
					return
				}
			}
		}
	}
}
