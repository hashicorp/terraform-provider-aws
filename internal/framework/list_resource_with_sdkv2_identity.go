// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type ListResourceWithSDKv2Identity struct {
	resourceSchema *schema.Resource
	identity       inttypes.Identity
	identitySchema *schema.ResourceIdentity
}

func (l *ListResourceWithSDKv2Identity) SetIdentitySpec(identity inttypes.Identity) {
	var identitySchema schema.ResourceIdentity
	out := make(map[string]*schema.Schema)
	for _, v := range identity.Attributes {
		out[v.Name()] = &schema.Schema{
			Type: schema.TypeString,
		}
		if v.Required() {
			out[v.Name()].Required = true
		} else {
			out[v.Name()].Optional = true
		}
	}

	identitySchema.SchemaFunc = func() map[string]*schema.Schema {
		return out
	}

	l.identitySchema = &identitySchema
	l.resourceSchema.Identity = &identitySchema
	l.identity = identity
}

func (l *ListResourceWithSDKv2Identity) GetResource() *schema.Resource {
	return l.resourceSchema
}

func (l *ListResourceWithSDKv2Identity) SetResource(resource *schema.Resource) {
	// Add region attribute if it does not exit
	if _, ok := resource.SchemaMap()[names.AttrRegion]; !ok {
		resource.SchemaMap()[names.AttrRegion] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		}
	}

	l.resourceSchema = resource
}

func (l *ListResourceWithSDKv2Identity) SetIdentity(ctx context.Context, client *conns.AWSClient, d *schema.ResourceData) error {
	ident, err := d.Identity()
	if err != nil {
		return err
	}

	for _, v := range l.identity.Attributes {
		switch v.Name() {
		case names.AttrRegion:
			if err := ident.Set(v.Name(), client.Region(ctx)); err != nil {
				return err
			}
		case names.AttrAccountID:
			if err := ident.Set(v.Name(), client.AccountID(ctx)); err != nil {
				return err
			}
		default:
			attributeValue := d.Get(v.Name()).(string)
			if err := ident.Set(v.Name(), attributeValue); err != nil {
				return err
			}
		}
	}

	return nil
}
