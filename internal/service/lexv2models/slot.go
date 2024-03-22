// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Slot")
func newResourceSlot(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceSlot{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameSlot = "Slot"

	slotIDPartCount = 5
)

type resourceSlot struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceSlot) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_lexv2models_slot"
}

func (r *resourceSlot) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	multValueSettingsLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[MultipleValuesSettingData](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"allow_multiple_values": schema.BoolAttribute{
					Optional: true,
				},
			},
		},
	}

	obfuscationSettingLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ObfuscationSettingData](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"obfuscation_setting_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.ObfuscationSettingType](),
					Required:   true,
				},
			},
		},
	}

	defaultValueSpecificationLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[DefaultValueSpecificationData](ctx),
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"default_value_list": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[DefaultValueData](ctx),
					Validators: []validator.List{
						listvalidator.IsRequired(),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"default_value": schema.StringAttribute{
								Required: true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(1, 202),
								},
							},
						},
					},
				},
			},
		},
	}

	messageNBO := schema.NestedBlockObject{
		Blocks: map[string]schema.Block{
			"custom_payload": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[CustomPayload](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"value": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"image_response_card": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[ImageResponseCard](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"image_url": schema.StringAttribute{
							Optional: true,
						},
						"subtitle": schema.StringAttribute{
							Optional: true,
						},
						"title": schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"button": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[Button](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"text": schema.StringAttribute{
										Required: true,
									},
									"value": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"plain_text_message": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[PlainTextMessage](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"value": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"ssml_message": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[SSMLMessage](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"value": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}

	messageGroupLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[MessageGroup](ctx),
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"message": schema.ListNestedBlock{
					Validators: []validator.List{
						listvalidator.SizeBetween(1, 1),
					},
					CustomType:   fwtypes.NewListNestedObjectTypeOf[Message](ctx),
					NestedObject: messageNBO,
				},
				"variation": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[Message](ctx),
					NestedObject: messageNBO,
				},
			},
		},
	}

	allowedInputTypesLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[AllowedInputTypes](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"allow_audio_input": schema.BoolAttribute{
					Required: true,
				},
				"allow_dtmf_input": schema.BoolAttribute{
					Required: true,
				},
			},
		},
	}

	audioSpecificationLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[AudioSpecification](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"end_timeout_ms": schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
				"max_length_ms": schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
			},
		},
	}

	dmfSpecificationLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[DTMFSpecification](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"deletion_character": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.RegexMatches(
							regexache.MustCompile(`^[A-D0-9#*]{1}$`),
							"alphanumeric characters",
						),
					},
				},
				"end_character": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.RegexMatches(
							regexache.MustCompile(`^[A-D0-9#*]{1}$`),
							"alphanumeric characters",
						),
					},
				},
				"end_timeout_ms": schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
				"max_length": schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						int64validator.Between(1, 1024),
					},
				},
			},
		},
	}

	audioAndDTMFInputSpecificationLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[AudioAndDTMFInputSpecification](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"start_timeout_ms": schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"audio_specification": audioSpecificationLNB,
				"dtmf_specification":  dmfSpecificationLNB,
			},
		},
	}

	textInputSpecificationLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[TextInputSpecification](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"start_timeout_ms": schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
			},
		},
	}

	promptAttemptsSpecificationLNB := schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[PromptAttemptsSpecification](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"map_block_key": schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[PromptAttemptsType](),
				},
				"allow_interrupt": schema.BoolAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"allowed_input_types":                allowedInputTypesLNB,
				"audio_and_dtmf_input_specification": audioAndDTMFInputSpecificationLNB,
				"text_input_specification":           textInputSpecificationLNB,
			},
		},
	}

	promptSpecificationLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[PromptSpecification](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"allow_interrupt": schema.BoolAttribute{
					Optional: true,
				},
				"max_retries": schema.Int64Attribute{
					Required: true,
				},
				"message_selection_strategy": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						enum.FrameworkValidate[awstypes.MessageSelectionStrategy](),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"message_group":                 messageGroupLNB,
				"prompt_attempts_specification": promptAttemptsSpecificationLNB,
			},
		},
	}

	sampleUtteranceLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[SampleUtterance](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"utterance": schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

	slotResolutionSettingLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[SlotResolutionSettingData](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"slot_resolution_strategy": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.SlotResolutionStrategy](),
					Required:   true,
				},
			},
		},
	}

	responseSpecificationLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ResponseSpecification](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"allow_interrupt": schema.BoolAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"message_group": messageGroupLNB,
			},
		},
	}

	stillWaitingResponseSpecificationLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[StillWaitingResponseSpecificationData](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"allow_interrupt": schema.BoolAttribute{
					Optional: true,
				},
				"frequency_in_seconds": schema.Int64Attribute{
					Required: true,
				},
				"timeout_in_seconds": schema.Int64Attribute{
					Required: true,
				},
			},
			Blocks: map[string]schema.Block{
				"message_group": messageGroupLNB,
			},
		},
	}

	waitAndContinueSpecificationLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[WaitAndContinueSpecificationData](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"active": schema.BoolAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"continue_response":      responseSpecificationLNB,
				"still_waiting_response": stillWaitingResponseSpecificationLNB,
				"waiting_response":       responseSpecificationLNB,
			},
		},
	}

	valueElicitationSettingLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ValueElicitationSettingData](ctx),
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"slot_constraint": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						enum.FrameworkValidate[awstypes.SlotConstraint](),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"default_value_specification":     defaultValueSpecificationLNB,
				"prompt_specification":            promptSpecificationLNB,
				"sample_utterance":                sampleUtteranceLNB,
				"slot_resolution_setting":         slotResolutionSettingLNB,
				"wait_and_continue_specification": waitAndContinueSpecificationLNB,
			},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bot_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bot_version": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
			"intent_id": schema.StringAttribute{
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
			"slot_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"slot_type_id": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"multiple_values_setting":   multValueSettingsLNB,
			"obfuscation_setting":       obfuscationSettingLNB,
			"value_elicitation_setting": valueElicitationSettingLNB,
			//sub_slot_setting
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceSlot) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var plan resourceSlotData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lexmodelsv2.CreateSlotInput{
		SlotName: aws.String(plan.Name.ValueString()),
	}

	resp.Diagnostics.Append(flex.Expand(context.WithValue(ctx, flex.ResourcePrefix, ResNameSlot), &plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateSlot(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameSlot, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameSlot, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	idParts := []string{
		aws.ToString(out.BotId),
		aws.ToString(out.BotVersion),
		aws.ToString(out.IntentId),
		aws.ToString(out.LocaleId),
		aws.ToString(out.SlotId),
	}
	id, err := intflex.FlattenResourceId(idParts, slotIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameSlot, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(id)

	resp.Diagnostics.Append(flex.Flatten(context.WithValue(ctx, flex.ResourcePrefix, ResNameSlot), out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceSlot) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var state resourceSlotData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findSlotByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionSetting, ResNameSlot, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(context.WithValue(ctx, flex.ResourcePrefix, ResNameSlot), out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceSlot) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var plan, state resourceSlotData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if slotHasChanges(ctx, plan, state) {
		input := &lexmodelsv2.UpdateSlotInput{}

		resp.Diagnostics.Append(flex.Expand(context.WithValue(ctx, flex.ResourcePrefix, ResNameSlot), plan, input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateSlot(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameSlot, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameSlot, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(context.WithValue(ctx, flex.ResourcePrefix, ResNameSlot), input, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceSlot) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var state resourceSlotData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lexmodelsv2.DeleteSlotInput{
		BotId:      aws.String(state.BotID.ValueString()),
		BotVersion: aws.String(state.BotVersion.ValueString()),
		IntentId:   aws.String(state.IntentID.ValueString()),
		LocaleId:   aws.String(state.LocaleID.ValueString()),
		SlotId:     aws.String(state.SlotID.ValueString()),
	}

	_, err := conn.DeleteSlot(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if errs.IsAErrorMessageContains[*awstypes.PreconditionFailedException](err, "does not exist") {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionDeleting, ResNameSlot, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findSlotByID(ctx context.Context, conn *lexmodelsv2.Client, id string) (*lexmodelsv2.DescribeSlotOutput, error) {
	parts, err := intflex.ExpandResourceId(id, slotIDPartCount, false)
	if err != nil {
		return nil, err
	}

	in := &lexmodelsv2.DescribeSlotInput{
		BotId:      aws.String(parts[0]),
		BotVersion: aws.String(parts[1]),
		IntentId:   aws.String(parts[2]),
		LocaleId:   aws.String(parts[3]),
		SlotId:     aws.String(parts[4]),
	}

	out, err := conn.DescribeSlot(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceSlotData struct {
	BotID                   types.String                                                 `tfsdk:"bot_id"`
	BotVersion              types.String                                                 `tfsdk:"bot_version"`
	Description             types.String                                                 `tfsdk:"description"`
	ID                      types.String                                                 `tfsdk:"id"`
	IntentID                types.String                                                 `tfsdk:"intent_id"`
	LocaleID                types.String                                                 `tfsdk:"locale_id"`
	SlotID                  types.String                                                 `tfsdk:"slot_id"`
	MultipleValuesSetting   fwtypes.ListNestedObjectValueOf[MultipleValuesSettingData]   `tfsdk:"multiple_values_setting"`
	Name                    types.String                                                 `tfsdk:"name"`
	ObfuscationSetting      fwtypes.ListNestedObjectValueOf[ObfuscationSettingData]      `tfsdk:"obfuscation_setting"`
	Timeouts                timeouts.Value                                               `tfsdk:"timeouts"`
	SlotTypeID              types.String                                                 `tfsdk:"slot_type_id"`
	ValueElicitationSetting fwtypes.ListNestedObjectValueOf[ValueElicitationSettingData] `tfsdk:"value_elicitation_setting"`
}

type MultipleValuesSettingData struct {
	AllowMultipleValues types.Bool `tfsdk:"allow_multiple_values"`
}

type ObfuscationSettingData struct {
	ObfuscationSettingType fwtypes.StringEnum[awstypes.ObfuscationSettingType] `tfsdk:"obfuscation_setting_type"`
}

type DefaultValueSpecificationData struct {
	DefaultValueList fwtypes.ListNestedObjectValueOf[DefaultValueData] `tfsdk:"default_value_list"`
}

type DefaultValueData struct {
	DefaultValue types.String `tfsdk:"default_value"`
}

type SlotResolutionSettingData struct {
	SlotResolutionStrategy fwtypes.StringEnum[awstypes.SlotResolutionStrategy] `tfsdk:"slot_resolution_strategy"`
}

type StillWaitingResponseSpecificationData struct {
	AllowInterrupt     types.Bool                                    `tfsdk:"allow_interrupt"`
	FrequencyInSeconds types.Int64                                   `tfsdk:"frequency_in_seconds"`
	MessageGroup       fwtypes.ListNestedObjectValueOf[MessageGroup] `tfsdk:"message_group"`
	TimeoutInSeconds   types.Int64                                   `tfsdk:"timeout_in_seconds"`
}

type WaitAndContinueSpecificationData struct {
	Active               types.Bool                                                             `tfsdk:"active"`
	ContinueResponse     fwtypes.ListNestedObjectValueOf[ResponseSpecification]                 `tfsdk:"continue_response"`
	StillWaitingResponse fwtypes.ListNestedObjectValueOf[StillWaitingResponseSpecificationData] `tfsdk:"still_waiting_response"`
	WaitingResponse      fwtypes.ListNestedObjectValueOf[ResponseSpecification]                 `tfsdk:"waiting_response"`
}

type ValueElicitationSettingData struct {
	SlotConstraint               fwtypes.StringEnum[awstypes.SlotConstraint]                       `tfsdk:"slot_constraint"`
	DefaultValueSpecification    fwtypes.ListNestedObjectValueOf[DefaultValueSpecificationData]    `tfsdk:"default_value_specification"`
	PromptSpecification          fwtypes.ListNestedObjectValueOf[PromptSpecification]              `tfsdk:"prompt_specification"`
	SampleUtterance              fwtypes.ListNestedObjectValueOf[SampleUtterance]                  `tfsdk:"sample_utterance"`
	SlotResolutionSetting        fwtypes.ListNestedObjectValueOf[SlotResolutionSettingData]        `tfsdk:"slot_resolution_setting"`
	WaitAndContinueSpecification fwtypes.ListNestedObjectValueOf[WaitAndContinueSpecificationData] `tfsdk:"wait_and_continue_specification"`
}

func slotHasChanges(_ context.Context, plan, state resourceSlotData) bool {
	return !plan.Description.Equal(state.Description) ||
		!plan.MultipleValuesSetting.Equal(state.MultipleValuesSetting) ||
		!plan.SlotTypeID.Equal(state.SlotTypeID)
}
