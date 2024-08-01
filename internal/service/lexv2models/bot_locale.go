// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Bot Locale")
func newResourceBotLocale(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceBotLocale{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameBotLocale = "Bot Locale"
)

type resourceBotLocale struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceBotLocale) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_lexv2models_bot_locale"
}

func (r *resourceBotLocale) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"bot_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bot_version": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"locale_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"n_lu_intent_confidence_threshold": schema.Float64Attribute{
				Required: true,
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"voice_settings": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"voice_id": schema.StringAttribute{
							Required: true,
						},
						names.AttrEngine: schema.StringAttribute{
							Optional: true,
							Computed: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.VoiceEngine](),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

const (
	botLocaleIDPartCount = 3
)

func (r *resourceBotLocale) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var plan resourceBotLocaleData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lexmodelsv2.CreateBotLocaleInput{
		BotId:                        aws.String(plan.BotID.ValueString()),
		BotVersion:                   aws.String(plan.BotVersion.ValueString()),
		LocaleId:                     aws.String(plan.LocaleID.ValueString()),
		NluIntentConfidenceThreshold: aws.Float64(plan.NluIntentCOnfidenceThreshold.ValueFloat64()),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}
	if !plan.VoiceSettings.IsNull() {
		var tfList []voiceSettingsData
		resp.Diagnostics.Append(plan.VoiceSettings.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		vsInput := expandVoiceSettings(ctx, tfList)
		in.VoiceSettings = vsInput
	}

	out, err := conn.CreateBotLocale(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameBotLocale, plan.LocaleID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.LocaleId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameBotLocale, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	idParts := []string{
		aws.ToString(out.LocaleId),
		aws.ToString(out.BotId),
		aws.ToString(out.BotVersion),
	}
	id, _ := fwflex.FlattenResourceId(idParts, botLocaleIDPartCount, false)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.LocaleID = flex.StringToFramework(ctx, out.LocaleId)
	plan.Id = types.StringValue(id)
	state := plan
	state.LocaleID = flex.StringToFramework(ctx, out.LocaleId)
	state.BotID = flex.StringToFramework(ctx, out.BotId)
	state.Id = types.StringValue(id)
	state.Description = flex.StringToFramework(ctx, out.Description)
	state.Name = flex.StringToFramework(ctx, out.LocaleName)

	vs, _ := flattenVoiceSettings(ctx, out.VoiceSettings)
	if resp.Diagnostics.HasError() {
		return
	}
	state.VoiceSettings = vs

	state.BotVersion = flex.StringValueToFramework(ctx, *out.BotVersion)
	state.NluIntentCOnfidenceThreshold = flex.Float64ToFramework(ctx, out.NluIntentConfidenceThreshold)

	createTimeout := r.CreateTimeout(ctx, state.Timeouts)
	_, err = waitBotLocaleCreated(ctx, conn, state.Id.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForCreation, ResNameBotLocale, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceBotLocale) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)
	var state resourceBotLocaleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindBotLocaleByID(ctx, conn, state.Id.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionSetting, ResNameBotLocale, state.LocaleID.String(), err),
			err.Error(),
		)
		return
	}

	state.LocaleID = flex.StringToFramework(ctx, out.LocaleId)
	state.BotID = flex.StringToFramework(ctx, out.BotId)
	state.Description = flex.StringToFramework(ctx, out.Description)
	state.BotVersion = flex.StringValueToFramework(ctx, *out.BotVersion)
	state.Name = flex.StringToFramework(ctx, out.LocaleName)
	state.NluIntentCOnfidenceThreshold = flex.Float64ToFramework(ctx, out.NluIntentConfidenceThreshold)

	vs, _ := flattenVoiceSettings(ctx, out.VoiceSettings)
	if resp.Diagnostics.HasError() {
		return
	}

	state.VoiceSettings = vs
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceBotLocale) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var plan, state resourceBotLocaleData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.BotID.Equal(state.BotID) ||
		!plan.Description.Equal(state.Description) ||
		!plan.BotVersion.Equal(state.BotVersion) ||
		!plan.LocaleID.Equal(state.LocaleID) ||
		!plan.Name.Equal(state.Name) ||
		!plan.VoiceSettings.Equal(state.VoiceSettings) ||
		!plan.NluIntentCOnfidenceThreshold.Equal(state.NluIntentCOnfidenceThreshold) {
		in := &lexmodelsv2.UpdateBotLocaleInput{
			BotId:                        aws.String(plan.BotID.ValueString()),
			BotVersion:                   aws.String(plan.BotVersion.ValueString()),
			LocaleId:                     aws.String(plan.LocaleID.ValueString()),
			NluIntentConfidenceThreshold: aws.Float64(plan.NluIntentCOnfidenceThreshold.ValueFloat64()),
		}

		if !plan.Description.IsNull() {
			in.Description = aws.String(plan.Description.ValueString())
		}
		if !plan.VoiceSettings.IsNull() {
			var tfList []voiceSettingsData
			resp.Diagnostics.Append(plan.VoiceSettings.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			in.VoiceSettings = expandVoiceSettings(ctx, tfList)
		}

		_, err := conn.UpdateBotLocale(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameBotLocale, plan.LocaleID.String(), err),
				err.Error(),
			)
			return
		}
		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		out, err := waitBotLocaleUpdated(ctx, conn, plan.Id.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForUpdate, ResNameBotLocale, plan.LocaleID.String(), err),
				err.Error(),
			)
			return
		}

		if out == nil || out.LocaleId == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameBotLocale, plan.LocaleID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		state.refreshFromOutput(ctx, out)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceBotLocale) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var state resourceBotLocaleData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lexmodelsv2.DeleteBotLocaleInput{
		LocaleId:   aws.String(state.LocaleID.ValueString()),
		BotId:      aws.String(state.BotID.ValueString()),
		BotVersion: aws.String(state.BotVersion.ValueString()),
	}

	_, err := conn.DeleteBotLocale(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionDeleting, ResNameBotLocale, state.LocaleID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitBotLocaleDeleted(ctx, conn, state.Id.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForDeletion, ResNameBotLocale, state.LocaleID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceBotLocale) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitBotLocaleCreated(ctx context.Context, conn *lexmodelsv2.Client, id string, timeout time.Duration) (*lexmodelsv2.DescribeBotLocaleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.BotLocaleStatusCreating),
		Target:                    enum.Slice(awstypes.BotLocaleStatusBuilt, awstypes.BotLocaleStatusNotBuilt),
		Refresh:                   statusBotLocale(ctx, conn, id),
		Timeout:                   timeout,
		MinTimeout:                5 * time.Second,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexmodelsv2.DescribeBotLocaleOutput); ok {
		return out, err
	}

	return nil, err
}

func waitBotLocaleUpdated(ctx context.Context, conn *lexmodelsv2.Client, id string, timeout time.Duration) (*lexmodelsv2.DescribeBotLocaleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.BotLocaleStatusBuilding),
		Target:                    enum.Slice(awstypes.BotLocaleStatusBuilt, awstypes.BotLocaleStatusNotBuilt),
		Refresh:                   statusBotLocale(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexmodelsv2.DescribeBotLocaleOutput); ok {
		return out, err
	}

	return nil, err
}

func waitBotLocaleDeleted(ctx context.Context, conn *lexmodelsv2.Client, id string, timeout time.Duration) (*lexmodelsv2.DescribeBotLocaleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BotLocaleStatusDeleting),
		Target:  []string{},
		Refresh: statusBotLocale(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexmodelsv2.DescribeBotLocaleOutput); ok {
		return out, err
	}

	return nil, err
}

func statusBotLocale(ctx context.Context, conn *lexmodelsv2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindBotLocaleByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.BotLocaleStatus)), nil
	}
}

func FindBotLocaleByID(ctx context.Context, conn *lexmodelsv2.Client, id string) (*lexmodelsv2.DescribeBotLocaleOutput, error) {
	parts, err := fwflex.ExpandResourceId(id, botLocaleIDPartCount, false)
	if err != nil {
		return nil, err
	}
	in := &lexmodelsv2.DescribeBotLocaleInput{
		LocaleId:   aws.String(parts[0]),
		BotId:      aws.String(parts[1]),
		BotVersion: aws.String(parts[2]),
	}

	out, err := conn.DescribeBotLocale(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.LocaleId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenVoiceSettings(ctx context.Context, apiObject *awstypes.VoiceSettings) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: voiceSettingsAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"voice_id":       flex.StringValueToFramework(ctx, *apiObject.VoiceId),
		names.AttrEngine: flex.StringValueToFramework(ctx, apiObject.Engine),
	}
	objVal, d := types.ObjectValue(voiceSettingsAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func expandVoiceSettings(ctx context.Context, tfList []voiceSettingsData) *awstypes.VoiceSettings {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	return &awstypes.VoiceSettings{
		VoiceId: flex.StringFromFramework(ctx, tfObj.VoiceId),
		Engine:  awstypes.VoiceEngine(tfObj.Engine.ValueString()),
	}
}

type resourceBotLocaleData struct {
	BotID                        types.String   `tfsdk:"bot_id"`
	BotVersion                   types.String   `tfsdk:"bot_version"`
	LocaleID                     types.String   `tfsdk:"locale_id"`
	Name                         types.String   `tfsdk:"name"`
	VoiceSettings                types.List     `tfsdk:"voice_settings"`
	Description                  types.String   `tfsdk:"description"`
	NluIntentCOnfidenceThreshold types.Float64  `tfsdk:"n_lu_intent_confidence_threshold"`
	Id                           types.String   `tfsdk:"id"`
	Timeouts                     timeouts.Value `tfsdk:"timeouts"`
}

type voiceSettingsData struct {
	VoiceId types.String `tfsdk:"voice_id"`
	Engine  types.String `tfsdk:"engine"`
}

var voiceSettingsAttrTypes = map[string]attr.Type{
	"voice_id":       types.StringType,
	names.AttrEngine: types.StringType,
}

// refreshFromOutput writes state data from an AWS response object
func (rd *resourceBotLocaleData) refreshFromOutput(ctx context.Context, out *lexmodelsv2.DescribeBotLocaleOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	rd.LocaleID = flex.StringToFramework(ctx, out.LocaleId)
	rd.BotID = flex.StringToFramework(ctx, out.BotId)
	rd.Description = flex.StringToFramework(ctx, out.Description)
	vs, d := flattenVoiceSettings(ctx, out.VoiceSettings)
	diags.Append(d...)
	rd.VoiceSettings = vs
	rd.BotVersion = flex.StringValueToFramework(ctx, *out.BotVersion)
	rd.Name = flex.StringToFramework(ctx, out.LocaleName)
	rd.NluIntentCOnfidenceThreshold = flex.Float64ToFramework(ctx, out.NluIntentConfidenceThreshold)

	return diags
}
