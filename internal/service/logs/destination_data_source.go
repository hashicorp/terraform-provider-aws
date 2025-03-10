// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_cloudwatch_log_destination", name="Destination")
func newDataSourceDestination(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDestination{}, nil
}

const (
	DSNameDestination = "Destination Data Source"
)

type dataSourceDestination struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceDestination) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed: true,
			},
			"destination_name": schema.StringAttribute{
				Required: true,
			},
			"role_arn": schema.StringAttribute{
				Computed: true,
			},
			"access_policy": schema.StringAttribute{
				Computed: true,
			},
			"target_arn": schema.StringAttribute{
				Computed: true,
			},
			"creation_time": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *dataSourceDestination) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().LogsClient(ctx)

	// TIP: -- 2. Fetch the config
	var data dataSourceDestinationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDestinationByName(ctx, conn, data.DestinationName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionReading, DSNameDestination, data.DestinationName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("Destination"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceDestinationModel struct {
	ARN             types.String `tfsdk:"arn"`
	DestinationName types.String `tfsdk:"destination_name"`
	RoleARN         types.String `tfsdk:"role_arn"`
	AccessPolicy    types.String `tfsdk:"access_policy"`
	TargetARN       types.String `tfsdk:"target_arn"`
	CreationTime    types.Int64  `tfsdk:"creation_time"`
}
