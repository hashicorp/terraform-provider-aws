// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	case OnError:
		identity := response.Identity
		if identity == nil {
			break
		}

		var diags diag.Diagnostics

	identityLoop:
		for _, att := range r.attributes {
			switch att.Name() {
			case names.AttrAccountID:
				diags.Append(identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.AccountID(ctx))...)
				if diags.HasError() {
					break identityLoop
				}

			case names.AttrRegion:
				diags.Append(identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.Region(ctx))...)
				if diags.HasError() {
					break identityLoop
				}

			default:
				var attrVal attr.Value
				diags.Append(response.State.GetAttribute(ctx, path.Root(att.ResourceAttributeName()), &attrVal)...)
				if diags.HasError() {
					break identityLoop
				}

				diags.Append(identity.SetAttribute(ctx, path.Root(att.Name()), attrVal)...)
				if diags.HasError() {
					break identityLoop
				}
			}
		}

		if diags.HasError() {
			response.Identity = nil
		}

		opts.response.Diagnostics.Append(diags...)
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
	case OnError:
		if response.State.Raw.IsNull() {
			break
		}
		identity := response.Identity
		if identity == nil {
			break
		}

		if identityIsFullyNull(ctx, identity, r.attributes) {
			var diags diag.Diagnostics

		identityLoop:
			for _, att := range r.attributes {
				switch att.Name() {
				case names.AttrAccountID:
					diags.Append(identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.AccountID(ctx))...)
					if diags.HasError() {
						break identityLoop
					}

				case names.AttrRegion:
					diags.Append(identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.Region(ctx))...)
					if diags.HasError() {
						break identityLoop
					}

				default:
					var attrVal attr.Value
					diags.Append(response.State.GetAttribute(ctx, path.Root(att.ResourceAttributeName()), &attrVal)...)
					if diags.HasError() {
						break identityLoop
					}

					diags.Append(identity.SetAttribute(ctx, path.Root(att.Name()), attrVal)...)
					if diags.HasError() {
						break identityLoop
					}
				}

				if diags.HasError() {
					response.Identity = nil
				}

				opts.response.Diagnostics.Append(diags...)
			}
		}
	}
}

func (r identityInterceptor) delete(ctx context.Context, opts interceptorOptions[resource.DeleteRequest, resource.DeleteResponse]) {
}

// identityIsFullyNull returns true if a resource supports identity and
// all attributes are set to null values
func identityIsFullyNull(ctx context.Context, identity *tfsdk.ResourceIdentity, attributes []inttypes.IdentityAttribute) bool {
	if identity == nil {
		return true
	}

	for _, attr := range attributes {
		var attrVal types.String
		if diags := identity.GetAttribute(ctx, path.Root(attr.Name()), &attrVal); diags.HasError() {
			return false
		}
		if !attrVal.IsNull() && attrVal.ValueString() != "" {
			return false
		}
	}

	return true
}

func newIdentityInterceptor(attributes []inttypes.IdentityAttribute) identityInterceptor {
	return identityInterceptor{
		attributes: attributes,
	}
}
