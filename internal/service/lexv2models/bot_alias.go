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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lexv2models_bot_alias", name="Bot Alias")
// @Tags(identifierAttribute="arn")
func newBotAliasResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &botAliasResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type botAliasResource struct {
	framework.ResourceWithModel[botAliasResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *botAliasResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"bot_alias_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bot_alias_name": schema.StringAttribute{
				Required: true,
			},
			"bot_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bot_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID:      framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"bot_alias_locale_settings": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[botAliasLocaleSettingsModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrEnabled: schema.BoolAttribute{
							Required: true,
						},
						"locale_id": schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"code_hook_specification": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[codeHookSpecificationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"lambda_code_hook": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[lambdaCodeHookModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"code_hook_interface_version": schema.StringAttribute{
													Required: true,
												},
												"lambda_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"conversation_log_settings": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[conversationLogSettingsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"audio_log_settings": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[audioLogSettingModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrEnabled: schema.BoolAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									names.AttrDestination: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[audioLogDestinationModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												names.AttrS3Bucket: schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3BucketLogDestinationModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtLeast(1),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrKMSKeyARN: schema.StringAttribute{
																CustomType: fwtypes.ARNType,
																Optional:   true,
															},
															"log_prefix": schema.StringAttribute{
																Required: true,
															},
															"s3_bucket_arn": schema.StringAttribute{
																CustomType: fwtypes.ARNType,
																Required:   true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"text_log_settings": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[textLogSettingModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrEnabled: schema.BoolAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									names.AttrDestination: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[textLogDestinationModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"cloudwatch": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[cloudWatchLogGroupLogDestinationModel](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
														listvalidator.SizeAtLeast(1),
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrCloudWatchLogGroupARN: schema.StringAttribute{
																CustomType: fwtypes.ARNType,
																Required:   true,
															},
															"log_prefix": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"sentiment_analysis_settings": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[sentimentAnalysisSettingsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"detect_sentiment": schema.BoolAttribute{
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
	botAliasResourceIDPartCount = 2
)

func (r *botAliasResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data botAliasResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	var input lexmodelsv2.CreateBotAliasInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateBotAlias(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Lex v2 Bot Alias (%s)", data.BotAliasName.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	botID, botAliasID := aws.ToString(output.BotId), aws.ToString(output.BotAliasId)
	id, _ := intflex.FlattenResourceId([]string{botID, botAliasID}, botAliasResourceIDPartCount, false)
	data.ARN = fwflex.StringValueToFramework(ctx, r.botAliasARN(ctx, botID, botAliasID))
	data.BotAliasID = fwflex.StringValueToFramework(ctx, botAliasID)
	data.BotVersion = fwflex.StringToFramework(ctx, output.BotVersion)
	data.ID = fwflex.StringValueToFramework(ctx, id)

	if _, err := waitBotAliasCreated(ctx, conn, botID, botAliasID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Lex v2 Bot Alias (%s) create", id), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *botAliasResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data botAliasResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	parts, err := intflex.ExpandResourceId(id, botAliasResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	botID, botAliasID := parts[0], parts[1]
	output, err := findBotAliasByTwoPartKey(ctx, conn, botID, botAliasID)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Lex v2 Bot Alias (%s)", id), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.ARN = fwflex.StringValueToFramework(ctx, r.botAliasARN(ctx, botID, botAliasID))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *botAliasResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old botAliasResourceModel
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
	parts, err := intflex.ExpandResourceId(id, botAliasResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	botID, botAliasID := parts[0], parts[1]

	if !new.BotAliasLocaleSettings.Equal(old.BotAliasLocaleSettings) ||
		!new.BotAliasName.Equal(old.BotAliasName) ||
		!new.BotVersion.Equal(old.BotVersion) ||
		!new.ConversationLogSettings.Equal(old.ConversationLogSettings) ||
		!new.Description.Equal(old.Description) ||
		!new.SentimentAnalysisSettings.Equal(old.SentimentAnalysisSettings) {
		var input lexmodelsv2.UpdateBotAliasInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateBotAlias(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Lex v2 Bot Alias (%s)", id), err.Error())

			return
		}

		if _, err := waitBotAliasUpdated(ctx, conn, botID, botAliasID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Lex v2 Bot Alias (%s) update", id), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *botAliasResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data botAliasResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LexV2ModelsClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	parts, err := intflex.ExpandResourceId(id, botAliasResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	botID, botAliasID := parts[0], parts[1]
	input := lexmodelsv2.DeleteBotAliasInput{
		BotAliasId: aws.String(botAliasID),
		BotId:      aws.String(botID),
	}
	_, err = conn.DeleteBotAlias(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
		errs.IsAErrorMessageContains[*awstypes.PreconditionFailedException](err, "does not exist") {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Lex v2 Bot Alias (%s)", id), err.Error())

		return
	}

	if _, err := waitBotAliasDeleted(ctx, conn, botID, botAliasID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Lex v2 Bot Alias (%s) delete", id), err.Error())

		return
	}
}

func (r *botAliasResource) botAliasARN(ctx context.Context, botID, botAliasID string) string {
	return r.Meta().RegionalARN(ctx, "lex", fmt.Sprintf("bot-alias/%s/%s", botID, botAliasID))
}

func findBotAliasByTwoPartKey(ctx context.Context, conn *lexmodelsv2.Client, botID, botAliasID string) (*lexmodelsv2.DescribeBotAliasOutput, error) {
	input := lexmodelsv2.DescribeBotAliasInput{
		BotAliasId: aws.String(botAliasID),
		BotId:      aws.String(botID),
	}
	output, err := conn.DescribeBotAlias(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.BotAliasId == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusBotAlias(conn *lexmodelsv2.Client, botID, botAliasID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findBotAliasByTwoPartKey(ctx, conn, botID, botAliasID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.BotAliasStatus), nil
	}
}

func waitBotAliasCreated(ctx context.Context, conn *lexmodelsv2.Client, botID, botAliasID string, timeout time.Duration) (*lexmodelsv2.DescribeBotAliasOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.BotAliasStatusCreating),
		Target:                    enum.Slice(awstypes.BotAliasStatusAvailable),
		Refresh:                   statusBotAlias(conn, botID, botAliasID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lexmodelsv2.DescribeBotAliasOutput); ok {
		return output, err
	}

	return nil, err
}

func waitBotAliasUpdated(ctx context.Context, conn *lexmodelsv2.Client, botID, botAliasID string, timeout time.Duration) (*lexmodelsv2.DescribeBotAliasOutput, error) {
	// The AWS API for Bot Alias does not expose a distinct "Updating" status; once
	// UpdateBotAlias returns the alias is briefly back in Creating before it
	// settles on Available. Treat Creating as the pending state and Available as
	// the target.
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.BotAliasStatusCreating),
		Target:                    enum.Slice(awstypes.BotAliasStatusAvailable),
		Refresh:                   statusBotAlias(conn, botID, botAliasID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lexmodelsv2.DescribeBotAliasOutput); ok {
		return output, err
	}

	return nil, err
}

func waitBotAliasDeleted(ctx context.Context, conn *lexmodelsv2.Client, botID, botAliasID string, timeout time.Duration) (*lexmodelsv2.DescribeBotAliasOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BotAliasStatusDeleting),
		Target:  []string{},
		Refresh: statusBotAlias(conn, botID, botAliasID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lexmodelsv2.DescribeBotAliasOutput); ok {
		return output, err
	}

	return nil, err
}

type botAliasResourceModel struct {
	framework.WithRegionModel
	ARN                       types.String                                                    `tfsdk:"arn"`
	BotAliasID                types.String                                                    `tfsdk:"bot_alias_id"`
	BotAliasLocaleSettings    fwtypes.SetNestedObjectValueOf[botAliasLocaleSettingsModel]     `tfsdk:"bot_alias_locale_settings"`
	BotAliasName              types.String                                                    `tfsdk:"bot_alias_name"`
	BotID                     types.String                                                    `tfsdk:"bot_id"`
	BotVersion                types.String                                                    `tfsdk:"bot_version"`
	ConversationLogSettings   fwtypes.ListNestedObjectValueOf[conversationLogSettingsModel]   `tfsdk:"conversation_log_settings"`
	Description               types.String                                                    `tfsdk:"description"`
	ID                        types.String                                                    `tfsdk:"id"`
	SentimentAnalysisSettings fwtypes.ListNestedObjectValueOf[sentimentAnalysisSettingsModel] `tfsdk:"sentiment_analysis_settings"`
	Tags                      tftags.Map                                                      `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                      `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                                  `tfsdk:"timeouts"`
}

type botAliasLocaleSettingsModel struct {
	CodeHookSpecification fwtypes.ListNestedObjectValueOf[codeHookSpecificationModel] `tfsdk:"code_hook_specification"`
	Enabled               types.Bool                                                  `tfsdk:"enabled"`
	MapBlockKey           types.String                                                `tfsdk:"locale_id"`
}

type codeHookSpecificationModel struct {
	LambdaCodeHook fwtypes.ListNestedObjectValueOf[lambdaCodeHookModel] `tfsdk:"lambda_code_hook"`
}

type lambdaCodeHookModel struct {
	CodeHookInterfaceVersion types.String `tfsdk:"code_hook_interface_version"`
	LambdaARN                fwtypes.ARN  `tfsdk:"lambda_arn"`
}

type conversationLogSettingsModel struct {
	AudioLogSettings fwtypes.SetNestedObjectValueOf[audioLogSettingModel] `tfsdk:"audio_log_settings"`
	TextLogSettings  fwtypes.SetNestedObjectValueOf[textLogSettingModel]  `tfsdk:"text_log_settings"`
}

type audioLogSettingModel struct {
	Destination fwtypes.ListNestedObjectValueOf[audioLogDestinationModel] `tfsdk:"destination"`
	Enabled     types.Bool                                                `tfsdk:"enabled"`
}

type audioLogDestinationModel struct {
	S3Bucket fwtypes.ListNestedObjectValueOf[s3BucketLogDestinationModel] `tfsdk:"s3_bucket"`
}

type s3BucketLogDestinationModel struct {
	KmsKeyArn   fwtypes.ARN  `tfsdk:"kms_key_arn"`
	LogPrefix   types.String `tfsdk:"log_prefix"`
	S3BucketArn fwtypes.ARN  `tfsdk:"s3_bucket_arn"`
}

type textLogSettingModel struct {
	Destination fwtypes.ListNestedObjectValueOf[textLogDestinationModel] `tfsdk:"destination"`
	Enabled     types.Bool                                               `tfsdk:"enabled"`
}

type textLogDestinationModel struct {
	CloudWatch fwtypes.ListNestedObjectValueOf[cloudWatchLogGroupLogDestinationModel] `tfsdk:"cloudwatch"`
}

type cloudWatchLogGroupLogDestinationModel struct {
	CloudWatchLogGroupArn fwtypes.ARN  `tfsdk:"cloudwatch_log_group_arn"`
	LogPrefix             types.String `tfsdk:"log_prefix"`
}

type sentimentAnalysisSettingsModel struct {
	DetectSentiment types.Bool `tfsdk:"detect_sentiment"`
}
