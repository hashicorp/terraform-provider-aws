// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssmcontacts_contact")
func DataSourceContact() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceContactRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameContact = "Contact Data Source"
)

func dataSourceContactRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	arn := d.Get("arn").(string)

	out, err := findContactByID(ctx, conn, arn)
	if err != nil {
		return create.DiagError(names.SSMContacts, create.ErrActionReading, DSNameContact, arn, err)
	}

	d.SetId(aws.ToString(out.ContactArn))

	if err := setContactResourceData(d, out); err != nil {
		return create.DiagError(names.SSMContacts, create.ErrActionSetting, DSNameContact, d.Id(), err)
	}

	tags, err := listTags(ctx, conn, d.Id())
	if err != nil {
		return create.DiagError(names.SSMContacts, create.ErrActionReading, DSNameContact, d.Id(), err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	//lintignore:AWSR002
	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.DiagError(names.SSMContacts, create.ErrActionSetting, DSNameContact, d.Id(), err)
	}

	return nil
}
