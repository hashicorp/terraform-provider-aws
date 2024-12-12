// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ec2_capacity_block_offering", name="Capacity Block Offering")
func newCapacityBlockOfferingDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &capacityBlockOfferingDataSource{}

	return d, nil
}

type capacityBlockOfferingDataSource struct {
	framework.DataSourceWithConfigure
}

func (*capacityBlockOfferingDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_ec2_capacity_block_offering"
}

func (d *capacityBlockOfferingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAvailabilityZone: schema.StringAttribute{
				Computed: true,
			},
			"capacity_block_offering_id": framework.IDAttribute(),
			"capacity_duration_hours": schema.Int64Attribute{
				Required: true,
			},
			"currency_code": schema.StringAttribute{
				Computed: true,
			},
			"end_date_range": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Optional:   true,
				Computed:   true,
			},
			names.AttrInstanceCount: schema.Int64Attribute{
				Required: true,
			},
			names.AttrInstanceType: schema.StringAttribute{
				Required: true,
			},
			"start_date_range": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Optional:   true,
				Computed:   true,
			},
			"tenancy": schema.StringAttribute{
				Computed: true,
			},
			"upfront_fee": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *capacityBlockOfferingDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data capacityBlockOfferingDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	input := &ec2.DescribeCapacityBlockOfferingsInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findCapacityBlockOffering(ctx, conn, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Capacity Block Offering (%s)", data.InstanceType.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type capacityBlockOfferingDataSourceModel struct {
	AvailabilityZone        types.String      `tfsdk:"availability_zone"`
	CapacityDurationHours   types.Int64       `tfsdk:"capacity_duration_hours"`
	CurrencyCode            types.String      `tfsdk:"currency_code"`
	EndDateRange            timetypes.RFC3339 `tfsdk:"end_date_range"`
	CapacityBlockOfferingID types.String      `tfsdk:"capacity_block_offering_id"`
	InstanceCount           types.Int64       `tfsdk:"instance_count"`
	InstanceType            types.String      `tfsdk:"instance_type"`
	StartDateRange          timetypes.RFC3339 `tfsdk:"start_date_range"`
	Tenancy                 types.String      `tfsdk:"tenancy"`
	UpfrontFee              types.String      `tfsdk:"upfront_fee"`
}
