// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	"github.com/aws/aws-sdk-go-v2/service/grafana/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_grafana_workspaces", name="Workspaces")
func dataSourceWorkspaces() *schema.Resource { // nosemgrep:ci.caps0-in-func-name
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWorkspacesRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"names": {
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

func dataSourceWorkspacesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics { // nosemgrep:ci.caps0-in-func-name
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	name := d.Get("name").(string)
	workspaces, err := findWorkspaces(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Prometheus Workspaces: %s", err)
	}

	var names, workspaceIDs []string
	for _, w := range workspaces {
		names = append(names, aws.ToString(w.Name))
		workspaceIDs = append(workspaceIDs, aws.ToString(w.Id))
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set("names", names)
	d.Set("workspace_ids", workspaceIDs)

	return diags
}

func findWorkspaces(ctx context.Context, conn *grafana.Client, name string) ([]types.WorkspaceSummary, error) { // nosemgrep:ci.caps0-in-func-name
	input := &grafana.ListWorkspacesInput{}

	var output []types.WorkspaceSummary

	pages := grafana.NewListWorkspacesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}
		if name == "" {
			output = append(output, page.Workspaces...)
		} else {
			for _, workspace := range page.Workspaces {
				if *workspace.Name == name {
					output = append(output, workspace)
				}
			}
		}
	}

	return output, nil
}
