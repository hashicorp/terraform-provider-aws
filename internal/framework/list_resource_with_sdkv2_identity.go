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
	resource           *schema.Resource
	identityAttributes []inttypes.IdentityAttribute
	identity           *schema.ResourceIdentity
}

func (l *ListResourceWithSDKv2Identity) AddIdentity(attrs []inttypes.IdentityAttribute) {
	var identity schema.ResourceIdentity
	out := make(map[string]*schema.Schema)
	for _, v := range attrs {
		out[v.Name()] = &schema.Schema{
			Type: schema.TypeString,
		}
		if v.Required() {
			out[v.Name()].Required = true
		} else {
			out[v.Name()].Optional = true
		}
	}

	identity.SchemaFunc = func() map[string]*schema.Schema {
		return out
	}

	l.identity = &identity
	l.resource.Identity = &identity
	l.identityAttributes = attrs
}

func (l *ListResourceWithSDKv2Identity) GetResource() *schema.Resource {
	return l.resource
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

	l.resource = resource
}

func (l *ListResourceWithSDKv2Identity) SetIdentity(ctx context.Context, client *conns.AWSClient, d *schema.ResourceData) error {
	ident, err := d.Identity()
	if err != nil {
		return err
	}

	for _, v := range l.identityAttributes {
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
