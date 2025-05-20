// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_cloudwatch_log_destinations", name="Destinations")
func newDataSourceDestinations(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDestinations{}, nil
}

const (
	DSNameDestinations = "Destinations Data Source"
)

type dataSourceDestinations struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceDestinations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"destination_name_prefix": schema.StringAttribute{
				Optional: true,
			},
			"destinations": schema.ListAttribute{
				Computed:   true,
				CustomType: fwtypes.NewListNestedObjectTypeOf[logDestinationModel](ctx),
			},
		},
	}
}

func (d *dataSourceDestinations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().LogsClient(ctx)

	var data dataSourceLogDestinationsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findLogDestinationsByPrefix(ctx, conn, data.DestinationNamePrefix.ValueStringPointer())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudWatch, create.ErrActionReading, DSNameDestinations, data.DestinationNamePrefix.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findLogDestinationsByPrefix(ctx context.Context, conn *cloudwatchlogs.Client, destinationNamePrefix *string) (*cloudwatchlogs.DescribeDestinationsOutput, error) {
	input := &cloudwatchlogs.DescribeDestinationsInput{}

	if destinationNamePrefix != nil {
		input.DestinationNamePrefix = destinationNamePrefix
	}

	out := make([]awstypes.Destination, 0)

	destPaginator := cloudwatchlogs.NewDescribeDestinationsPaginator(conn, input)
	for destPaginator.HasMorePages() {
		page, err := destPaginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		out = append(out, page.Destinations...)
	}

	destinationsLogOut := &cloudwatchlogs.DescribeDestinationsOutput{
		Destinations: out,
	}

	return destinationsLogOut, nil
}

type dataSourceLogDestinationsModel struct {
	DestinationNamePrefix types.String                                         `tfsdk:"destination_name_prefix"`
	Destinations          fwtypes.ListNestedObjectValueOf[logDestinationModel] `tfsdk:"destinations"`
}

type logDestinationModel struct {
	ARN             types.String `tfsdk:"arn"`
	DestinationName types.String `tfsdk:"destination_name"`
	RoleARN         types.String `tfsdk:"role_arn"`
	AccessPolicy    types.String `tfsdk:"access_policy"`
	TargetARN       types.String `tfsdk:"target_arn"`
	CreationTime    types.Int64  `tfsdk:"creation_time"`
}
