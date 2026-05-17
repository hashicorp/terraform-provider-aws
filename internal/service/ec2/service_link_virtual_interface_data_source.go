// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ec2_service_link_virtual_interface", name="Service Link Virtual Interface")
// @Tags
// @Testing(tagsTest=false)
func newServiceLinkVirtualInterfaceDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &serviceLinkVirtualInterfaceDataSource{}, nil
}

type serviceLinkVirtualInterfaceDataSource struct {
	framework.DataSourceWithModel[serviceLinkVirtualInterfaceDataSourceModel]
}

func (d *serviceLinkVirtualInterfaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"configuration_state": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ServiceLinkVirtualInterfaceConfigurationState](),
				Computed:   true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"local_address": schema.StringAttribute{
				Computed: true,
			},
			"outpost_arn": schema.StringAttribute{
				Computed: true,
			},
			"outpost_id": schema.StringAttribute{
				Computed: true,
			},
			"outpost_lag_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
			},
			"peer_address": schema.StringAttribute{
				Computed: true,
			},
			"peer_bgp_asn": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"vlan": schema.Int32Attribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: customFiltersBlock(ctx),
		},
	}
}

func (d *serviceLinkVirtualInterfaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serviceLinkVirtualInterfaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	input := ec2.DescribeServiceLinkVirtualInterfacesInput{
		Filters: newCustomFilterListFramework(ctx, data.Filters),
	}

	if !data.ServiceLinkVirtualInterfaceID.IsNull() && !data.ServiceLinkVirtualInterfaceID.IsUnknown() {
		input.ServiceLinkVirtualInterfaceIds = []string{data.ServiceLinkVirtualInterfaceID.ValueString()}
	}

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	out, err := findServiceLinkVirtualInterface(ctx, conn, &input)

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading EC2 Service Link Virtual Interface (%s)", data.ServiceLinkVirtualInterfaceID.ValueString()), tfresource.SingularDataSourceFindError("EC2 Service Link Virtual Interface", err).Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *serviceLinkVirtualInterfaceDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot(names.AttrID),
			path.MatchRoot(names.AttrFilter),
		),
	}
}

type serviceLinkVirtualInterfaceDataSourceModel struct {
	framework.WithRegionModel
	ConfigurationState             fwtypes.StringEnum[awstypes.ServiceLinkVirtualInterfaceConfigurationState] `tfsdk:"configuration_state"`
	Filters                        customFilters                                                              `tfsdk:"filter"`
	LocalAddress                   types.String                                                               `tfsdk:"local_address"`
	OutpostARN                     types.String                                                               `tfsdk:"outpost_arn"`
	OutpostID                      types.String                                                               `tfsdk:"outpost_id"`
	OutpostLagID                   types.String                                                               `tfsdk:"outpost_lag_id"`
	OwnerID                        types.String                                                               `tfsdk:"owner_id"`
	PeerAddress                    types.String                                                               `tfsdk:"peer_address"`
	PeerBGPASN                     types.Int64                                                                `tfsdk:"peer_bgp_asn"`
	ServiceLinkVirtualInterfaceARN types.String                                                               `tfsdk:"arn"`
	ServiceLinkVirtualInterfaceID  types.String                                                               `tfsdk:"id"`
	Tags                           tftags.Map                                                                 `tfsdk:"tags"`
	Vlan                           types.Int32                                                                `tfsdk:"vlan"`
}
