// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// @FrameworkDataSource("aws_ecs_clusters", name="Clusters")
func newClustersDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &clustersDataSource{}, nil
}

type clustersDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *clustersDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *clustersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceClustersModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ECSClient(ctx)

	input := ecs.ListClustersInput{}
	arns, err := listClusters(ctx, conn, &input)

	if err != nil {
		response.Diagnostics.AddError("listing ECS Clusters", err.Error())
		return
	}

	data.ClusterARNs = fwflex.FlattenFrameworkStringValueListOfString(ctx, arns)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func listClusters(ctx context.Context, conn *ecs.Client, input *ecs.ListClustersInput) ([]string, error) {
	var output []string

	pages := ecs.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.ClusterArns...)
	}

	return output, nil
}

type dataSourceClustersModel struct {
	ClusterARNs fwtypes.ListOfString `tfsdk:"cluster_arns"`
}
