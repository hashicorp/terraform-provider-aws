// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_backup_vault")
func DataSourceVault() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVaultRead,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyARN: {
				Type:     schema.TypeString,
				Computed: true,
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
	input := &backup.DescribeBackupVaultInput{
		BackupVaultName: aws.String(name),
	}

	resp, err := conn.DescribeBackupVault(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Backup Vault: %s", err)
	}

	d.SetId(aws.ToString(resp.BackupVaultName))
	d.Set(names.AttrARN, resp.BackupVaultArn)
	d.Set(names.AttrKMSKeyARN, resp.EncryptionKeyArn)
	d.Set(names.AttrName, resp.BackupVaultName)
	d.Set("recovery_points", resp.NumberOfRecoveryPoints)

	tags, err := listTags(ctx, conn, aws.ToString(resp.BackupVaultArn))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Backup Vault (%s): %s", name, err)
	}
	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
