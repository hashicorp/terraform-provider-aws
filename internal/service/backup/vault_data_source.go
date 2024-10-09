// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_backup_vault", name="Vault")
func dataSourceVault() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVaultRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"recovery_points": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceVaultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get(names.AttrName).(string)
	output, err := findBackupVaultByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Vault (%s): %s", name, err)
	}

	d.SetId(name)
	d.Set(names.AttrARN, output.BackupVaultArn)
	d.Set(names.AttrKMSKeyARN, output.EncryptionKeyArn)
	d.Set(names.AttrName, output.BackupVaultName)
	d.Set("recovery_points", output.NumberOfRecoveryPoints)

	tags, err := listTags(ctx, conn, d.Get(names.AttrARN).(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Backup Vault (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
