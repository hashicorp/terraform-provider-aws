// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_prometheus_workspace")
func DataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"prometheus_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AMPConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	workspaceID := d.Get("workspace_id").(string)
	workspace, err := FindWorkspaceByID(ctx, conn, workspaceID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AMP Workspace (%s): %s", workspaceID, err)
	}

	d.SetId(workspaceID)

	d.Set("alias", workspace.Alias)
	d.Set("arn", workspace.Arn)
	d.Set("created_date", workspace.CreatedAt.Format(time.RFC3339))
	d.Set("prometheus_endpoint", workspace.PrometheusEndpoint)
	d.Set("status", workspace.Status.StatusCode)

	if err := d.Set("tags", KeyValueTags(ctx, workspace.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
