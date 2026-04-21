// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
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
			"cluster": schema.StringAttribute{
				Required: true,
			},
			"daemon_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
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

	var daemonArns []string
	for _, summary := range summaries {
		if summary.DaemonArn != nil {
			daemonArns = append(daemonArns, aws.ToString(summary.DaemonArn))
		}
	}

	data.DaemonArns = fwflex.FlattenFrameworkStringValueListOfString(ctx, daemonArns)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type daemonsDataSourceModel struct {
	CapacityProviderArns fwtypes.ListOfString `tfsdk:"capacity_provider_arns"`
	Cluster              types.String         `tfsdk:"cluster"`
	DaemonArns           fwtypes.ListOfString `tfsdk:"daemon_arns"`
	Region               types.String         `tfsdk:"region"`
}
