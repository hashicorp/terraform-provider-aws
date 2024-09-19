// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource
func newDataSourceRegion(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceRegion{}

	return d, nil
}

type dataSourceRegion struct {
	framework.DataSourceWithConfigure
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceRegion) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_region"
}

// Schema returns the schema for this data source.
func (d *dataSourceRegion) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrEndpoint: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourceRegion) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceRegionData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	var region *awstypes.Region

	conn := d.Meta().EC2Client(ctx)

	if !data.Endpoint.IsNull() {
		matchingRegion, err := FindRegionByEndpoint(ctx, conn, data.Endpoint.ValueString())

		if err != nil {
			response.Diagnostics.AddError("finding Region by endpoint", err.Error())

			return
		}

		region = matchingRegion
	}

	if !data.Name.IsNull() {
		matchingRegion, err := FindRegionByName(ctx, conn, data.Name.ValueString())

		if err != nil {
			response.Diagnostics.AddError("finding Region by name", err.Error())

			return
		}

		if region != nil && region.RegionName != matchingRegion.RegionName {
			response.Diagnostics.AddError("multiple Regions matched", "use additional constraints to reduce matches to a single Region")

			return
		}

		region = matchingRegion
	}

	// Default to provider current region if no other filters matched
	if region == nil {
		matchingRegion, err := FindRegionByName(ctx, conn, d.Meta().Region)

		if err != nil {
			response.Diagnostics.AddError("finding Region by name", err.Error())

			return
		}

		region = matchingRegion
	}

	data.Description = types.StringValue(aws.ToString(region.RegionName))
	data.Endpoint = types.StringValue(aws.ToString(region.Endpoint))
	data.ID = types.StringValue(aws.ToString(region.RegionName))
	data.Name = types.StringValue(aws.ToString(region.RegionName))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceRegionData struct {
	Description types.String `tfsdk:"description"`
	Endpoint    types.String `tfsdk:"endpoint"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
}

func FindRegionByEndpoint(ctx context.Context, conn *ec2.Client, endpoint string) (*awstypes.Region, error) {
	input := &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	}

	output, err := conn.DescribeRegions(ctx, input)
	if err != nil {
		return nil, err
	}

	for _, region := range output.Regions {
		if aws.ToString(region.Endpoint) == endpoint {
			return &region, nil
		}
	}

	return nil, fmt.Errorf("region not found for endpoint %q", endpoint)
}

func FindRegionByName(ctx context.Context, conn *ec2.Client, name string) (*awstypes.Region, error) {
	input := &ec2.DescribeRegionsInput{
		RegionNames: []string{name},
		AllRegions:  aws.Bool(true),
	}

	output, err := conn.DescribeRegions(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(output.Regions) == 0 {
		return nil, fmt.Errorf("region not found for name %q", name)
	}

	return &output.Regions[0], nil
}
