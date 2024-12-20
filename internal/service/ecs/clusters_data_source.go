// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkDataSource("aws_ecs_clusters", name="Clusters")
func newDataSourceClusters(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceClusters{}, nil
}

const (
	DSNameClusters = "Clusters Data Source"
)

type dataSourceClusters struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceClusters) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_ecs_clusters"
}

func (d *dataSourceClusters) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_arns": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *dataSourceClusters) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ECSClient(ctx)

	var data dataSourceClustersModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterArns, err := listClusters(ctx, conn, &ecs.ListClustersInput{})
	if err != nil {
		resp.Diagnostics.AddError("Listing ECS clusters", err.Error())
		return
	}

	out := &ecs.ListClustersOutput{
		ClusterArns: clusterArns,
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("Clusters"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func listClusters(ctx context.Context, conn *ecs.Client, input *ecs.ListClustersInput) ([]string, error) {
	var clusterArns []string

	pages := ecs.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		clusterArns = append(clusterArns, page.ClusterArns...)
	}

	return clusterArns, nil
}

type dataSourceClustersModel struct {
	ClusterArns basetypes.ListValue `tfsdk:"cluster_arns"`
}
