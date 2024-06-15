// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ec2_capacity_block_offering", name="Capacity Block Offering")
func newDataSourceCapacityBlockOffering(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceCapacityBlockOffering{}

	return d, nil
}

type dataSourceCapacityBlockOffering struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceCapacityBlockOffering) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_ec2_capacity_block_offering"
}

func (d *dataSourceCapacityBlockOffering) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
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

const (
	DSNameCapacityBlockOffering = "Capacity Block Offering"
)

func (d *dataSourceCapacityBlockOffering) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().EC2Client(ctx)
	var data dataSourceCapacityBlockOfferingData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	input := &ec2.DescribeCapacityBlockOfferingsInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)

	if response.Diagnostics.HasError() {
		return
	}

	output, err := findCapacityBLockOffering(ctx, conn, input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, DSNameCapacityBlockOffering, data.InstanceType.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceCapacityBlockOfferingData struct {
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

func findCapacityBLockOffering(ctx context.Context, conn *ec2.Client, in *ec2.DescribeCapacityBlockOfferingsInput) (*awstypes.CapacityBlockOffering, error) {
	output, err := conn.DescribeCapacityBlockOfferings(ctx, in)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CapacityBlockOfferings) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	if len(output.CapacityBlockOfferings) > 1 {
		return nil, tfresource.NewTooManyResultsError(len(output.CapacityBlockOfferings), in)
	}

	return tfresource.AssertSingleValueResult(output.CapacityBlockOfferings)
}
