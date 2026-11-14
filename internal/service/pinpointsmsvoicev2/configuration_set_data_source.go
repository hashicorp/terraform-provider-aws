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

// @FrameworkDataSource("aws_pinpointsmsvoicev2_configuration_set", name="Configuration Set")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newConfigurationSetDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &configurationSetDataSource{}, nil
}

type configurationSetDataSource struct {
	framework.DataSourceWithModel[configurationSetDataSourceModel]
}

func (d *configurationSetDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"created_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"default_message_feedback_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"default_message_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.MessageType](),
				Computed:   true,
			},
			"default_sender_id": schema.StringAttribute{
				Computed: true,
			},
			"event_destinations": framework.DataSourceComputedListOfObjectAttribute[configurationSetEventDestinationModel](ctx),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"protect_configuration_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *configurationSetDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().PinpointSMSVoiceV2Client(ctx)

	var data configurationSetDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	name := fwflex.StringValueFromFramework(ctx, data.ConfigurationSetName)
	out, err := findConfigurationSetByID(ctx, conn, name)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data), smerr.ID, name)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data), smerr.ID, name)
}

type configurationSetDataSourceModel struct {
	framework.WithRegionModel
	ConfigurationSetARN           types.String                                                           `tfsdk:"arn"`
	ConfigurationSetName          types.String                                                           `tfsdk:"name"`
	CreatedTimestamp              timetypes.RFC3339                                                      `tfsdk:"created_timestamp"`
	DefaultMessageFeedbackEnabled types.Bool                                                             `tfsdk:"default_message_feedback_enabled"`
	DefaultMessageType            fwtypes.StringEnum[awstypes.MessageType]                               `tfsdk:"default_message_type"`
	DefaultSenderID               types.String                                                           `tfsdk:"default_sender_id"`
	EventDestinations             fwtypes.ListNestedObjectValueOf[configurationSetEventDestinationModel] `tfsdk:"event_destinations"`
	ProtectConfigurationID        types.String                                                           `tfsdk:"protect_configuration_id"`
	Tags                          tftags.Map                                                             `tfsdk:"tags"`
}

type configurationSetEventDestinationModel struct {
	CloudWatchLogsDestination  fwtypes.ListNestedObjectValueOf[cloudWatchLogsDestinationModel]  `tfsdk:"cloudwatch_logs_destination"`
	Enabled                    types.Bool                                                       `tfsdk:"enabled"`
	EventDestinationName       types.String                                                     `tfsdk:"event_destination_name"`
	KinesisFirehoseDestination fwtypes.ListNestedObjectValueOf[kinesisFirehoseDestinationModel] `tfsdk:"kinesis_firehose_destination"`
	MatchingEventTypes         fwtypes.SetOfStringEnum[awstypes.EventType]                      `tfsdk:"matching_event_types"`
	SnsDestination             fwtypes.ListNestedObjectValueOf[snsDestinationModel]             `tfsdk:"sns_destination"`
}
