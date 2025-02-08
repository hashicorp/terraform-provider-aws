// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_regions", name="Regions")
func newRegionsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &regionsDataSource{}

	return d, nil
}

type regionsDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *regionsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
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

func (d *regionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data regionsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	input := &ec2.DescribeRegionsInput{
		AllRegions: fwflex.BoolFromFramework(ctx, data.AllRegions),
		Filters:    tfec2.NewCustomFilterListFramework(ctx, data.Filters),
	}

	output, err := conn.DescribeRegions(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("reading Regions", err.Error())

		return
	}

	data.ID = fwflex.StringValueToFrameworkLegacy(ctx, d.Meta().Partition(ctx))
	names := tfslices.ApplyToAll(output.Regions, func(v awstypes.Region) string {
		return aws.ToString(v.RegionName)
	})
	data.Names = fwflex.FlattenFrameworkStringValueSetLegacy(ctx, names)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type regionsDataSourceModel struct {
	AllRegions types.Bool   `tfsdk:"all_regions"`
	Filters    types.Set    `tfsdk:"filter"`
	ID         types.String `tfsdk:"id"`
	Names      types.Set    `tfsdk:"names"`
}
