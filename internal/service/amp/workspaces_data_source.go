// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKDataSource("aws_prometheus_workspaces")
func DataSourceWorkspaces() *schema.Resource { // nosemgrep:ci.caps0-in-func-name
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWorkspacesRead,

		Schema: map[string]*schema.Schema{
			"alias_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"aliases": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"workspace_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceWorkspacesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.caps0-in-func-name
	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	alias_prefix := d.Get("alias_prefix").(string)
	workspaces, err := FindWorkspaces(ctx, conn, alias_prefix)

	if err != nil {
		return diag.Errorf("reading AMP Workspaces: %s", err)
	}

	var arns, aliases, workspace_ids []string
	for _, w := range workspaces {
		arns = append(arns, aws.StringValue(w.Arn))
		aliases = append(aliases, aws.StringValue(w.Alias))
		workspace_ids = append(workspace_ids, aws.StringValue(w.WorkspaceId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("aliases", aliases)
	d.Set("arns", arns)
	d.Set("workspace_ids", workspace_ids)

	return nil
}
