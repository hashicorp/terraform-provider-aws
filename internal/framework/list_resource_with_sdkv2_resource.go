// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type ListResourceWithSDKv2Resource struct {
	resourceSchema *schema.Resource
	identitySpec   inttypes.Identity
	identitySchema *schema.ResourceIdentity
}

func (l *ListResourceWithSDKv2Resource) SetIdentitySpec(identitySpec inttypes.Identity) {
	out := make(map[string]*schema.Schema)
	for _, v := range identitySpec.Attributes {
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
	l.identitySpec = identitySpec
}

func (l *ListResourceWithSDKv2Resource) RawV5Schemas(ctx context.Context, _ list.RawV5SchemaRequest, response *list.RawV5SchemaResponse) {
	response.ProtoV5Schema = l.resourceSchema.ProtoSchema(ctx)()
	response.ProtoV5IdentitySchema = l.resourceSchema.ProtoIdentitySchema(ctx)()
}

func (l *ListResourceWithSDKv2Resource) SetResourceSchema(resource *schema.Resource) {
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

func (l *ListResourceWithSDKv2Resource) ResourceData() *schema.ResourceData {
	return l.resourceSchema.Data(&terraform.InstanceState{})
}

func (l *ListResourceWithSDKv2Resource) SetIdentity(ctx context.Context, client *conns.AWSClient, d *schema.ResourceData) error {
	identity, err := d.Identity()
	if err != nil {
		return err
	}

	for _, attr := range l.identitySpec.Attributes {
		switch attr.Name() {
		case names.AttrAccountID:
			if err := identity.Set(attr.Name(), client.AccountID(ctx)); err != nil {
				return err
			}

		case names.AttrRegion:
			if err := identity.Set(attr.Name(), client.Region(ctx)); err != nil {
				return err
			}

		default:
			val, ok := getAttributeOk(d, attr.ResourceAttributeName())
			if !ok {
				continue
			}
			if err := identity.Set(attr.Name(), val); err != nil {
				return err
			}
		}
	}

	return nil
}

type resourceData interface {
	Id() string
	GetOk(string) (any, bool)
}

func getAttributeOk(d resourceData, name string) (string, bool) {
	if name == "id" {
		return d.Id(), true
	}
	if v, ok := d.GetOk(name); !ok {
		return "", false
	} else {
		return v.(string), true
	}
}
