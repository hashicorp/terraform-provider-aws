// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name=Regions)
func newDataSourceRegions(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceRegions{}

	return d, nil
}

type dataSourceRegions struct {
	framework.DataSourceWithConfigure
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceRegions) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_regions"
}

// Schema returns the schema for this data source.
func (d *dataSourceRegions) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"all_regions": schema.BoolAttribute{
				Optional: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrNames: schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: tfec2.CustomFiltersBlock(),
		},
	}
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourceRegions) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceRegionsData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	input := &ec2.DescribeRegionsInput{
		AllRegions: flex.BoolFromFramework(ctx, data.AllRegions),
		Filters:    tfec2.NewCustomFilterListFramework(ctx, data.Filters),
	}

	output, err := conn.DescribeRegions(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("reading Regions", err.Error())

		return
	}

	var names []string
	for _, v := range output.Regions {
		names = append(names, aws.ToString(v.RegionName))
	}

	data.ID = types.StringValue(d.Meta().Partition)
	data.Names = flex.FlattenFrameworkStringValueSetLegacy(ctx, names)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceRegionsData struct {
	AllRegions types.Bool   `tfsdk:"all_regions"`
	Filters    types.Set    `tfsdk:"filter"`
	ID         types.String `tfsdk:"id"`
	Names      types.Set    `tfsdk:"names"`
}
