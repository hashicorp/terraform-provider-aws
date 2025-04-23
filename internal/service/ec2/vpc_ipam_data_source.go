// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_vpc_ipam", name="IPAM")
// @Tags
// @Testing(tagsTest=false)
func newIPAMDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &ipamDataSource{}, nil
}

type ipamDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *ipamDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"default_resource_discovery_association_id": schema.StringAttribute{
				Computed: true,
			},
			"default_resource_discovery_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"enable_private_gua": schema.BoolAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			"ipam_region": schema.StringAttribute{
				Computed: true,
			},
			"operating_regions": framework.DataSourceComputedListOfObjectAttribute[ipamOperatingRegionModel](ctx),
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
			},
			"private_default_scope_id": schema.StringAttribute{
				Computed: true,
			},
			"public_default_scope_id": schema.StringAttribute{
				Computed: true,
			},
			"resource_discovery_association_count": schema.Int32Attribute{
				Computed: true,
			},
			"scope_count": schema.Int32Attribute{
				Computed: true,
			},
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.IpamState](),
				Computed:   true,
			},
			"state_message": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"tier": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.IpamTier](),
				Computed:   true,
			},
		},
	}
}

func (d *ipamDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data ipamDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	ipam, err := findIPAMByID(ctx, conn, data.IpamID.ValueString())

	if err != nil {
		response.Diagnostics.AddError("reading IPAM", tfresource.SingularDataSourceFindError("IPAM", err).Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, ipam, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, ipam.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type ipamDataSourceModel struct {
	ipamModel
	Tags tftags.Map `tfsdk:"tags"`
}

type ipamModel struct {
	DefaultResourceDiscoveryAssociationId types.String                                              `tfsdk:"default_resource_discovery_association_id"`
	DefaultResourceDiscoveryId            types.String                                              `tfsdk:"default_resource_discovery_id"`
	Description                           types.String                                              `tfsdk:"description"`
	EnablePrivateGUA                      types.Bool                                                `tfsdk:"enable_private_gua"`
	IpamARN                               types.String                                              `tfsdk:"arn"`
	IpamID                                types.String                                              `tfsdk:"id"`
	IpamRegion                            types.String                                              `tfsdk:"ipam_region"`
	OperatingRegions                      fwtypes.ListNestedObjectValueOf[ipamOperatingRegionModel] `tfsdk:"operating_regions"`
	OwnerID                               types.String                                              `tfsdk:"owner_id"`
	PrivateDefaultScopeID                 types.String                                              `tfsdk:"private_default_scope_id"`
	PublicDefaultScopeID                  types.String                                              `tfsdk:"public_default_scope_id"`
	ResourceDiscoveryAssociationCount     types.Int32                                               `tfsdk:"resource_discovery_association_count"`
	ScopeCount                            types.Int32                                               `tfsdk:"scope_count"`
	State                                 fwtypes.StringEnum[awstypes.IpamState]                    `tfsdk:"state"`
	StateMessage                          types.String                                              `tfsdk:"state_message"`
	Tier                                  fwtypes.StringEnum[awstypes.IpamTier]                     `tfsdk:"tier"`
}

type ipamOperatingRegionModel struct {
	RegionName types.String `tfsdk:"region_name"`
}
