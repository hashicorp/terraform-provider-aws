// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_spot_datafeed_subscription", name="Spot Data Feed Subscription Data Source")
func newDataSourceSpotDataFeedSubscription(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceSpotDataFeedSubscription{}, nil
}

const (
	DSNameSpotDataFeedSubscription = "Spot Data Feed Subscription Data Source"
)

type dataSourceSpotDataFeedSubscription struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceSpotDataFeedSubscription) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrBucket: schema.StringAttribute{
				Computed: true,
			},
			names.AttrPrefix: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *dataSourceSpotDataFeedSubscription) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().EC2Client(ctx)
	accountID := d.Meta().AccountID(ctx)

	var data dataSourceSpotDataFeedSubscriptionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.DescribeSpotDatafeedSubscriptionInput{}
	out, err := conn.DescribeSpotDatafeedSubscription(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, DSNameSpotDataFeedSubscription, accountID, err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.SpotDatafeedSubscription, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceSpotDataFeedSubscriptionModel struct {
	Bucket types.String `tfsdk:"bucket"`
	Prefix types.String `tfsdk:"prefix"`
}
