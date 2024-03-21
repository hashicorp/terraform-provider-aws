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

// @SDKDataSource("aws_backup_plan")
func DataSourcePlan() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePlanRead,

		Schema: map[string]*schema.Schema{
			"plan_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	if v, ok := d.GetOk("plan_id"); ok {
		resp, err := conn.GetBackupPlanWithContext(ctx, &backup.GetBackupPlanInput{
			BackupPlanId: aws.String(v.(string)),
		})

		d.SetId(aws.StringValue(resp.BackupPlanId))
		d.Set("arn", resp.BackupPlanArn)
		d.Set("name", resp.BackupPlan.BackupPlanName)
		d.Set("version", resp.VersionId)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "getting Backup Plan: %s", err)
		}

		tags, err := listTags(ctx, conn, aws.StringValue(resp.BackupPlanArn))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing tags for Backup Plan (%s): %s", v, err)
		}
		if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
		}
	} else if v, ok := d.GetOk("name"); ok {
		resp, err := FindPlanByName(ctx, conn, v.(string))

		d.SetId(aws.StringValue(resp.BackupPlanId))
		d.Set("arn", resp.BackupPlanArn)
		d.Set("name", resp.BackupPlan.BackupPlanName)
		d.Set("version", resp.VersionId)
		d.Set("plan_id", aws.StringValue(resp.BackupPlanId))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "Getting Backup Plan: %s", err)
		}

		tags, err := listTags(ctx, conn, aws.StringValue(resp.BackupPlanArn))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing tags for Backup Plan (%s): %s", v, err)
		}
		if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
		}
	} else {
		return sdkdiag.AppendErrorf(diags, "plan_id or name must be set")
	}

	return diags
}
