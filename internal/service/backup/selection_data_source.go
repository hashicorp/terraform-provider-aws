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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_backup_selection")
func DataSourceSelection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSelectionRead,

		Schema: map[string]*schema.Schema{
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"selection_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrIAMRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResources: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSelectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	input := &backup.GetBackupSelectionInput{
		BackupPlanId: aws.String(d.Get("plan_id").(string)),
		SelectionId:  aws.String(d.Get("selection_id").(string)),
	}

	resp, err := conn.GetBackupSelection(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Backup Selection: %s", err)
	}

	d.SetId(aws.ToString(resp.SelectionId))
	d.Set(names.AttrIAMRoleARN, resp.BackupSelection.IamRoleArn)
	d.Set(names.AttrName, resp.BackupSelection.SelectionName)

	if err := d.Set(names.AttrResources, resp.BackupSelection.Resources); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting resources: %s", err)
	}

	return diags
}
