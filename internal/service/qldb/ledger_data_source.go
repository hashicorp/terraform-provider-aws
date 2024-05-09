// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qldb

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_qldb_ledger")
func dataSourceLedger() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLedgerRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"kms_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
				),
			},
			"permissions_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceLedgerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QLDBClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get(names.AttrName).(string)
	ledger, err := findLedgerByName(ctx, conn, name)

	if err != nil {
		return diag.Errorf("reading QLDB Ledger (%s): %s", name, err)
	}

	d.SetId(aws.ToString(ledger.Name))
	d.Set(names.AttrARN, ledger.Arn)
	d.Set("deletion_protection", ledger.DeletionProtection)
	if ledger.EncryptionDescription != nil {
		d.Set("kms_key", ledger.EncryptionDescription.KmsKeyArn)
	} else {
		d.Set("kms_key", nil)
	}
	d.Set(names.AttrName, ledger.Name)
	d.Set("permissions_mode", ledger.PermissionsMode)

	tags, err := listTags(ctx, conn, d.Get(names.AttrARN).(string))

	if err != nil {
		return diag.Errorf("listing tags for QLDB Ledger (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
