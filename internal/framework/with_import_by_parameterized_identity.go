// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
				// accountIDRaw, ok := identity.GetOk(names.AttrAccountID)
				// if ok {
				// 	accountID, ok := accountIDRaw.(string)
				// 	if !ok {
				// 		return nil, fmt.Errorf("identity attribute %q: expected string, got %T", names.AttrAccountID, accountIDRaw)
				// 	}
				// 	client := meta.(*conns.AWSClient)
				// 	if accountID != client.AccountID(ctx) {
				// 		return nil, fmt.Errorf("Unable to import\n\nidentity attribute %q: Provider configured with Account ID %q, got %q", names.AttrAccountID, client.AccountID(ctx), accountID)
				// 	}
				// }

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
