// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_backup_selection", name="Selection")
func dataSourceSelection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSelectionRead,

		Schema: map[string]*schema.Schema{
			names.AttrIAMRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrResources: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"selection_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceSelectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)

	planID, selectionID := d.Get("plan_id").(string), d.Get("selection_id").(string)
	output, err := findSelectionByTwoPartKey(ctx, conn, planID, selectionID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Selection (%s): %s", selectionID, err)
	}

	d.SetId(selectionID)
	d.Set(names.AttrIAMRoleARN, output.IamRoleArn)
	d.Set(names.AttrName, output.SelectionName)
	d.Set(names.AttrResources, output.Resources)

	return diags
}
