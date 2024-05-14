// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Intent")
func newResourceIntent(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceIntent{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameIntent = "Intent"
)

type resourceIntent struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceIntent) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_lexv2models_intent"
}

func (r *resourceIntent) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	slotPriorityLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[SlotPriority](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrPriority: schema.Int64Attribute{
					Required: true,
				},
				"slot_id": schema.StringAttribute{
					Required: true,
				},
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

	outputContextLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(10),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[OutputContext](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrName: schema.StringAttribute{
					Required: true,
				},
				"time_to_live_in_seconds": schema.Int64Attribute{
					Required: true,
				},
				"turns_to_live": schema.Int64Attribute{
					Required: true,
				},
			},
		},
	}

	kendraConfigurationLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[KendraConfiguration](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"kendra_index": schema.StringAttribute{
					Required: true,
				},
				"query_filter_string": schema.StringAttribute{
					Optional: true,
				},
				"query_filter_string_enabled": schema.BoolAttribute{
					Optional: true,
				},
			},
		},
	}

	customPayloadLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[CustomPayload](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrValue: schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

	buttonLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[Button](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"text": schema.StringAttribute{
					Required: true,
				},
				names.AttrValue: schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

	imageResponseCardLNB := schema.ListNestedBlock{
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
				"button": buttonLNB,
			},
		},
	}

	plainTextMessageLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[PlainTextMessage](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrValue: schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

	ssmlMessageLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[SSMLMessage](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrValue: schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

	messageNBO := schema.NestedBlockObject{
		Blocks: map[string]schema.Block{
			"custom_payload":      customPayloadLNB,
			"image_response_card": imageResponseCardLNB,
			"plain_text_message":  plainTextMessageLNB,
			"ssml_message":        ssmlMessageLNB,
		},
	}

	messageGroupLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[MessageGroup](ctx),
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				names.AttrMessage: schema.ListNestedBlock{
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

	slotValueLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[SlotValue](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"interpreted_value": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
			},
		},
	}

	slotValueOverrideLNB := schema.SetNestedBlock{
		CustomType: fwtypes.NewSetNestedObjectTypeOf[SlotValueOverride](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"map_block_key": schema.StringAttribute{
					Required: true,
				},
				"shape": schema.StringAttribute{
					Optional:   true,
					CustomType: fwtypes.StringEnumType[awstypes.SlotShape](),
				},
			},
			Blocks: map[string]schema.Block{
				names.AttrValue: slotValueLNB,
			},
		},
	}

	// slotValueOverrideLNB.NestedObject.Blocks["values"] = slotValueOverrideLNB // recursive type, purposely left out, future feature

	dialogActionLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[DialogAction](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrType: schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[awstypes.DialogActionType](),
				},
				"slot_to_elicit": schema.StringAttribute{
					Optional: true,
				},
				"suppress_next_message": schema.BoolAttribute{
					Optional: true,
				},
			},
		},
	}

	intentOverrideLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[IntentOverride](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrName: schema.StringAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"slot": slotValueOverrideLNB,
			},
		},
	}

	dialogStateNBO := schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"session_attributes": schema.MapAttribute{
				ElementType: types.StringType,
				CustomType:  fwtypes.NewMapTypeOf[types.String](ctx),
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"dialog_action": dialogActionLNB,
			"intent":        intentOverrideLNB,
		},
	}

	responseSpecificationLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
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

	conditionLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[Condition](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"expression_string": schema.StringAttribute{
					Required: true,
				},
			},
		},
	}

	conditionalBranchLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[ConditionalBranch](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrName: schema.StringAttribute{
					Required: true,
				},
			},
			Blocks: map[string]schema.Block{
				names.AttrCondition: conditionLNB,
				"next_step": schema.ListNestedBlock{
					Validators: []validator.List{
						listvalidator.SizeBetween(1, 1),
					},
					CustomType:   fwtypes.NewListNestedObjectTypeOf[DialogState](ctx),
					NestedObject: dialogStateNBO,
				},
				"response": responseSpecificationLNB,
			},
		},
	}

	nextStepLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType:   fwtypes.NewListNestedObjectTypeOf[DialogState](ctx),
		NestedObject: dialogStateNBO,
	}

	defaultBranchLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[DefaultConditionalBranch](ctx),
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"next_step": nextStepLNB,
				"response":  responseSpecificationLNB,
			},
		},
	}

	conditionalSpecificationLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[ConditionalSpecification](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"active": schema.BoolAttribute{
					Required: true,
				},
			},
			Blocks: map[string]schema.Block{
				"conditional_branch": conditionalBranchLNB,
				"default_branch":     defaultBranchLNB,
			},
		},
	}

	closingSettingLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[IntentClosingSetting](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"active": schema.BoolAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"closing_response": responseSpecificationLNB,
				"conditional":      conditionalSpecificationLNB,
				"next_step":        nextStepLNB,
			},
		},
	}

	inputContextLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(5),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[InputContext](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrName: schema.StringAttribute{
					Required: true,
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

	dtmfSpecificationLNB := schema.ListNestedBlock{
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
				"dtmf_specification":  dtmfSpecificationLNB,
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
					//Min:       1,
				},
			},
		},
	}

	promptAttemptsSpecificationLNB := schema.SetNestedBlock{
		Validators: []validator.Set{
			setvalidator.SizeAtMost(6),
		},
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
					Optional:   true,
					CustomType: fwtypes.StringEnumType[awstypes.MessageSelectionStrategy](),
				},
			},
			Blocks: map[string]schema.Block{
				"message_group":                 messageGroupLNB,
				"prompt_attempts_specification": promptAttemptsSpecificationLNB,
			},
		},
	}

	failureSuccessTimeoutNBO := schema.NestedBlockObject{
		Blocks: map[string]schema.Block{
			"failure_conditional": conditionalSpecificationLNB,
			"failure_next_step":   nextStepLNB,
			"failure_response":    responseSpecificationLNB,
			"success_conditional": conditionalSpecificationLNB,
			"success_next_step":   nextStepLNB,
			"success_response":    responseSpecificationLNB,
			"timeout_conditional": conditionalSpecificationLNB,
			"timeout_next_step":   nextStepLNB,
			"timeout_response":    responseSpecificationLNB,
		},
	}

	postCodeHookSpecificationLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
		CustomType:   fwtypes.NewListNestedObjectTypeOf[FailureSuccessTimeout](ctx),
		NestedObject: failureSuccessTimeoutNBO,
	}

	codeHookLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[DialogCodeHookInvocationSetting](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"active": schema.BoolAttribute{
					Required: true,
				},
				"enable_code_hook_invocation": schema.BoolAttribute{
					Required: true,
				},
				"invocation_label": schema.StringAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"post_code_hook_specification": postCodeHookSpecificationLNB,
			},
		},
	}

	elicitationCodeHookInvocationSettingLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[ElicitationCodeHookInvocationSetting](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enable_code_hook_invocation": schema.BoolAttribute{
					Optional: true,
				},
				"invocation_label": schema.StringAttribute{
					Optional: true,
				},
			},
		},
	}

	confirmationSettingLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[IntentConfirmationSetting](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"active": schema.BoolAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"code_hook":                codeHookLNB,
				"confirmation_conditional": conditionalSpecificationLNB,
				"confirmation_next_step":   nextStepLNB,
				"confirmation_response":    responseSpecificationLNB,
				"declination_conditional":  conditionalSpecificationLNB,
				"declination_next_step":    nextStepLNB,
				"declination_response":     responseSpecificationLNB,
				"elicitation_code_hook":    elicitationCodeHookInvocationSettingLNB,
				"failure_conditional":      conditionalSpecificationLNB,
				"failure_next_step":        nextStepLNB,
				"failure_response":         responseSpecificationLNB,
				"prompt_specification":     promptSpecificationLNB,
			},
		},
	}

	initialResponseSettingLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[InitialResponseSetting](ctx),
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"code_hook":        codeHookLNB,
				"conditional":      conditionalSpecificationLNB,
				"initial_response": responseSpecificationLNB,
				"next_step":        nextStepLNB,
			},
		},
	}

	updateResponseLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[FulfillmentUpdateResponseSpecification](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"allow_interrupt": schema.BoolAttribute{
					Optional: true,
				},
				"frequency_in_seconds": schema.Int64Attribute{
					Required: true,
				},
			},
			Blocks: map[string]schema.Block{
				"message_group": messageGroupLNB,
			},
		},
	}

	startResponseLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[FulfillmentStartResponseSpecification](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"allow_interrupt": schema.BoolAttribute{
					Optional: true,
				},
				"delay_in_seconds": schema.Int64Attribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"message_group": messageGroupLNB,
			},
		},
	}

	fulfillmentUpdatesSpecificationLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[FulfillmentUpdatesSpecification](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"active": schema.BoolAttribute{
					Required: true,
				},
				"timeout_in_seconds": schema.Int64Attribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"start_response":  startResponseLNB,
				"update_response": updateResponseLNB,
			},
		},
	}

	fulfillmentCodeHookLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[FulfillmentCodeHookSettings](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"active": schema.BoolAttribute{
					Optional: true,
				},
				names.AttrEnabled: schema.BoolAttribute{
					Required: true,
				},
			},
			Blocks: map[string]schema.Block{
				"fulfillment_updates_specification": fulfillmentUpdatesSpecificationLNB,
				"post_fulfillment_status_specification": schema.ListNestedBlock{
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					CustomType:   fwtypes.NewListNestedObjectTypeOf[FailureSuccessTimeout](ctx),
					NestedObject: failureSuccessTimeoutNBO,
				},
			},
		},
	}

	dialogCodeHookLNB := schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[DialogCodeHookSettings](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrEnabled: schema.BoolAttribute{
					Required: true,
				},
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
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creation_date_time": schema.StringAttribute{
				Computed:   true,
				CustomType: timetypes.RFC3339Type{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"intent_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated_date_time": schema.StringAttribute{
				Computed:   true,
				CustomType: timetypes.RFC3339Type{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"locale_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"parent_intent_signature": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"dialog_code_hook":         dialogCodeHookLNB,
			"fulfillment_code_hook":    fulfillmentCodeHookLNB,
			"initial_response_setting": initialResponseSettingLNB,
			"input_context":            inputContextLNB,
			"closing_setting":          closingSettingLNB,
			"confirmation_setting":     confirmationSettingLNB,
			"kendra_configuration":     kendraConfigurationLNB,
			"output_context":           outputContextLNB,
			"sample_utterance":         sampleUtteranceLNB,
			"slot_priority":            slotPriorityLNB,
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceIntent) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var data ResourceIntentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lexmodelsv2.CreateIntentInput{}
	resp.Diagnostics.Append(flex.Expand(context.WithValue(ctx, flex.ResourcePrefix, ResNameIntent), &data, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateIntent(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameIntent, data.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionCreating, ResNameIntent, data.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	data.IntentID = flex.StringToFramework(ctx, out.IntentId)
	data.setID()

	intent, err := waitIntentNormal(ctx, conn, data.IntentID.ValueString(), data.BotID.ValueString(), data.BotVersion.ValueString(), data.LocaleID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForCreation, ResNameIntent, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	// get some data from the intent
	var dataAfter ResourceIntentData
	resp.Diagnostics.Append(flex.Flatten(context.WithValue(ctx, flex.ResourcePrefix, ResNameIntent), intent, &dataAfter)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// unknowns must be set to satisfy apply
	data.CreationDateTime = dataAfter.CreationDateTime
	data.LastUpdatedDateTime = dataAfter.LastUpdatedDateTime

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceIntent) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var data ResourceIntentData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	out, err := findIntentByIDs(ctx, conn, data.IntentID.ValueString(), data.BotID.ValueString(), data.BotVersion.ValueString(), data.LocaleID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionSetting, ResNameIntent, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(context.WithValue(ctx, flex.ResourcePrefix, ResNameIntent), out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceIntent) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var old, new ResourceIntentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	change := false
	if !new.ClosingSetting.Equal(old.ClosingSetting) {
		change = true
	}
	if !new.ConfirmationSetting.Equal(old.ConfirmationSetting) {
		change = true
	}
	if !new.Description.Equal(old.Description) {
		change = true
	}
	if !new.DialogCodeHook.Equal(old.DialogCodeHook) {
		change = true
	}
	if !new.FulfillmentCodeHook.Equal(old.FulfillmentCodeHook) {
		change = true
	}
	if !new.InitialResponseSetting.Equal(old.InitialResponseSetting) {
		change = true
	}
	if !new.InputContext.Equal(old.InputContext) {
		change = true
	}
	if !new.KendraConfiguration.Equal(old.KendraConfiguration) {
		change = true
	}
	if !new.Name.Equal(old.Name) {
		change = true
	}
	if !new.OutputContext.Equal(old.OutputContext) {
		change = true
	}
	if !new.ParentIntentSignature.Equal(old.ParentIntentSignature) {
		change = true
	}
	if !new.SampleUtterance.Equal(old.SampleUtterance) {
		change = true
	}
	if !new.SlotPriority.Equal(old.SlotPriority) {
		change = true
	}

	if !change {
		return
	}

	input := &lexmodelsv2.UpdateIntentInput{}
	resp.Diagnostics.Append(flex.Expand(context.WithValue(ctx, flex.ResourcePrefix, ResNameIntent), &new, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateIntent(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionUpdating, ResNameIntent, new.ID.String(), err),
			err.Error(),
		)
		return
	}

	_, err = waitIntentNormal(ctx, conn, new.IntentID.ValueString(), new.BotID.ValueString(), new.BotVersion.ValueString(), new.LocaleID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForUpdate, ResNameIntent, new.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceIntent) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LexV2ModelsClient(ctx)

	var state ResourceIntentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lexmodelsv2.DeleteIntentInput{
		IntentId:   aws.String(state.IntentID.ValueString()),
		BotId:      aws.String(state.BotID.ValueString()),
		BotVersion: aws.String(state.BotVersion.ValueString()),
		LocaleId:   aws.String(state.LocaleID.ValueString()),
	}

	_, err := conn.DeleteIntent(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException // lexv2models does not seem to use this approach like other services
		if errors.As(err, &nfe) {
			return
		}

		var pfe *awstypes.PreconditionFailedException // PreconditionFailedException: Failed to retrieve resource since it does not exist
		if errors.As(err, &pfe) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionDeleting, ResNameIntent, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	_, err = waitIntentDeleted(ctx, conn, state.IntentID.ValueString(), state.BotID.ValueString(), state.BotVersion.ValueString(), state.LocaleID.ValueString(), r.DeleteTimeout(ctx, state.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForDeletion, ResNameIntent, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (model *ResourceIntentData) InitFromID() error {
	parts := strings.Split(model.ID.ValueString(), ":")
	if len(parts) != 4 {
		return fmt.Errorf("Unexpected format for import key (%s), use: IntentID:BotID:BotVersion:LocaleID", model.ID)
	}
	model.IntentID = types.StringValue(parts[0])
	model.BotID = types.StringValue(parts[1])
	model.BotVersion = types.StringValue(parts[2])
	model.LocaleID = types.StringValue(parts[3])

	return nil
}

func (model *ResourceIntentData) setID() {
	model.ID = types.StringValue(strings.Join([]string{
		model.IntentID.ValueString(),
		model.BotID.ValueString(),
		model.BotVersion.ValueString(),
		model.LocaleID.ValueString(),
	}, ":"))
}

const (
	statusNormal = "Normal"
)

func waitIntentNormal(ctx context.Context, conn *lexmodelsv2.Client, intentID, botID, botVersion, localeID string, timeout time.Duration) (*lexmodelsv2.DescribeIntentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusIntent(ctx, conn, intentID, botID, botVersion, localeID),
		Timeout:                   timeout,
		MinTimeout:                5 * time.Second,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexmodelsv2.DescribeIntentOutput); ok {
		return out, err
	}

	return nil, err
}

func waitIntentDeleted(ctx context.Context, conn *lexmodelsv2.Client, intentID, botID, botVersion, localeID string, timeout time.Duration) (*lexmodelsv2.DescribeIntentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{statusNormal},
		Target:     []string{},
		Refresh:    statusIntent(ctx, conn, intentID, botID, botVersion, localeID),
		Timeout:    timeout,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lexmodelsv2.DescribeIntentOutput); ok {
		return out, err
	}

	return nil, err
}

func statusIntent(ctx context.Context, conn *lexmodelsv2.Client, intentID, botID, botVersion, localeID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findIntentByIDs(ctx, conn, intentID, botID, botVersion, localeID)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, statusNormal, nil
	}
}

func findIntentByIDs(ctx context.Context, conn *lexmodelsv2.Client, intentID, botID, botVersion, localeID string) (*lexmodelsv2.DescribeIntentOutput, error) {
	in := &lexmodelsv2.DescribeIntentInput{
		BotId:      aws.String(botID),
		BotVersion: aws.String(botVersion),
		IntentId:   aws.String(intentID),
		LocaleId:   aws.String(localeID),
	}

	out, err := conn.DescribeIntent(ctx, in)
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

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type PromptAttemptsType string

// Enum values for PromptAttemptsType
const (
	PromptAttemptsTypeInitial PromptAttemptsType = "Initial"
	PromptAttemptsTypeRetry1  PromptAttemptsType = "Retry1"
	PromptAttemptsTypeRetry2  PromptAttemptsType = "Retry2"
	PromptAttemptsTypeRetry3  PromptAttemptsType = "Retry3"
	PromptAttemptsTypeRetry4  PromptAttemptsType = "Retry4"
	PromptAttemptsTypeRetry5  PromptAttemptsType = "Retry5"
)

func (PromptAttemptsType) Values() []PromptAttemptsType {
	return []PromptAttemptsType{
		"Initial",
		"Retry1",
		"Retry2",
		"Retry3",
		"Retry4",
		"Retry5",
	}
}

type SlotPriority struct {
	Priority types.Int64  `tfsdk:"priority"`
	SlotID   types.String `tfsdk:"slot_id"`
}

type SampleUtterance struct {
	Utterance types.String `tfsdk:"utterance"`
}

type OutputContext struct {
	Name                types.String `tfsdk:"name"`
	TimeToLiveInSeconds types.Int64  `tfsdk:"time_to_live_in_seconds"`
	TurnsToLive         types.Int64  `tfsdk:"turns_to_live"`
}

type KendraConfiguration struct {
	KendraIndex              types.String `tfsdk:"kendra_index"`
	QueryFilterString        types.String `tfsdk:"query_filter_string"`
	QueryFilterStringEnabled types.Bool   `tfsdk:"query_filter_string_enabled"`
}

type CustomPayload struct {
	Value types.String `tfsdk:"value"`
}

type Button struct {
	Text  types.String `tfsdk:"text"`
	Value types.String `tfsdk:"value"`
}

type ImageResponseCard struct {
	Button   fwtypes.ListNestedObjectValueOf[Button] `tfsdk:"button"`
	ImageURL types.String                            `tfsdk:"image_url"`
	Subtitle types.String                            `tfsdk:"subtitle"`
	Title    types.String                            `tfsdk:"title"`
}

type PlainTextMessage struct {
	Value types.String `tfsdk:"value"`
}

type SSMLMessage struct {
	Value types.String `tfsdk:"value"`
}

type Message struct {
	CustomPayload     fwtypes.ListNestedObjectValueOf[CustomPayload]     `tfsdk:"custom_payload"`
	ImageResponseCard fwtypes.ListNestedObjectValueOf[ImageResponseCard] `tfsdk:"image_response_card"`
	PlainTextMessage  fwtypes.ListNestedObjectValueOf[PlainTextMessage]  `tfsdk:"plain_text_message"`
	SSMLMessage       fwtypes.ListNestedObjectValueOf[SSMLMessage]       `tfsdk:"ssml_message"`
}

type MessageGroup struct {
	Message   fwtypes.ListNestedObjectValueOf[Message] `tfsdk:"message"`
	Variation fwtypes.ListNestedObjectValueOf[Message] `tfsdk:"variation"`
}

type SlotValue struct {
	InterpretedValue types.String `tfsdk:"interpreted_value"`
}

type SlotValueOverride struct {
	MapBlockKey types.String                               `tfsdk:"map_block_key"`
	Shape       fwtypes.StringEnum[awstypes.SlotShape]     `tfsdk:"shape"`
	Value       fwtypes.ListNestedObjectValueOf[SlotValue] `tfsdk:"value"`
	//Values fwtypes.ListNestedObjectValueOf[SlotValueOverride] `tfsdk:"values"` // recursive type, future support, needs additional development
}

type DialogAction struct {
	Type                fwtypes.StringEnum[awstypes.DialogActionType] `tfsdk:"type"`
	SlotToElicit        types.String                                  `tfsdk:"slot_to_elicit"`
	SuppressNextMessage types.Bool                                    `tfsdk:"suppress_next_message"`
}

type IntentOverride struct {
	Name types.String                                      `tfsdk:"name"`
	Slot fwtypes.SetNestedObjectValueOf[SlotValueOverride] `tfsdk:"slot"`
}

type DialogState struct {
	DialogAction      fwtypes.ListNestedObjectValueOf[DialogAction]   `tfsdk:"dialog_action"`
	Intent            fwtypes.ListNestedObjectValueOf[IntentOverride] `tfsdk:"intent"`
	SessionAttributes fwtypes.MapValueOf[basetypes.StringValue]       `tfsdk:"session_attributes"`
}

type ResponseSpecification struct {
	AllowInterrupt types.Bool                                    `tfsdk:"allow_interrupt"`
	MessageGroup   fwtypes.ListNestedObjectValueOf[MessageGroup] `tfsdk:"message_group"`
}

type Condition struct {
	ExpressionString types.String `tfsdk:"expression_string"`
}

type ConditionalBranch struct {
	Condition fwtypes.ListNestedObjectValueOf[Condition]             `tfsdk:"condition"`
	Name      types.String                                           `tfsdk:"name"`
	NextStep  fwtypes.ListNestedObjectValueOf[DialogState]           `tfsdk:"next_step"`
	Response  fwtypes.ListNestedObjectValueOf[ResponseSpecification] `tfsdk:"response"`
}

type DefaultConditionalBranch struct {
	NextStep fwtypes.ListNestedObjectValueOf[DialogState]           `tfsdk:"next_step"`
	Response fwtypes.ListNestedObjectValueOf[ResponseSpecification] `tfsdk:"response"`
}

type ConditionalSpecification struct {
	Active            types.Bool                                                `tfsdk:"active"`
	ConditionalBranch fwtypes.ListNestedObjectValueOf[ConditionalBranch]        `tfsdk:"conditional_branch"`
	DefaultBranch     fwtypes.ListNestedObjectValueOf[DefaultConditionalBranch] `tfsdk:"default_branch"`
}

type IntentClosingSetting struct {
	Active          types.Bool                                                `tfsdk:"active"`
	ClosingResponse fwtypes.ListNestedObjectValueOf[ResponseSpecification]    `tfsdk:"closing_response"`
	Conditional     fwtypes.ListNestedObjectValueOf[ConditionalSpecification] `tfsdk:"conditional"`
	NextStep        fwtypes.ListNestedObjectValueOf[DialogState]              `tfsdk:"next_step"`
}

type InputContext struct {
	Name types.String `tfsdk:"name"`
}

type AllowedInputTypes struct {
	AllowAudioInput types.Bool `tfsdk:"allow_audio_input"`
	AllowDTMFInput  types.Bool `tfsdk:"allow_dtmf_input"`
}

type AudioSpecification struct {
	EndTimeoutMs types.Int64 `tfsdk:"end_timeout_ms"`
	MaxLengthMs  types.Int64 `tfsdk:"max_length_ms"`
}

type DTMFSpecification struct {
	DeletionCharacter types.String `tfsdk:"deletion_character"`
	EndCharacter      types.String `tfsdk:"end_character"`
	EndTimeoutMs      types.Int64  `tfsdk:"end_timeout_ms"`
	MaxLength         types.Int64  `tfsdk:"max_length"`
}

type AudioAndDTMFInputSpecification struct {
	StartTimeoutMs     types.Int64                                         `tfsdk:"start_timeout_ms"`
	AudioSpecification fwtypes.ListNestedObjectValueOf[AudioSpecification] `tfsdk:"audio_specification"`
	DTMFSpecification  fwtypes.ListNestedObjectValueOf[DTMFSpecification]  `tfsdk:"dtmf_specification"`
}

type TextInputSpecification struct {
	StartTimeoutMs types.Int64 `tfsdk:"start_timeout_ms"`
}

type PromptAttemptsSpecification struct {
	AllowedInputTypes              fwtypes.ListNestedObjectValueOf[AllowedInputTypes]              `tfsdk:"allowed_input_types"`
	AllowInterrupt                 types.Bool                                                      `tfsdk:"allow_interrupt"`
	AudioAndDTMFInputSpecification fwtypes.ListNestedObjectValueOf[AudioAndDTMFInputSpecification] `tfsdk:"audio_and_dtmf_input_specification"`
	MapBlockKey                    fwtypes.StringEnum[PromptAttemptsType]                          `tfsdk:"map_block_key"`
	TextInputSpecification         fwtypes.ListNestedObjectValueOf[TextInputSpecification]         `tfsdk:"text_input_specification"`
}

type PromptSpecification struct {
	AllowInterrupt              types.Bool                                                  `tfsdk:"allow_interrupt"`
	MaxRetries                  types.Int64                                                 `tfsdk:"max_retries"`
	MessageGroup                fwtypes.ListNestedObjectValueOf[MessageGroup]               `tfsdk:"message_group"`
	MessageSelectionStrategy    fwtypes.StringEnum[awstypes.MessageSelectionStrategy]       `tfsdk:"message_selection_strategy"`
	PromptAttemptsSpecification fwtypes.SetNestedObjectValueOf[PromptAttemptsSpecification] `tfsdk:"prompt_attempts_specification"`
}

type FailureSuccessTimeout struct {
	FailureConditional fwtypes.ListNestedObjectValueOf[ConditionalSpecification] `tfsdk:"failure_conditional"`
	FailureNextStep    fwtypes.ListNestedObjectValueOf[DialogState]              `tfsdk:"failure_next_step"`
	FailureResponse    fwtypes.ListNestedObjectValueOf[ResponseSpecification]    `tfsdk:"failure_response"`
	SuccessConditional fwtypes.ListNestedObjectValueOf[ConditionalSpecification] `tfsdk:"success_conditional"`
	SuccessNextStep    fwtypes.ListNestedObjectValueOf[DialogState]              `tfsdk:"success_next_step"`
	SuccessResponse    fwtypes.ListNestedObjectValueOf[ResponseSpecification]    `tfsdk:"success_response"`
	TimeoutConditional fwtypes.ListNestedObjectValueOf[ConditionalSpecification] `tfsdk:"timeout_conditional"`
	TimeoutNextStep    fwtypes.ListNestedObjectValueOf[DialogState]              `tfsdk:"timeout_next_step"`
	TimeoutResponse    fwtypes.ListNestedObjectValueOf[ResponseSpecification]    `tfsdk:"timeout_response"`
}

type DialogCodeHookInvocationSetting struct {
	Active                    types.Bool                                             `tfsdk:"active"`
	EnableCodeHookInvocation  types.Bool                                             `tfsdk:"enable_code_hook_invocation"`
	InvocationLabel           types.String                                           `tfsdk:"invocation_label"`
	PostCodeHookSpecification fwtypes.ListNestedObjectValueOf[FailureSuccessTimeout] `tfsdk:"post_code_hook_specification"`
}

type ElicitationCodeHookInvocationSetting struct {
	EnableCodeHookInvocation types.Bool   `tfsdk:"enable_code_hook_invocation"`
	InvocationLabel          types.String `tfsdk:"invocation_label"`
}

type IntentConfirmationSetting struct {
	Active                  types.Bool                                                            `tfsdk:"active"`
	CodeHook                fwtypes.ListNestedObjectValueOf[DialogCodeHookInvocationSetting]      `tfsdk:"code_hook"`
	ConfirmationConditional fwtypes.ListNestedObjectValueOf[ConditionalSpecification]             `tfsdk:"confirmation_conditional"`
	ConfirmationNextStep    fwtypes.ListNestedObjectValueOf[DialogState]                          `tfsdk:"confirmation_next_step"`
	ConfirmationResponse    fwtypes.ListNestedObjectValueOf[ResponseSpecification]                `tfsdk:"confirmation_response"`
	DeclinationConditional  fwtypes.ListNestedObjectValueOf[ConditionalSpecification]             `tfsdk:"declination_conditional"`
	DeclinationNextStep     fwtypes.ListNestedObjectValueOf[DialogState]                          `tfsdk:"declination_next_step"`
	DeclinationResponse     fwtypes.ListNestedObjectValueOf[ResponseSpecification]                `tfsdk:"declination_response"`
	ElicitationCodeHook     fwtypes.ListNestedObjectValueOf[ElicitationCodeHookInvocationSetting] `tfsdk:"elicitation_code_hook"`
	FailureConditional      fwtypes.ListNestedObjectValueOf[ConditionalSpecification]             `tfsdk:"failure_conditional"`
	FailureNextStep         fwtypes.ListNestedObjectValueOf[DialogState]                          `tfsdk:"failure_next_step"`
	FailureResponse         fwtypes.ListNestedObjectValueOf[ResponseSpecification]                `tfsdk:"failure_response"`
	PromptSpecification     fwtypes.ListNestedObjectValueOf[PromptSpecification]                  `tfsdk:"prompt_specification"`
}

type InitialResponseSetting struct {
	CodeHook        fwtypes.ListNestedObjectValueOf[DialogCodeHookInvocationSetting] `tfsdk:"code_hook"`
	Conditional     fwtypes.ListNestedObjectValueOf[ConditionalSpecification]        `tfsdk:"conditional"`
	InitialResponse fwtypes.ListNestedObjectValueOf[ResponseSpecification]           `tfsdk:"initial_response"`
	NextStep        fwtypes.ListNestedObjectValueOf[DialogState]                     `tfsdk:"next_step"`
}

type FulfillmentUpdateResponseSpecification struct {
	AllowInterrupt     types.Bool                                    `tfsdk:"allow_interrupt"`
	FrequencyInSeconds types.Int64                                   `tfsdk:"frequency_in_seconds"`
	MessageGroup       fwtypes.ListNestedObjectValueOf[MessageGroup] `tfsdk:"message_group"`
}

type FulfillmentStartResponseSpecification struct {
	AllowInterrupt types.Bool                                    `tfsdk:"allow_interrupt"`
	DelayInSeconds types.Int64                                   `tfsdk:"delay_in_seconds"`
	MessageGroup   fwtypes.ListNestedObjectValueOf[MessageGroup] `tfsdk:"message_group"`
}

type FulfillmentUpdatesSpecification struct {
	Active           types.Bool                                                              `tfsdk:"active"`
	StartResponse    fwtypes.ListNestedObjectValueOf[FulfillmentStartResponseSpecification]  `tfsdk:"start_response"`
	TimeoutInSeconds types.Int64                                                             `tfsdk:"timeout_in_seconds"`
	UpdateResponse   fwtypes.ListNestedObjectValueOf[FulfillmentUpdateResponseSpecification] `tfsdk:"update_response"`
}

type FulfillmentCodeHookSettings struct {
	Active                             types.Bool                                                       `tfsdk:"active"`
	Enabled                            types.Bool                                                       `tfsdk:"enabled"`
	FulfillmentUpdatesSpecification    fwtypes.ListNestedObjectValueOf[FulfillmentUpdatesSpecification] `tfsdk:"fulfillment_updates_specification"`
	PostFulfillmentStatusSpecification fwtypes.ListNestedObjectValueOf[FailureSuccessTimeout]           `tfsdk:"post_fulfillment_status_specification"`
}

type DialogCodeHookSettings struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type ResourceIntentData struct {
	BotID                  types.String                                                 `tfsdk:"bot_id"`
	BotVersion             types.String                                                 `tfsdk:"bot_version"`
	ClosingSetting         fwtypes.ListNestedObjectValueOf[IntentClosingSetting]        `tfsdk:"closing_setting"`
	ConfirmationSetting    fwtypes.ListNestedObjectValueOf[IntentConfirmationSetting]   `tfsdk:"confirmation_setting"`
	CreationDateTime       timetypes.RFC3339                                            `tfsdk:"creation_date_time"`
	Description            types.String                                                 `tfsdk:"description"`
	DialogCodeHook         fwtypes.ListNestedObjectValueOf[DialogCodeHookSettings]      `tfsdk:"dialog_code_hook"`
	FulfillmentCodeHook    fwtypes.ListNestedObjectValueOf[FulfillmentCodeHookSettings] `tfsdk:"fulfillment_code_hook"`
	ID                     types.String                                                 `tfsdk:"id"`
	IntentID               types.String                                                 `tfsdk:"intent_id"`
	InitialResponseSetting fwtypes.ListNestedObjectValueOf[InitialResponseSetting]      `tfsdk:"initial_response_setting"`
	InputContext           fwtypes.ListNestedObjectValueOf[InputContext]                `tfsdk:"input_context"`
	KendraConfiguration    fwtypes.ListNestedObjectValueOf[KendraConfiguration]         `tfsdk:"kendra_configuration"`
	LastUpdatedDateTime    timetypes.RFC3339                                            `tfsdk:"last_updated_date_time"`
	LocaleID               types.String                                                 `tfsdk:"locale_id"`
	Name                   types.String                                                 `tfsdk:"name"`
	OutputContext          fwtypes.ListNestedObjectValueOf[OutputContext]               `tfsdk:"output_context"`
	ParentIntentSignature  types.String                                                 `tfsdk:"parent_intent_signature"`
	SampleUtterance        fwtypes.ListNestedObjectValueOf[SampleUtterance]             `tfsdk:"sample_utterance"`
	SlotPriority           fwtypes.ListNestedObjectValueOf[SlotPriority]                `tfsdk:"slot_priority"`
	Timeouts               timeouts.Value                                               `tfsdk:"timeouts"`
}
