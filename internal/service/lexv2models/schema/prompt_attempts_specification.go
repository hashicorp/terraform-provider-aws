// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func PromptAttemptsSpecificationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[PromptAttemptsSpecificationData](ctx),
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
				"allowed_input_types":                AllowedInputTypesBlock(ctx),
				"audio_and_dtmf_input_specification": AudioAndDTMFInputSpecificationBlock(ctx),
				"text_input_specification":           TextInputSpecificationBlock(ctx),
			},
		},
	}
}

type PromptAttemptsSpecificationData struct {
	AllowedInputTypes              fwtypes.ListNestedObjectValueOf[AllowedInputTypes]              `tfsdk:"allowed_input_types"`
	AllowInterrupt                 types.Bool                                                      `tfsdk:"allow_interrupt"`
	AudioAndDTMFInputSpecification fwtypes.ListNestedObjectValueOf[AudioAndDTMFInputSpecification] `tfsdk:"audio_and_dtmf_input_specification"`
	MapBlockKey                    fwtypes.StringEnum[PromptAttemptsType]                          `tfsdk:"map_block_key"`
	TextInputSpecification         fwtypes.ListNestedObjectValueOf[TextInputSpecification]         `tfsdk:"text_input_specification"`
}

type PromptAttemptsType string

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
