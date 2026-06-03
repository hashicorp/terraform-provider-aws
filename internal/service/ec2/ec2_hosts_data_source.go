// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

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
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ec2_hosts", name="Hosts")
func newHostsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &hostsDataSource{}, nil
}

type hostsDataSource struct {
	framework.DataSourceWithModel[hostsDataSourceModel]
}

func (d *hostsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrIDs: schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"outpost_arn": schema.StringAttribute{
				Optional: true,
			},
			names.AttrTags: tftags.TagsAttribute(),
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: customFiltersBlock(ctx),
		},
	}
}

func (d *hostsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data hostsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	input := ec2.DescribeHostsInput{
		Filter: append(newCustomFilterListFramework(ctx, data.Filters), newTagFilterList(svcTags(tftags.New(ctx, data.Tags)))...),
	}

	if len(input.Filter) == 0 {
		input.Filter = nil
	}

	output, err := findHosts(ctx, conn, &input)

	if err != nil {
		resp.Diagnostics.AddError("reading EC2 Hosts", err.Error())
		return
	}

	// Client-side filter on OutpostArn since the API does not support it as a server-side filter.
	if !data.OutpostARN.IsNull() && !data.OutpostARN.IsUnknown() {
		outpostARN := data.OutpostARN.ValueString()
		output = tfslices.Filter(output, func(v awstypes.Host) bool {
			return aws.ToString(v.OutpostArn) == outpostARN
		})
	}

	data.IDs = fwflex.FlattenFrameworkStringValueListOfString(ctx, tfslices.ApplyToAll(output, func(v awstypes.Host) string {
		return aws.ToString(v.HostId)
	}))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type hostsDataSourceModel struct {
	framework.WithRegionModel
	Filters    customFilters        `tfsdk:"filter"`
	IDs        fwtypes.ListOfString `tfsdk:"ids"`
	OutpostARN types.String         `tfsdk:"outpost_arn"`
	Tags       tftags.Map           `tfsdk:"tags"`
}
