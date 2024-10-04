// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	"github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_prometheus_workspaces", name="Workspaces")
func dataSourceWorkspaces() *schema.Resource { // nosemgrep:ci.caps0-in-func-name
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
			names.AttrARNs: {
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	alias_prefix := d.Get("alias_prefix").(string)
	workspaces, err := findWorkspaces(ctx, conn, alias_prefix)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Prometheus Workspaces: %s", err)
	}

	var arns, aliases, workspaceIDs []string
	for _, w := range workspaces {
		arns = append(arns, aws.ToString(w.Arn))
		aliases = append(aliases, aws.ToString(w.Alias))
		workspaceIDs = append(workspaceIDs, aws.ToString(w.WorkspaceId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("aliases", aliases)
	d.Set(names.AttrARNs, arns)
	d.Set("workspace_ids", workspaceIDs)

	return diags
}

func findWorkspaces(ctx context.Context, conn *amp.Client, alias string) ([]types.WorkspaceSummary, error) { // nosemgrep:ci.caps0-in-func-name
	input := &amp.ListWorkspacesInput{}
	if alias != "" {
		input.Alias = aws.String(alias)
	}

	var output []types.WorkspaceSummary
	pages := amp.NewListWorkspacesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Workspaces...)
	}

	return output, nil
}
