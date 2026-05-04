// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// @FrameworkDataSource("aws_ecs_daemons", name="Daemons")
func newDaemonsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &daemonsDataSource{}, nil
}

type daemonsDataSource struct {
	framework.DataSourceWithModel[daemonsDataSourceModel]
}

func (d *daemonsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"capacity_provider_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Optional:    true,
				ElementType: types.StringType,
			},
			"cluster_arn": schema.StringAttribute{
				Required: true,
			},
			"daemons": framework.DataSourceComputedListOfObjectAttribute[daemonSummaryModel](ctx),
		},
	}
}

func (d *daemonsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data daemonsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ECSClient(ctx)

	clusterArn := data.Cluster.ValueString()
	input := &ecs.ListDaemonsInput{
		ClusterArn: aws.String(clusterArn),
	}

	if !data.CapacityProviderArns.IsNull() {
		input.CapacityProviderArns = fwflex.ExpandFrameworkStringValueList(ctx, data.CapacityProviderArns)
	}

	summaries, err := findDaemons(ctx, conn, input)
	if err != nil {
		response.Diagnostics.AddError("listing ECS Daemons", err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, summaries, &data.Daemons)...)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type daemonsDataSourceModel struct {
	framework.WithRegionModel
	CapacityProviderArns fwtypes.ListOfString                                `tfsdk:"capacity_provider_arns"`
	Cluster              types.String                                        `tfsdk:"cluster_arn"`
	Daemons              fwtypes.ListNestedObjectValueOf[daemonSummaryModel] `tfsdk:"daemons"`
}

type daemonSummaryModel struct {
	CreatedAt timetypes.RFC3339 `tfsdk:"created_at"`
	DaemonArn types.String      `tfsdk:"daemon_arn"`
	Status    types.String      `tfsdk:"status"`
	UpdatedAt timetypes.RFC3339 `tfsdk:"updated_at"`
}
