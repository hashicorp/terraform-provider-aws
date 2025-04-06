// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var _ resourceInterceptor = identityInterceptor{}

type identityInterceptor struct {
	attributes []string
}

func (r identityInterceptor) create(ctx context.Context, opts interceptorOptions[resource.CreateRequest, resource.CreateResponse]) diag.Diagnostics {
	var diags diag.Diagnostics
	awsClient := opts.c

	switch response, when := opts.response, opts.when; when {
	case After:
		identity := response.Identity
		if identity == nil {
			break
		}

		for _, attr := range r.attributes {
			switch attr {
			case names.AttrAccountID:
				diags.Append(identity.SetAttribute(ctx, path.Root(attr), awsClient.AccountID(ctx))...)
				if diags.HasError() {
					return diags
				}

			case names.AttrRegion:
				diags.Append(identity.SetAttribute(ctx, path.Root(attr), awsClient.Region(ctx))...)
				if diags.HasError() {
					return diags
				}

			default:
				var attrVal types.String
				diags.Append(response.State.GetAttribute(ctx, path.Root(attr), &attrVal)...)
				if diags.HasError() {
					return diags
				}

				diags.Append(identity.SetAttribute(ctx, path.Root(attr), attrVal)...)
				if diags.HasError() {
					return diags
				}
			}
		}
	}

	return diags
}

func (r identityInterceptor) read(ctx context.Context, opts interceptorOptions[resource.ReadRequest, resource.ReadResponse]) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func (r identityInterceptor) update(ctx context.Context, opts interceptorOptions[resource.UpdateRequest, resource.UpdateResponse]) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func (r identityInterceptor) delete(ctx context.Context, opts interceptorOptions[resource.DeleteRequest, resource.DeleteResponse]) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func newIdentityInterceptor(attributes []itypes.IdentityAttribute) resourceInterceptor {
	return identityInterceptor{
		attributes: tfslices.ApplyToAll(attributes, func(v itypes.IdentityAttribute) string {
			return v.Name
		}),
	}
}
