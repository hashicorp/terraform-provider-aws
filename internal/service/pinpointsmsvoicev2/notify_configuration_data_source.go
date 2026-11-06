// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_pinpointsmsvoicev2_notify_configuration", name="Notify Configuration")
// @Tags(identifierAttribute="arn")
// @Testing(generator="testAccRandomNotifyConfigurationDisplayName(t)")
func newNotifyConfigurationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &notifyConfigurationDataSource{}, nil
}

type notifyConfigurationDataSource struct {
	framework.DataSourceWithModel[notifyConfigurationDataSourceModel]
}

func (d *notifyConfigurationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"created_timestamp": schema.StringAttribute{
				Description: "RFC3339 creation timestamp of the notify configuration.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			"default_template_id": schema.StringAttribute{
				Description: "Default template ID if not specified at send time.",
				Computed:    true,
			},
			"deletion_protection_enabled": schema.BoolAttribute{
				Description: "Whether deletion protection is enabled.",
				Computed:    true,
			},
			names.AttrDisplayName: schema.StringAttribute{
				Description: "Display name (1–15 chars).",
				Computed:    true,
			},
			"enabled_channels": schema.SetAttribute{
				Description: "Communication channels enabled for this config. Valid values: SMS, VOICE, MMS, RCS.",
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"enabled_countries": schema.SetAttribute{
				Description: "List of two-letter ISO country codes.",
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrID: schema.StringAttribute{
				Description: "The Notify Configuration ID.",
				Required:    true,
			},
			"pool_id": schema.StringAttribute{
				Description: "ID or ARN of phone number pool associated with this configuration.",
				Computed:    true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"use_case": schema.StringAttribute{
				Description: "The use case (e.g. CODE_VERIFICATION).",
				Computed:    true,
			},
		},
	}
}

func (d *notifyConfigurationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data notifyConfigurationDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().PinpointSMSVoiceV2Client(ctx)
	id := data.ID.ValueString()

	out, err := findNotifyConfigurationByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, data.flatten(ctx, out))
	if resp.Diagnostics.HasError() {
		return
	}

	tags, err := listTags(ctx, conn, data.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ARN.ValueString())
		return
	}
	data.Tags = tftags.FlattenStringValueMap(ctx, tags.Map())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type notifyConfigurationDataSourceModel struct {
	framework.WithRegionModel
	ARN                       types.String        `tfsdk:"arn"`
	CreatedTimestamp          timetypes.RFC3339   `tfsdk:"created_timestamp"`
	DefaultTemplateID         types.String        `tfsdk:"default_template_id"`
	DeletionProtectionEnabled types.Bool          `tfsdk:"deletion_protection_enabled"`
	DisplayName               types.String        `tfsdk:"display_name"`
	EnabledChannels           fwtypes.SetOfString `tfsdk:"enabled_channels"`
	EnabledCountries          fwtypes.SetOfString `tfsdk:"enabled_countries"`
	ID                        types.String        `tfsdk:"id"`
	PoolID                    types.String        `tfsdk:"pool_id"`
	Tags                      tftags.Map          `tfsdk:"tags"`
	UseCase                   types.String        `tfsdk:"use_case"`
}

func (m *notifyConfigurationDataSourceModel) flatten(ctx context.Context, info *awstypes.NotifyConfigurationInformation) diag.Diagnostics {
	var diags diag.Diagnostics

	if info == nil {
		return diags
	}

	m.ID = types.StringValue(aws.ToString(info.NotifyConfigurationId))
	m.ARN = types.StringValue(aws.ToString(info.NotifyConfigurationArn))
	m.DisplayName = types.StringValue(aws.ToString(info.DisplayName))
	m.UseCase = types.StringValue(string(info.UseCase))
	m.EnabledChannels = fwtypes.SetValueOf[basetypes.StringValue]{SetValue: fwflex.FlattenFrameworkStringValueSet(ctx, info.EnabledChannels)}
	m.EnabledCountries = fwflex.FlattenFrameworkStringValueSetOfString(ctx, info.EnabledCountries)
	m.DefaultTemplateID = fwflex.StringToFramework(ctx, info.DefaultTemplateId)
	m.PoolID = fwflex.StringToFramework(ctx, info.PoolId)
	m.DeletionProtectionEnabled = types.BoolValue(info.DeletionProtectionEnabled)

	if info.CreatedTimestamp != nil {
		m.CreatedTimestamp = timetypes.NewRFC3339TimeValue(*info.CreatedTimestamp)
	} else {
		m.CreatedTimestamp = timetypes.NewRFC3339Null()
	}

	return diags
}
