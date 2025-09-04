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
	resource           *schema.Resource
	identityAttributes []inttypes.IdentityAttribute
	identity           *schema.ResourceIdentity
}

func (l *ListResourceWithSDKv2Resource) AddIdentity(attrs []inttypes.IdentityAttribute) {
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

	identity := schema.ResourceIdentity{
		SchemaFunc: func() map[string]*schema.Schema {
			return out
		},
	}

	l.identity = &identity
	l.resource.Identity = &identity
	l.identityAttributes = attrs
}

func (l *ListResourceWithSDKv2Resource) RawV5Schemas(ctx context.Context, _ list.RawV5SchemaRequest, response *list.RawV5SchemaResponse) {
	response.ProtoV5Schema = l.resource.ProtoSchema(ctx)()
	response.ProtoV5IdentitySchema = l.resource.ProtoIdentitySchema(ctx)()
}

func (l *ListResourceWithSDKv2Resource) GetResource() *schema.Resource {
	return l.resource
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

	l.resource = resource
}

func (l *ListResourceWithSDKv2Resource) SetIdentity(ctx context.Context, client *conns.AWSClient, d *schema.ResourceData) error {
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
