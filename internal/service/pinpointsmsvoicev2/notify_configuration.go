// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_pinpointsmsvoicev2_notify_configuration", name="Notify Configuration")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckNotifyConfiguration")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types;awstypes;awstypes.NotifyConfigurationInformation")
// @Testing(generator="testAccRandomNotifyConfigurationDisplayName(t)")
func newNotifyConfigurationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &notifyConfigurationResource{}

	return r, nil
}

type notifyConfigurationResource struct {
	framework.ResourceWithModel[notifyConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *notifyConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"created_timestamp": schema.StringAttribute{
				Description: "RFC3339 creation timestamp of the notify configuration.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"default_template_id": schema.StringAttribute{
				Description: "Default template ID if not specified at send time.",
				Optional:    true,
			},
			"deletion_protection_enabled": schema.BoolAttribute{
				Description: "Protect against accidental deletion. When set to true, the notify configuration cannot be deleted.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			names.AttrDisplayName: schema.StringAttribute{
				Description: "Display name (1–15 chars).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 15),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_ -]+$`), "must contain only alphanumeric characters, underscores, hyphens, and spaces"),
				},
			},
			"enabled_channels": schema.SetAttribute{
				Description: "Communication channels enabled for this config. Valid values: SMS, VOICE, MMS, RCS.",
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 4),
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf("SMS", "VOICE", "MMS", "RCS"),
					),
				},
			},
			"enabled_countries": schema.SetAttribute{
				Description: "List of two-letter ISO country codes (e.g. ['US', 'AU']).",
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Z]{2}$`), "must be in ISO 3166-1 alpha-2 format"),
					),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"pool_id": schema.StringAttribute{
				Description: "ID or ARN of phone number pool to associate.",
				Optional:    true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"use_case": schema.StringAttribute{
				Description: "The use case (e.g. CODE_VERIFICATION).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *notifyConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notifyConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	input := pinpointsmsvoicev2.CreateNotifyConfigurationInput{
		ClientToken:     aws.String(create.UniqueId(ctx)),
		DisplayName:     plan.DisplayName.ValueStringPointer(),
		UseCase:         awstypes.NotifyConfigurationUseCase(plan.UseCase.ValueString()),
		EnabledChannels: fwflex.ExpandFrameworkStringyValueSet[awstypes.NumberCapability](ctx, plan.EnabledChannels),
		Tags:            getTagsIn(ctx),
	}

	if !plan.EnabledCountries.IsNull() && !plan.EnabledCountries.IsUnknown() {
		input.EnabledCountries = fwflex.ExpandFrameworkStringValueSet(ctx, plan.EnabledCountries)
	}

	if !plan.DefaultTemplateID.IsNull() && !plan.DefaultTemplateID.IsUnknown() {
		input.DefaultTemplateId = plan.DefaultTemplateID.ValueStringPointer()
	}

	if !plan.PoolID.IsNull() && !plan.PoolID.IsUnknown() {
		input.PoolId = plan.PoolID.ValueStringPointer()
	}

	if !plan.DeletionProtectionEnabled.IsNull() && !plan.DeletionProtectionEnabled.IsUnknown() {
		input.DeletionProtectionEnabled = plan.DeletionProtectionEnabled.ValueBoolPointer()
	}

	out, err := conn.CreateNotifyConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.DisplayName.ValueString())
		return
	}

	id := aws.ToString(out.NotifyConfigurationId)
	plan.ID = types.StringValue(id)
	plan.ARN = types.StringValue(aws.ToString(out.NotifyConfigurationArn))
	if out.CreatedTimestamp != nil {
		plan.CreatedTimestamp = timetypes.NewRFC3339TimeValue(*out.CreatedTimestamp)
	}

	info, err := findNotifyConfigurationByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, plan.flatten(ctx, info))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *notifyConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data notifyConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	out, err := findNotifyConfigurationByID(ctx, conn, data.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, data.flatten(ctx, out))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (r *notifyConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var plan, state notifyConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	var hasChange bool
	input := &pinpointsmsvoicev2.UpdateNotifyConfigurationInput{
		NotifyConfigurationId: aws.String(id),
	}

	if !plan.EnabledChannels.Equal(state.EnabledChannels) {
		hasChange = true
		input.EnabledChannels = fwflex.ExpandFrameworkStringyValueSet[awstypes.NumberCapability](ctx, plan.EnabledChannels)
	}

	if !plan.EnabledCountries.Equal(state.EnabledCountries) {
		hasChange = true
		input.EnabledCountries = fwflex.ExpandFrameworkStringValueSet(ctx, plan.EnabledCountries)
	}

	if !plan.DeletionProtectionEnabled.Equal(state.DeletionProtectionEnabled) {
		hasChange = true
		input.DeletionProtectionEnabled = plan.DeletionProtectionEnabled.ValueBoolPointer()
	}

	if !plan.DefaultTemplateID.Equal(state.DefaultTemplateID) {
		hasChange = true
		if plan.DefaultTemplateID.IsNull() {
			input.DefaultTemplateId = aws.String("UNSET_DEFAULT_TEMPLATE")
		} else {
			input.DefaultTemplateId = plan.DefaultTemplateID.ValueStringPointer()
		}
	}

	if !plan.PoolID.Equal(state.PoolID) {
		hasChange = true
		if plan.PoolID.IsNull() {
			input.PoolId = aws.String("UNSET_DEFAULT_POOL_FOR_NOTIFY")
		} else {
			input.PoolId = plan.PoolID.ValueStringPointer()
		}
	}

	if hasChange {
		_, err := conn.UpdateNotifyConfiguration(ctx, input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	out, err := findNotifyConfigurationByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, plan.flatten(ctx, out))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *notifyConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notifyConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	if state.DeletionProtectionEnabled.ValueBool() {
		resp.Diagnostics.AddError(
			"cannot delete End User Messaging SMS Notify Configuration with deletion_protection_enabled = true",
			fmt.Sprintf("To delete Notify Configuration (%s), deletion_protection_enabled must first be set to false and applied.", state.ID.ValueString()),
		)
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	_, err := conn.DeleteNotifyConfiguration(ctx, &pinpointsmsvoicev2.DeleteNotifyConfigurationInput{
		NotifyConfigurationId: state.ID.ValueStringPointer(),
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}
}

type notifyConfigurationResourceModel struct {
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
	TagsAll                   tftags.Map          `tfsdk:"tags_all"`
	UseCase                   types.String        `tfsdk:"use_case"`
}

func (m *notifyConfigurationResourceModel) flatten(ctx context.Context, info *awstypes.NotifyConfigurationInformation) diag.Diagnostics {
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

func findNotifyConfigurationByID(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string) (*awstypes.NotifyConfigurationInformation, error) {
	if arn.IsARN(id) {
		if parsedARN, err := arn.Parse(id); err == nil {
			id = strings.TrimPrefix(parsedARN.Resource, "notify-configuration/")
			id = strings.TrimPrefix(id, "notify-configuration:")
		}
	}

	input := &pinpointsmsvoicev2.DescribeNotifyConfigurationsInput{
		NotifyConfigurationIds: []string{id},
	}

	return findNotifyConfiguration(ctx, conn, input)
}

func findNotifyConfiguration(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeNotifyConfigurationsInput) (*awstypes.NotifyConfigurationInformation, error) {
	output, err := findNotifyConfigurations(ctx, conn, input)
	if err != nil {
		return nil, err
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

func findNotifyConfigurations(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeNotifyConfigurationsInput) ([]awstypes.NotifyConfigurationInformation, error) {
	var output []awstypes.NotifyConfigurationInformation

	pages := pinpointsmsvoicev2.NewDescribeNotifyConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, page.NotifyConfigurations...)
	}

	return output, nil
}
