// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func PromptSpecificationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeBetween(1, 1),
		},
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
				"message_groups":                MessageGroupsBlock(ctx),
				"prompt_attempts_specification": PromptAttemptsSpecificationBlock(ctx),
			},
		},
	}
}

type PromptSpecificationData struct {
	AllowInterrupt              types.Bool                                               `tfsdk:"allow_interrupt"`
	MaxRetries                  types.Int64                                              `tfsdk:"max_retries"`
	MessageGroup                fwtypes.ListNestedObjectValueOf[MessageGroupData]        `tfsdk:"message_groups"`
	MessageSelectionStrategy    fwtypes.StringEnum[awstypes.MessageSelectionStrategy]    `tfsdk:"message_selection_strategy"`
	PromptAttemptsSpecification fwtypes.ObjectMapValueOf[PromptAttemptSpecificationData] `tfsdk:"prompt_attempts_specification"`
}

type PromptAttemptSpecificationData struct {
	AllowedInputTypes              fwtypes.ListNestedObjectValueOf[AllowedInputTypesData]              `tfsdk:"allowed_input_types"`
	AllowInterrupt                 types.Bool                                                          `tfsdk:"allow_interrupt"`
	AudioAndDTMFInputSpecification fwtypes.ListNestedObjectValueOf[AudioAndDTMFInputSpecificationData] `tfsdk:"audio_and_dtmf_input_specification"`
	TextInputSpecification         fwtypes.ListNestedObjectValueOf[TextInputSpecificationData]         `tfsdk:"text_input_specification"`
}

type DTMFSpecificationData struct {
	EndCharacter      types.String `tfsdk:"end_character"`
	EndTimeoutMs      types.Int64  `tfsdk:"end_timeout_ms"`
	DeletionCharacter types.String `tfsdk:"deletion_character"`
	MaxLength         types.Int64  `tfsdk:"max_length"`
}

type TextInputSpecificationData struct {
	StartTimeoutMs types.Int64 `tfsdk:"start_timeout_ms"`
}

type AllowedInputTypesData struct {
	AllowAudioInput types.Bool `tfsdk:"allow_audio_input"`
	AllowDTMFInput  types.Bool `tfsdk:"allow_dtmf_input"`
}

type AudioAndDTMFInputSpecificationData struct {
	AudioSpecification fwtypes.ListNestedObjectValueOf[AudioSpecificationData] `tfsdk:"audio_specification"`
	StartTimeoutMs     types.Int64                                             `tfsdk:"start_timeout_ms"`
	DTMFSpecification  fwtypes.ListNestedObjectValueOf[DTMFSpecificationData]  `tfsdk:"dtmf_specification"`
}

type AudioSpecificationData struct {
	EndTimeoutMs types.Int64 `tfsdk:"end_timeout_ms"`
	MaxLengthMs  types.Int64 `tfsdk:"max_length_ms"`
}
