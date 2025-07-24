// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var _ resourceCRUDInterceptor = identityInterceptor{}

type identityInterceptor struct {
	attributes []inttypes.IdentityAttribute
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

		for _, att := range r.attributes {
			switch att.Name() {
			case names.AttrAccountID:
				diags.Append(identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.AccountID(ctx))...)
				if diags.HasError() {
					return diags
				}

			case names.AttrRegion:
				diags.Append(identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.Region(ctx))...)
				if diags.HasError() {
					return diags
				}

			default:
				var attrVal attr.Value
				diags.Append(response.State.GetAttribute(ctx, path.Root(att.ResourceAttributeName()), &attrVal)...)
				if diags.HasError() {
					return diags
				}

				diags.Append(identity.SetAttribute(ctx, path.Root(att.Name()), attrVal)...)
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
	awsClient := opts.c

	switch response, when := opts.response, opts.when; when {
	case After:
		if response.State.Raw.IsNull() {
			break
		}
		identity := response.Identity
		if identity == nil {
			break
		}

		for _, att := range r.attributes {
			switch att.Name() {
			case names.AttrAccountID:
				diags.Append(identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.AccountID(ctx))...)
				if diags.HasError() {
					return diags
				}

			case names.AttrRegion:
				diags.Append(identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.Region(ctx))...)
				if diags.HasError() {
					return diags
				}

			default:
				var attrVal attr.Value
				diags.Append(response.State.GetAttribute(ctx, path.Root(att.ResourceAttributeName()), &attrVal)...)
				if diags.HasError() {
					return diags
				}

				diags.Append(identity.SetAttribute(ctx, path.Root(att.Name()), attrVal)...)
				if diags.HasError() {
					return diags
				}
			}
		}
	}

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

func newIdentityInterceptor(attributes []inttypes.IdentityAttribute) identityInterceptor {
	return identityInterceptor{
		attributes: attributes,
	}
}
