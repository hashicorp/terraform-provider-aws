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

// @FrameworkDataSource("aws_pinpointsmsvoicev2_phone_number", name="Phone Number")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newPhoneNumberDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &phoneNumberDataSource{}
	return d, nil
}

type phoneNumberDataSource struct {
	framework.DataSourceWithModel[phoneNumberDataSourceModel]
}

func (d *phoneNumberDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"iso_country_code": schema.StringAttribute{
				Computed: true,
			},
			"message_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.MessageType](),
				Computed:   true,
			},
			"monthly_leasing_price": schema.StringAttribute{
				Computed: true,
			},
			"number_capabilities": schema.SetAttribute{
				CustomType:  fwtypes.NewSetTypeOf[fwtypes.StringEnum[awstypes.NumberCapability]](ctx),
				ElementType: fwtypes.StringEnumType[awstypes.NumberCapability](),
				Computed:    true,
			},
			"number_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.NumberType](),
				Computed:   true,
			},
			"opt_out_list_name": schema.StringAttribute{
				Computed: true,
			},
			"phone_number": schema.StringAttribute{
				Computed: true,
			},
			"pool_id": schema.StringAttribute{
				Computed: true,
			},
			"registration_id": schema.StringAttribute{
				Computed: true,
			},
			"self_managed_opt_outs_enabled": schema.BoolAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.NumberStatus](),
				Computed:   true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"two_way_channel_arn": schema.StringAttribute{
				Computed: true,
			},
			"two_way_channel_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"two_way_channel_role": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{},
	}
}

func (d *phoneNumberDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().PinpointSMSVoiceV2Client(ctx)

	var data phoneNumberDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	phoneNumberID := fwflex.StringValueFromFramework(ctx, data.PhoneNumberID)
	phoneNumber, err := findPhoneNumberByID(ctx, conn, phoneNumberID)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, phoneNumberID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, phoneNumber, &data), smerr.ID, phoneNumberID)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, phoneNumberID)
}

type phoneNumberDataSourceModel struct {
	framework.WithRegionModel
	CreatedTimestamp          timetypes.RFC3339                                  `tfsdk:"created_timestamp"`
	DeletionProtectionEnabled types.Bool                                         `tfsdk:"deletion_protection_enabled"`
	ISOCountryCode            types.String                                       `tfsdk:"iso_country_code"`
	MessageType               fwtypes.StringEnum[awstypes.MessageType]           `tfsdk:"message_type"`
	MonthlyLeasingPrice       types.String                                       `tfsdk:"monthly_leasing_price"`
	NumberCapabilities        fwtypes.SetOfStringEnum[awstypes.NumberCapability] `tfsdk:"number_capabilities"`
	NumberType                fwtypes.StringEnum[awstypes.NumberType]            `tfsdk:"number_type"`
	OptOutListName            types.String                                       `tfsdk:"opt_out_list_name"`
	PhoneNumber               types.String                                       `tfsdk:"phone_number"`
	PhoneNumberARN            types.String                                       `tfsdk:"arn"`
	PhoneNumberID             types.String                                       `tfsdk:"id"`
	PoolID                    types.String                                       `tfsdk:"pool_id"`
	RegistrationID            types.String                                       `tfsdk:"registration_id"`
	SelfManagedOptOutsEnabled types.Bool                                         `tfsdk:"self_managed_opt_outs_enabled"`
	Status                    fwtypes.StringEnum[awstypes.NumberStatus]          `tfsdk:"status"`
	Tags                      tftags.Map                                         `tfsdk:"tags"`
	TwoWayChannelARN          types.String                                       `tfsdk:"two_way_channel_arn"`
	TwoWayChannelRole         fwtypes.ARN                                        `tfsdk:"two_way_channel_role"`
	TwoWayEnabled             types.Bool                                         `tfsdk:"two_way_channel_enabled"`
}
