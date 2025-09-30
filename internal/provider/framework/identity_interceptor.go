// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var _ resourceCRUDInterceptor = identityInterceptor{}

type identityInterceptor struct {
	attributes []inttypes.IdentityAttribute
}

func (r identityInterceptor) create(ctx context.Context, opts interceptorOptions[resource.CreateRequest, resource.CreateResponse]) {
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
				opts.response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.AccountID(ctx))...)
				if opts.response.Diagnostics.HasError() {
					return
				}

			case names.AttrRegion:
				opts.response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.Region(ctx))...)
				if opts.response.Diagnostics.HasError() {
					return
				}

			default:
				var attrVal attr.Value
				opts.response.Diagnostics.Append(response.State.GetAttribute(ctx, path.Root(att.ResourceAttributeName()), &attrVal)...)
				if opts.response.Diagnostics.HasError() {
					return
				}

				opts.response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(att.Name()), attrVal)...)
				if opts.response.Diagnostics.HasError() {
					return
				}
			}
		}
	}
}

func (r identityInterceptor) read(ctx context.Context, opts interceptorOptions[resource.ReadRequest, resource.ReadResponse]) {
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
				opts.response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.AccountID(ctx))...)
				if opts.response.Diagnostics.HasError() {
					return
				}

			case names.AttrRegion:
				opts.response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.Region(ctx))...)
				if opts.response.Diagnostics.HasError() {
					return
				}

			default:
				var attrVal attr.Value
				opts.response.Diagnostics.Append(response.State.GetAttribute(ctx, path.Root(att.ResourceAttributeName()), &attrVal)...)
				if opts.response.Diagnostics.HasError() {
					return
				}

				opts.response.Diagnostics.Append(identity.SetAttribute(ctx, path.Root(att.Name()), attrVal)...)
				if opts.response.Diagnostics.HasError() {
					return
				}
			}
		}
	}
}

func (r identityInterceptor) update(ctx context.Context, opts interceptorOptions[resource.UpdateRequest, resource.UpdateResponse]) {
}

func (r identityInterceptor) delete(ctx context.Context, opts interceptorOptions[resource.DeleteRequest, resource.DeleteResponse]) {
}

func newIdentityInterceptor(attributes []inttypes.IdentityAttribute) identityInterceptor {
	return identityInterceptor{
		attributes: attributes,
	}
}
