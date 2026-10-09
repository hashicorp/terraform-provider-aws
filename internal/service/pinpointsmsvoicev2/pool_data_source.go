// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_pinpointsmsvoicev2_pool", name="Pool")
// @Tags(identifierAttribute="arn")
// @Testing(generator=false)
// @Testing(preCheck="testAccPreCheckPool")
func newPoolDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &poolDataSource{}, nil
}

type poolDataSource struct {
	framework.DataSourceWithModel[poolDataSourceModel]
}

func (d *poolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"created_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"deletion_protection_enabled": schema.BoolAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			"message_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.MessageType](),
				Computed:   true,
			},
			"opt_out_list_name": schema.StringAttribute{
				Computed: true,
			},
			"origination_identities": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"self_managed_opt_outs_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"shared_routes_enabled": schema.BoolAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PoolStatus](),
				Computed:   true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"two_way_channel_arn": schema.StringAttribute{
				Computed: true,
			},
			"two_way_channel_role": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"two_way_enabled": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *poolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().PinpointSMSVoiceV2Client(ctx)

	var data poolDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	poolID := fwflex.StringValueFromFramework(ctx, data.PoolID)
	pool, err := findPoolByID(ctx, conn, poolID)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, poolID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, pool, &data), smerr.ID, poolID)
	if resp.Diagnostics.HasError() {
		return
	}

	// DescribePools does not return a pool's origination identities.
	originationIdentities, err := findPoolOriginationIdentities(ctx, conn, poolID)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, poolID)
		return
	}
	data.OriginationIdentities = fwflex.FlattenFrameworkStringValueSetOfString(ctx, originationIdentities)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, poolID)
}

type poolDataSourceModel struct {
	framework.WithRegionModel
	CreatedTimestamp          timetypes.RFC3339                        `tfsdk:"created_timestamp"`
	DeletionProtectionEnabled types.Bool                               `tfsdk:"deletion_protection_enabled"`
	MessageType               fwtypes.StringEnum[awstypes.MessageType] `tfsdk:"message_type"`
	OptOutListName            types.String                             `tfsdk:"opt_out_list_name"`
	OriginationIdentities     fwtypes.SetOfString                      `tfsdk:"origination_identities"`
	PoolARN                   types.String                             `tfsdk:"arn"`
	PoolID                    types.String                             `tfsdk:"id"`
	SelfManagedOptOutsEnabled types.Bool                               `tfsdk:"self_managed_opt_outs_enabled"`
	SharedRoutesEnabled       types.Bool                               `tfsdk:"shared_routes_enabled"`
	Status                    fwtypes.StringEnum[awstypes.PoolStatus]  `tfsdk:"status"`
	Tags                      tftags.Map                               `tfsdk:"tags"`
	TwoWayChannelARN          types.String                             `tfsdk:"two_way_channel_arn"`
	TwoWayChannelRole         fwtypes.ARN                              `tfsdk:"two_way_channel_role"`
	TwoWayEnabled             types.Bool                               `tfsdk:"two_way_enabled"`
}
