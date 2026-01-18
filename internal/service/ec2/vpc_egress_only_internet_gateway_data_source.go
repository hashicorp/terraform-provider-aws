// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_egress_only_internet_gateway", name="Egress-Only Internet Gateway")
// @Tags
// @Testing(tagsTest=false)
func newEgressOnlyInternetGatewayDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &egressOnlyInternetGatewayDataSource{}, nil
}

type egressOnlyInternetGatewayDataSource struct {
	framework.DataSourceWithModel[egressOnlyInternetGatewayDataSourceModel]
}

func (d *egressOnlyInternetGatewayDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: customFiltersBlock(ctx),
		},
	}
}

func (d *egressOnlyInternetGatewayDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data egressOnlyInternetGatewayDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)
	//ignoreTagsConfig := d.Meta().IgnoreTagsConfig(ctx)

	input := &ec2.DescribeEgressOnlyInternetGatewaysInput{
		Filters: newCustomFilterListFramework(ctx, data.Filters),
	}

	if !data.ID.IsNull() {
		input.EgressOnlyInternetGatewayIds = []string{fwflex.StringValueFromFramework(ctx, data.ID)}
	}

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := findEgressOnlyInternetGateway(ctx, conn, input)

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.EgressOnlyInternetGatewayId)

	// Set state and VPC ID from attachments
	if len(output.Attachments) > 0 {
		attachment := output.Attachments[0]
		data.State = fwflex.StringValueToFramework(ctx, attachment.State)
		data.VpcID = fwflex.StringToFramework(ctx, attachment.VpcId)
	}

	//data.Tags = tftags.FlattenStringValueMap(ctx, keyValueTags(ctx, output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())
	setTagsOut(ctx, output.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.ID)
}

type egressOnlyInternetGatewayDataSourceModel struct {
	framework.WithRegionModel
	Filters customFilters `tfsdk:"filter"`
	ID      types.String  `tfsdk:"id"`
	State   types.String  `tfsdk:"state"`
	Tags    tftags.Map    `tfsdk:"tags"`
	VpcID   types.String  `tfsdk:"vpc_id"`
}
