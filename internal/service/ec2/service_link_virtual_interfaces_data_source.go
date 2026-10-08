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

// @FrameworkDataSource("aws_ec2_service_link_virtual_interfaces", name="Service Link Virtual Interfaces")
func newServiceLinkVirtualInterfacesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &serviceLinkVirtualInterfacesDataSource{}, nil
}

type serviceLinkVirtualInterfacesDataSource struct {
	framework.DataSourceWithModel[serviceLinkVirtualInterfacesDataSourceModel]
}

func (d *serviceLinkVirtualInterfacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrIDs: schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrTags: tftags.TagsAttribute(),
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: customFiltersBlock(ctx),
		},
	}
}

func (d *serviceLinkVirtualInterfacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serviceLinkVirtualInterfacesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	input := ec2.DescribeServiceLinkVirtualInterfacesInput{
		Filters: append(newCustomFilterListFramework(ctx, data.Filters), newTagFilterList(svcTags(tftags.New(ctx, data.Tags)))...),
	}

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := findServiceLinkVirtualInterfaces(ctx, conn, &input)

	if err != nil {
		resp.Diagnostics.AddError("reading EC2 Service Link Virtual Interfaces", err.Error())
		return
	}

	data.IDs = fwflex.FlattenFrameworkStringValueListOfString(ctx, tfslices.ApplyToAll(output, func(v awstypes.ServiceLinkVirtualInterface) string {
		return aws.ToString(v.ServiceLinkVirtualInterfaceId)
	}))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type serviceLinkVirtualInterfacesDataSourceModel struct {
	framework.WithRegionModel
	Filters customFilters        `tfsdk:"filter"`
	IDs     fwtypes.ListOfString `tfsdk:"ids"`
	Tags    tftags.Map           `tfsdk:"tags"`
}
