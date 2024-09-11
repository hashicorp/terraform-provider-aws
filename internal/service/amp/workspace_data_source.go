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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_prometheus_workspace", name="Workspace")
// @Tags
func dataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWorkspaceRead,

		Schema: map[string]*schema.Schema{
			names.AttrAlias: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"prometheus_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	workspaceID := d.Get("workspace_id").(string)
	workspace, err := findWorkspaceByID(ctx, conn, workspaceID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Prometheus Workspace (%s): %s", workspaceID, err)
	}

	d.SetId(workspaceID)
	d.Set(names.AttrAlias, workspace.Alias)
	d.Set(names.AttrARN, workspace.Arn)
	d.Set(names.AttrCreatedDate, workspace.CreatedAt.Format(time.RFC3339))
	d.Set(names.AttrKMSKeyARN, workspace.KmsKeyArn)
	d.Set("prometheus_endpoint", workspace.PrometheusEndpoint)
	d.Set(names.AttrStatus, workspace.Status.StatusCode)

	setTagsOut(ctx, workspace.Tags)

	return diags
}
