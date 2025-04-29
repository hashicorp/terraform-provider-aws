// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var _ interceptor = identityInterceptor{}

type identityInterceptor struct {
	attributes []string
}

func (r identityInterceptor) run(ctx context.Context, opts interceptorOptions) diag.Diagnostics {
	var diags diag.Diagnostics
	awsClient := opts.c

	switch d, when, why := opts.d, opts.when, opts.why; when {
	case After:
		switch why {
		case Create:
			identity, err := d.Identity()
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			for _, attr := range r.attributes {
				switch attr {
				case names.AttrAccountID:
					if err := identity.Set(attr, awsClient.AccountID(ctx)); err != nil {
						return sdkdiag.AppendFromErr(diags, err)
					}

				case names.AttrRegion:
					if err := identity.Set(attr, awsClient.Region(ctx)); err != nil {
						return sdkdiag.AppendFromErr(diags, err)
					}

				default:
					val, ok := r.getAttributeOk(d, attr)
					if !ok {
						continue
					}
					if err := identity.Set(attr, val); err != nil {
						return sdkdiag.AppendFromErr(diags, err)
					}
				}
			}
		}
	}

	return diags
}

func (r identityInterceptor) getAttributeOk(d schemaResourceData, name string) (string, bool) {
	if name == "id" {
		return d.Id(), true
	}
	v, ok := d.GetOk(name)
	return v.(string), ok
}

func newIdentityInterceptor(attributes []types.IdentityAttribute) interceptorItem {
	return interceptorItem{
		when: After,
		why:  Create, // TODO: probably need to do this after Read and Update as well
		interceptor: identityInterceptor{
			attributes: tfslices.ApplyToAll(attributes, func(v types.IdentityAttribute) string {
				return v.Name
			}),
		},
	}
}

func newResourceIdentity(v types.Identity) *schema.ResourceIdentity {
	return &schema.ResourceIdentity{
		SchemaFunc: func() map[string]*schema.Schema {
			return newIdentitySchema(v.Attributes)
		},
	}
}

func newIdentitySchema(attributes []types.IdentityAttribute) map[string]*schema.Schema {
	identitySchema := make(map[string]*schema.Schema, len(attributes))
	for _, attr := range attributes {
		identitySchema[attr.Name] = newIdentityAttribute(attr)
	}
	return identitySchema
}

func newIdentityAttribute(attribute types.IdentityAttribute) *schema.Schema {
	attr := &schema.Schema{
		Type: schema.TypeString,
	}
	if attribute.Required {
		attr.RequiredForImport = true
	} else {
		attr.OptionalForImport = true
	}
	return attr
}
