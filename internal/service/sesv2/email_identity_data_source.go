// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_sesv2_email_identity")
// @Tags(identifierAttribute="arn")
func DataSourceEmailIdentity() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEmailIdentityRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_set_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dkim_signing_attributes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"current_signing_key_length": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_signing_private_key": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"domain_signing_selector": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_key_generation_timestamp": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"next_signing_key_length": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"signing_attributes_origin": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tokens": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"email_identity": {
				Type:     schema.TypeString,
				Required: true,
			},
			"identity_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"verified_for_sending_status": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

const (
	DSNameEmailIdentity = "Email Identity Data Source"
)

func dataSourceEmailIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SESV2Client(ctx)

	name := d.Get("email_identity").(string)

	out, err := FindEmailIdentityByID(ctx, conn, name)
	if err != nil {
		return create.AppendDiagError(diags, names.SESV2, create.ErrActionReading, DSNameEmailIdentity, name, err)
	}

	arn := emailIdentityNameToARN(meta, name)

	d.SetId(name)
	d.Set(names.AttrARN, arn)
	d.Set("configuration_set_name", out.ConfigurationSetName)
	d.Set("email_identity", name)

	if out.DkimAttributes != nil {
		tfMap := flattenDKIMAttributes(out.DkimAttributes)
		tfMap["domain_signing_private_key"] = d.Get("dkim_signing_attributes.0.domain_signing_private_key").(string)
		tfMap["domain_signing_selector"] = d.Get("dkim_signing_attributes.0.domain_signing_selector").(string)

		if err := d.Set("dkim_signing_attributes", []interface{}{tfMap}); err != nil {
			return create.AppendDiagError(diags, names.SESV2, create.ErrActionSetting, ResNameEmailIdentity, name, err)
		}
	} else {
		d.Set("dkim_signing_attributes", nil)
	}

	d.Set("identity_type", string(out.IdentityType))
	d.Set("verified_for_sending_status", out.VerifiedForSendingStatus)

	return diags
}
