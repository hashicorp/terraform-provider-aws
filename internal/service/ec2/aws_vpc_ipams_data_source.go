// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_vpc_ipams", name="AWS IPAM")
func newDataSourceAwsVpcIpams(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAwsVpcIpams{}, nil
}

const (
	DSNameAwsVpcIpams = "AWS IPAMs Data Source"
)

type dataSourceAwsVpcIpams struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceAwsVpcIpams) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_vpc_ipams"
}

func (d *dataSourceAwsVpcIpams) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ipams": framework.DataSourceComputedListOfObjectAttribute[dataSourceAwsVpcIpamSummaryModel](ctx),
			"ipam_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[filterModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						names.AttrValues: schema.SetAttribute{
							CustomType: fwtypes.SetOfStringType,
							Required:   true,
						},
					},
				},
			},
		},
	}
}

func findAwsVpcIpams(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamsInput) ([]awstypes.Ipam, error) {
	var output []awstypes.Ipam

	pages := ec2.NewDescribeIpamsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		output = append(output, page.Ipams...)
	}
	return output, nil
}

func (d *dataSourceAwsVpcIpams) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().EC2Client(ctx)

	var data dataSourceAwsVpcIpamsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.DescribeIpamsInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := findAwsVpcIpams(ctx, conn, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, DSNameAwsVpcIpams, "", err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &data.Ipams, fwflex.WithFieldNamePrefix("ipam"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type filterModel struct {
	Name   types.String        `tfsdk:"name"`
	Values fwtypes.SetOfString `tfsdk:"values"`
}

type dataSourceAwsVpcIpamsModel struct {
	Ipams   fwtypes.ListNestedObjectValueOf[dataSourceAwsVpcIpamSummaryModel] `tfsdk:"ipams"`
	Filters fwtypes.ListNestedObjectValueOf[filterModel]                      `tfsdk:"filter"`
	IpamIds types.List                                                        `tfsdk:"ipam_ids"`
}
