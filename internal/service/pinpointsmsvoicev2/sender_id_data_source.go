// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_pinpointsmsvoicev2_sender_id", name="Sender ID")
// @Tags(identifierAttribute="arn")
func newSenderIDDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &senderIDDataSource{}, nil
}

type senderIDDataSource struct {
	framework.DataSourceWithModel[senderIDDataSourceModel]
}

func (d *senderIDDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"deletion_protection_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"iso_country_code": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Z]{2}$`), "must be in ISO 3166-1 alpha-2 format"),
				},
			},
			"message_types": schema.SetAttribute{
				CustomType:  fwtypes.NewSetTypeOf[fwtypes.StringEnum[awstypes.MessageType]](ctx),
				Computed:    true,
				ElementType: fwtypes.StringEnumType[awstypes.MessageType](),
			},
			"monthly_leasing_price": schema.StringAttribute{
				Computed: true,
			},
			"registered": schema.BoolAttribute{
				Computed: true,
			},
			"registration_id": schema.StringAttribute{
				Computed: true,
			},
			"sender_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Z0-9-]{3,11}$`), "must be between 3 and 11 characters and contain only uppercase letters, numbers, and dashes"),
					stringvalidator.RegexMatches(regexache.MustCompile(`[A-Z]`), "must contain at least one letter (numeric-only sender IDs are not supported)"),
				},
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *senderIDDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data senderIDDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().PinpointSMSVoiceV2Client(ctx)

	out, err := findSenderIDByTwoPartKey(ctx, conn, data.SenderID.ValueString(), data.ISOCountryCode.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.SenderID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	data.MessageTypes = normalizeMessageTypes(ctx, out.MessageTypes)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type senderIDDataSourceModel struct {
	framework.WithRegionModel
	DeletionProtectionEnabled types.Bool                                    `tfsdk:"deletion_protection_enabled"`
	ISOCountryCode            types.String                                  `tfsdk:"iso_country_code"`
	MessageTypes              fwtypes.SetOfStringEnum[awstypes.MessageType] `tfsdk:"message_types"`
	MonthlyLeasingPrice       types.String                                  `tfsdk:"monthly_leasing_price"`
	Registered                types.Bool                                    `tfsdk:"registered"`
	RegistrationID            types.String                                  `tfsdk:"registration_id"`
	SenderID                  types.String                                  `tfsdk:"sender_id"`
	SenderIDARN               types.String                                  `tfsdk:"arn"`
	Tags                      tftags.Map                                    `tfsdk:"tags"`
}
