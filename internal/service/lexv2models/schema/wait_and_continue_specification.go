// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type WaitAndContinueSpecificationData struct {
	Active               types.Bool                                                             `tfsdk:"active"`
	ContinueResponse     fwtypes.ListNestedObjectValueOf[ResponseSpecificationData]             `tfsdk:"continue_response"`
	StillWaitingResponse fwtypes.ListNestedObjectValueOf[StillWaitingResponseSpecificationData] `tfsdk:"still_waiting_response"`
	WaitingResponse      fwtypes.ListNestedObjectValueOf[ResponseSpecificationData]             `tfsdk:"waiting_response"`
}

func WaitAndContinueSpecificationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[WaitAndContinueSpecificationData](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"active": schema.BoolAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"continue_response":      ResponseSpecificationBlock(ctx),
				"still_waiting_response": StillWaitingResponseSpecificationBlock(ctx),
				"waiting_response":       ResponseSpecificationBlock(ctx),
			},
		},
	}
}

type ResponseSpecificationData struct {
	AllowInterrupt types.Bool                                        `tfsdk:"allow_interrupt"`
	MessageGroups  fwtypes.ListNestedObjectValueOf[MessageGroupData] `tfsdk:"message_groups"`
}

func ResponseSpecificationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ResponseSpecificationData](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"allow_interrupt": schema.BoolAttribute{
					Optional: true,
				},
			},
			Blocks: map[string]schema.Block{
				"message_groups": MessageGroupsBlock(ctx),
			},
		},
	}
}

type StillWaitingResponseSpecificationData struct {
	AllowInterrupt     types.Bool                                        `tfsdk:"allow_interrupt"`
	FrequencyInSeconds types.Int64                                       `tfsdk:"frequency_in_seconds"`
	MessageGroups      fwtypes.ListNestedObjectValueOf[MessageGroupData] `tfsdk:"message_groups"`
	TimeoutInSeconds   types.Int64                                       `tfsdk:"timeout_in_seconds"`
}

func StillWaitingResponseSpecificationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
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
				"message_groups": MessageGroupsBlock(ctx),
			},
		},
	}
}
