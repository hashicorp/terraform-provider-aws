// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_vpc_ipam", name="AWS IPAM")
func newDataSourceAwsVpcIpam(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAwsVpcIpam{}, nil
}

const (
	DSNameAwsVpcIpam = "AWS IPAM Data Source"
)

type dataSourceAwsVpcIpam struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceAwsVpcIpam) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_vpc_ipam"
}

func (d *dataSourceAwsVpcIpam) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"default_resource_discovery_association_id": schema.StringAttribute{
				Computed: true,
			},
			"default_resource_discovery_id": schema.StringAttribute{
				Computed: true,
			},
			"tier": schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.IpamTier](),
			},
			"scope_count": schema.Int32Attribute{
				Computed: true,
			},
			"enable_private_gua": schema.BoolAttribute{
				Computed: true,
			},
			"region": schema.StringAttribute{
				Computed: true,
			},
			"operating_regions": framework.DataSourceComputedListOfObjectAttribute[ipamOperatingRegionModel](ctx),
			"owner_id": schema.StringAttribute{
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
			"state": schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.IpamState](),
			},
			"state_message": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *dataSourceAwsVpcIpam) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().EC2Client(ctx)

	var data dataSourceAwsVpcIpamModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipam, err := findIPAMByID(ctx, conn, data.IpamId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionReading, DSNameAwsVpcIpam, data.IpamId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, ipam, &data, flex.WithFieldNamePrefix("ipam"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Tags = tftags.FlattenStringValueMap(ctx, keyValueTags(ctx, ipam.Tags).IgnoreAWS().Map())
	// ignoreTagsConfig := d.Meta().IgnoreTagsConfig(ctx)
	// tags := keyValueTags(ctx, ipam.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	// data.Tags = tftags.FlattenStringValueMap(ctx, keyValueTags(ctx, ipam.Tags).IgnoreAWS().Map())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceAwsVpcIpamSummaryModel struct {
	Description                           types.String                                              `tfsdk:"description"`
	DefaultResourceDiscoveryAssociationId types.String                                              `tfsdk:"default_resource_discovery_association_id"`
	DefaultResourceDiscoveryId            types.String                                              `tfsdk:"default_resource_discovery_id"`
	EnablePrivateGua                      types.Bool                                                `tfsdk:"enable_private_gua"`
	IpamARN                               types.String                                              `tfsdk:"arn"`
	IpamId                                types.String                                              `tfsdk:"id"`
	IpamRegion                            types.String                                              `tfsdk:"region"`
	OperatingRegions                      fwtypes.ListNestedObjectValueOf[ipamOperatingRegionModel] `tfsdk:"operating_regions"`
	OwnerID                               types.String                                              `tfsdk:"owner_id"`
	PrivateDefaultScopeId                 types.String                                              `tfsdk:"private_default_scope_id"`
	PublicDefaultScopeId                  types.String                                              `tfsdk:"public_default_scope_id"`
	ResourceDiscoveryAssociationCount     types.Int32                                               `tfsdk:"resource_discovery_association_count"`
	ScopeCount                            types.Int32                                               `tfsdk:"scope_count"`
	State                                 fwtypes.StringEnum[awstypes.IpamState]                    `tfsdk:"state"`
	StateMessage                          types.String                                              `tfsdk:"state_message"`
	Tier                                  fwtypes.StringEnum[awstypes.IpamTier]                     `tfsdk:"tier"`
}

type dataSourceAwsVpcIpamModel struct {
	dataSourceAwsVpcIpamSummaryModel
	Tags tftags.Map `tfsdk:"tags"`
}

type ipamOperatingRegionModel struct {
	RegionName types.String `tfsdk:"region_name"`
}
