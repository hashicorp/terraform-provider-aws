// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type ListResourceWithSDKv2Resource struct {
	resourceSchema *schema.Resource
	identity       inttypes.Identity
	identitySchema *schema.ResourceIdentity
}

func (l *ListResourceWithSDKv2Resource) SetIdentitySpec(identity inttypes.Identity) {
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

	identitySchema := schema.ResourceIdentity{
		SchemaFunc: func() map[string]*schema.Schema {
			return out
		},
	}

	l.identitySchema = &identitySchema
	l.resourceSchema.Identity = &identitySchema
	l.identity = identity
}

func (l *ListResourceWithSDKv2Resource) RawV5Schemas(ctx context.Context, _ list.RawV5SchemaRequest, response *list.RawV5SchemaResponse) {
	response.ProtoV5Schema = l.resourceSchema.ProtoSchema(ctx)()
	response.ProtoV5IdentitySchema = l.resourceSchema.ProtoIdentitySchema(ctx)()
}

func (l *ListResourceWithSDKv2Resource) GetResource() *schema.Resource {
	return l.resourceSchema
}

func (l *ListResourceWithSDKv2Resource) SetResource(resource *schema.Resource) {
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

func (l *ListResourceWithSDKv2Resource) SetIdentity(ctx context.Context, client *conns.AWSClient, d *schema.ResourceData) error {
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
