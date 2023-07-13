// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_backup_vault")
func DataSourceVault() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVaultRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"recovery_points": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceVaultRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	input := &backup.DescribeBackupVaultInput{
		BackupVaultName: aws.String(name),
	}

	resp, err := conn.DescribeBackupVaultWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Backup Vault: %s", err)
	}

	d.SetId(aws.StringValue(resp.BackupVaultName))
	d.Set("arn", resp.BackupVaultArn)
	d.Set("kms_key_arn", resp.EncryptionKeyArn)
	d.Set("name", resp.BackupVaultName)
	d.Set("recovery_points", resp.NumberOfRecoveryPoints)

	tags, err := listTags(ctx, conn, aws.StringValue(resp.BackupVaultArn))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Backup Vault (%s): %s", name, err)
	}
	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
