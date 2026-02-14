// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package lexv2models

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lexv2models_bot_locale", name="Bot Locale")
func newBotLocaleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &botLocaleResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type botLocaleResource struct {
	framework.ResourceWithModel[botLocaleResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *botLocaleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
			names.AttrID: framework.IDAttribute(),
			"locale_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
				CustomType: fwtypes.NewListNestedObjectTypeOf[voiceSettingsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrEngine: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.VoiceEngine](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"voice_id": schema.StringAttribute{
							Required: true,
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
	botLocaleResourceIDPartCount = 3
)

func (r *botLocaleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data botLocaleResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	var input lexmodelsv2.CreateBotLocaleInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateBotLocale(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Lex v2 Bot Locale", err.Error())

		return
	}

	// Set values for unknowns.
	localeID, botID, botVersion := aws.ToString(output.LocaleId), aws.ToString(output.BotId), aws.ToString(output.BotVersion)
	id, _ := intflex.FlattenResourceId([]string{localeID, botID, botVersion}, botLocaleResourceIDPartCount, false)
	data.ID = fwflex.StringValueToFramework(ctx, id)
	data.LocaleName = fwflex.StringToFramework(ctx, output.LocaleName)
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.VoiceSettings, &data.VoiceSettings)...)
	if response.Diagnostics.HasError() {
		return
	}

	if _, err := waitBotLocaleCreated(ctx, conn, localeID, botID, botVersion, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.BotID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Lex v2 Bot Locale (%s) create", id), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *botLocaleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data botLocaleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	parts, err := intflex.ExpandResourceId(id, botLocaleResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	localeID, botID, botVersion := parts[0], parts[1], parts[2]
	output, err := findBotLocaleByThreePartKey(ctx, conn, localeID, botID, botVersion)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Lex v2 Bot Locale (%s)", id), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *botLocaleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old botLocaleResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, new.ID)
	parts, err := intflex.ExpandResourceId(id, botLocaleResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	localeID, botID, botVersion := parts[0], parts[1], parts[2]

	if !new.BotID.Equal(old.BotID) ||
		!new.BotVersion.Equal(old.BotVersion) ||
		!new.Description.Equal(old.Description) ||
		!new.LocaleID.Equal(old.LocaleID) ||
		!new.LocaleName.Equal(old.LocaleName) ||
		!new.NLUIntentConfidenceThreshold.Equal(old.NLUIntentConfidenceThreshold) ||
		!new.VoiceSettings.Equal(old.VoiceSettings) {
		var input lexmodelsv2.UpdateBotLocaleInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateBotLocale(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Lex v2 Bot Locale (%s)", id), err.Error())

			return
		}

		if _, err := waitBotLocaleUpdated(ctx, conn, localeID, botID, botVersion, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Lex v2 Bot Locale (%s) update", id), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *botLocaleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data botLocaleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	parts, err := intflex.ExpandResourceId(id, botLocaleResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	localeID, botID, botVersion := parts[0], parts[1], parts[2]
	input := lexmodelsv2.DeleteBotLocaleInput{
		BotId:      aws.String(botID),
		BotVersion: aws.String(botVersion),
		LocaleId:   aws.String(localeID),
	}
	_, err = conn.DeleteBotLocale(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
		errs.IsAErrorMessageContains[*awstypes.PreconditionFailedException](err, "does not exist") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Lex v2 Bot Locale (%s)", id), err.Error())

		return
	}

	if _, err := waitBotLocaleDeleted(ctx, conn, localeID, botID, botVersion, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Lex v2 Bot Locale (%s) delete", id), err.Error())

		return
	}
}

func findBotLocaleByThreePartKey(ctx context.Context, conn *lexmodelsv2.Client, localeID, botID, botVersion string) (*lexmodelsv2.DescribeBotLocaleOutput, error) {
	input := lexmodelsv2.DescribeBotLocaleInput{
		BotId:      aws.String(botID),
		BotVersion: aws.String(botVersion),
		LocaleId:   aws.String(localeID),
	}
	output, err := conn.DescribeBotLocale(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.LocaleId == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusBotLocale(conn *lexmodelsv2.Client, localeID, botID, botVersion string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findBotLocaleByThreePartKey(ctx, conn, localeID, botID, botVersion)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.BotLocaleStatus), nil
	}
}

func waitBotLocaleCreated(ctx context.Context, conn *lexmodelsv2.Client, localeID, botID, botVersion string, timeout time.Duration) (*lexmodelsv2.DescribeBotLocaleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.BotLocaleStatusCreating),
		Target:                    enum.Slice(awstypes.BotLocaleStatusBuilt, awstypes.BotLocaleStatusNotBuilt),
		Refresh:                   statusBotLocale(conn, localeID, botID, botVersion),
		Timeout:                   timeout,
		MinTimeout:                5 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lexmodelsv2.DescribeBotLocaleOutput); ok {
		retry.SetLastError(err, botFailureReasons(output.FailureReasons))

		return output, err
	}

	return nil, err
}

func waitBotLocaleUpdated(ctx context.Context, conn *lexmodelsv2.Client, localeID, botID, botVersion string, timeout time.Duration) (*lexmodelsv2.DescribeBotLocaleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.BotLocaleStatusBuilding),
		Target:                    enum.Slice(awstypes.BotLocaleStatusBuilt, awstypes.BotLocaleStatusNotBuilt),
		Refresh:                   statusBotLocale(conn, localeID, botID, botVersion),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lexmodelsv2.DescribeBotLocaleOutput); ok {
		retry.SetLastError(err, botFailureReasons(output.FailureReasons))

		return output, err
	}

	return nil, err
}

func waitBotLocaleDeleted(ctx context.Context, conn *lexmodelsv2.Client, localeID, botID, botVersion string, timeout time.Duration) (*lexmodelsv2.DescribeBotLocaleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BotLocaleStatusDeleting),
		Target:  []string{},
		Refresh: statusBotLocale(conn, localeID, botID, botVersion),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lexmodelsv2.DescribeBotLocaleOutput); ok {
		retry.SetLastError(err, botFailureReasons(output.FailureReasons))

		return output, err
	}

	return nil, err
}

type botLocaleResourceModel struct {
	framework.WithRegionModel
	BotID                        types.String                                        `tfsdk:"bot_id"`
	BotVersion                   types.String                                        `tfsdk:"bot_version"`
	Description                  types.String                                        `tfsdk:"description"`
	ID                           types.String                                        `tfsdk:"id"`
	LocaleID                     types.String                                        `tfsdk:"locale_id"`
	LocaleName                   types.String                                        `tfsdk:"name"`
	NLUIntentConfidenceThreshold types.Float64                                       `tfsdk:"n_lu_intent_confidence_threshold"`
	Timeouts                     timeouts.Value                                      `tfsdk:"timeouts"`
	VoiceSettings                fwtypes.ListNestedObjectValueOf[voiceSettingsModel] `tfsdk:"voice_settings"`
}

type voiceSettingsModel struct {
	Engine  fwtypes.StringEnum[awstypes.VoiceEngine] `tfsdk:"engine"`
	VoiceID types.String                             `tfsdk:"voice_id"`
}
