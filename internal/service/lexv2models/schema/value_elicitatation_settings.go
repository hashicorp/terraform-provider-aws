// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func ValueElicitationSettingBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
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
				"default_value_specification": DefaultValueSpecificationBlock(ctx),
				"slot_resolution_setting":     SlotResolutionSettingBlock(ctx),
				"prompt_specification":        PromptSpecificationBlock(ctx),
				"sample_utterance":            SampleUtteranceBlock(ctx),
			},
		},
	}
}

type ValueElicitationSettingData struct {
	SlotConstraint            fwtypes.StringEnum[awstypes.SlotConstraint]                    `tfsdk:"slot_constraint"`
	DefaultValueSpecification fwtypes.ListNestedObjectValueOf[DefaultValueSpecificationData] `tfsdk:"default_value_specification"`
	SlotResolutionSetting     fwtypes.ListNestedObjectValueOf[SlotResolutionSettingData]     `tfsdk:"slot_resolution_setting"`
	PromptSpecification       fwtypes.ListNestedObjectValueOf[PromptSpecificationData]       `tfsdk:"prompt_specification"`
	SampleUtterance           fwtypes.ListNestedObjectValueOf[SampleUtteranceData]           `tfsdk:"sample_utterance"`
}
