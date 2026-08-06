// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// @FrameworkDataSource("aws_ecs_services", name="Services")
func newServicesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &servicesDataSource{}, nil
}

type servicesDataSource struct {
	framework.DataSourceWithModel[servicesDataSourceModel]
}

func (d *servicesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_arn": schema.StringAttribute{
				Required: true,
			},
			"launch_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.LaunchType](),
				Optional:   true,
			},
			"scheduling_strategy": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.SchedulingStrategy](),
				Optional:   true,
			},
			"service_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *servicesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data servicesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ECSClient(ctx)

	input := ecs.ListServicesInput{
		Cluster: fwflex.StringFromFramework(ctx, data.ClusterARN),
	}

	if !data.LaunchType.IsNull() {
		input.LaunchType = awstypes.LaunchType(data.LaunchType.ValueString())
	}

	if !data.SchedulingStrategy.IsNull() {
		input.SchedulingStrategy = awstypes.SchedulingStrategy(data.SchedulingStrategy.ValueString())
	}

	arns, err := listServices(ctx, conn, &input)

	if err != nil {
		response.Diagnostics.AddError("listing ECS Services", err.Error())
		return
	}

	data.ServiceARNs = fwflex.FlattenFrameworkStringValueListOfString(ctx, arns)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func listServices(ctx context.Context, conn *ecs.Client, input *ecs.ListServicesInput) ([]string, error) {
	var output []string

	pages := ecs.NewListServicesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.ServiceArns...)
	}

	return output, nil
}

type servicesDataSourceModel struct {
	framework.WithRegionModel
	ClusterARN         types.String                                    `tfsdk:"cluster_arn"`
	LaunchType         fwtypes.StringEnum[awstypes.LaunchType]         `tfsdk:"launch_type"`
	SchedulingStrategy fwtypes.StringEnum[awstypes.SchedulingStrategy] `tfsdk:"scheduling_strategy"`
	ServiceARNs        fwtypes.ListOfString                            `tfsdk:"service_arns"`
}
