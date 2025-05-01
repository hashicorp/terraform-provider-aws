// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var _ crudInterceptor = identityInterceptor{}

type identityInterceptor struct {
	attributes []string
}

func (r identityInterceptor) run(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
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
					val, ok := getAttributeOk(d, attr)
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

func getAttributeOk(d schemaResourceData, name string) (string, bool) {
	if name == "id" {
		return d.Id(), true
	}
	v, ok := d.GetOk(name)
	return v.(string), ok
}

func newIdentityInterceptor(attributes []inttypes.IdentityAttribute) interceptorInvocation {
	return interceptorInvocation{
		when: After,
		why:  Create, // TODO: probably need to do this after Read and Update as well
		interceptor: identityInterceptor{
			attributes: tfslices.ApplyToAll(attributes, func(v inttypes.IdentityAttribute) string {
				return v.Name
			}),
		},
	}
}

func newResourceIdentity(v inttypes.Identity) *schema.ResourceIdentity {
	return &schema.ResourceIdentity{
		SchemaFunc: func() map[string]*schema.Schema {
			return newIdentitySchema(v.Attributes)
		},
	}
}

func newIdentitySchema(attributes []inttypes.IdentityAttribute) map[string]*schema.Schema {
	identitySchema := make(map[string]*schema.Schema, len(attributes))
	for _, attr := range attributes {
		identitySchema[attr.Name] = newIdentityAttribute(attr)
	}
	return identitySchema
}

func newIdentityAttribute(attribute inttypes.IdentityAttribute) *schema.Schema {
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

func newIdentityImporter(v inttypes.Identity) *schema.ResourceImporter {
	importer := &schema.ResourceImporter{
		StateContext: func(ctx context.Context, rd *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
			if rd.Id() != "" {
				if v.IDAttrShadowsAttr != "id" {
					rd.Set(v.IDAttrShadowsAttr, rd.Id())
				}
				return []*schema.ResourceData{rd}, nil
			}

			identity, err := rd.Identity()
			if err != nil {
				return nil, err
			}

			for _, attr := range v.Attributes {
				var val string
				switch attr.Name {
				case names.AttrAccountID:
					// TODO: validate this is the correct account
					accountIDRaw, ok := identity.GetOk(names.AttrAccountID)
					if ok {
						val, ok = accountIDRaw.(string)
						if !ok {
							return nil, fmt.Errorf("identity attribute %q: expected string, got %T", names.AttrAccountID, accountIDRaw)
						}
						if v.IDAttrShadowsAttr == names.AttrAccountID {
							rd.Set(names.AttrAccountID, val)
						}
					}

				case names.AttrRegion:
					regionRaw, ok := identity.GetOk(names.AttrRegion)
					if ok {
						val, ok = regionRaw.(string)
						if !ok {
							return nil, fmt.Errorf("identity attribute %q: expected string, got %T", names.AttrRegion, regionRaw)
						}
						rd.Set(names.AttrRegion, val)
					}

				default:
					valRaw, ok := identity.GetOk(attr.Name)
					if attr.Required && !ok {
						return nil, fmt.Errorf("identity attribute %q is required", attr.Name)
					}
					val, ok = valRaw.(string)
					if !ok {
						return nil, fmt.Errorf("identity attribute %q: expected string, got %T", attr.Name, valRaw)
					}
					setAttribute(rd, attr.Name, val)
				}

				if attr.Name == v.IDAttrShadowsAttr {
					rd.SetId(val)
				}
			}

			return []*schema.ResourceData{rd}, nil
		},
	}
	return importer
}

func setAttribute(d *schema.ResourceData, name, value string) {
	if name == "id" {
		d.SetId(value)
	} else {
		d.Set(name, value)
	}
}
