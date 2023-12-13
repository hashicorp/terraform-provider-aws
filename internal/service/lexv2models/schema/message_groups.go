// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func MessageGroupsBlock(ctx context.Context) schema.ListNestedBlock {
	messageNBO := schema.NestedBlockObject{
		Blocks: map[string]schema.Block{
			"custom_playload": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[CustomPayloadData](ctx),
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
				CustomType: fwtypes.NewListNestedObjectTypeOf[ImageResponseCardData](ctx),
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[ButtonData](ctx),
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
				CustomType: fwtypes.NewListNestedObjectTypeOf[PlainTextMessageData](ctx),
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
				CustomType: fwtypes.NewListNestedObjectTypeOf[SSMLMessageData](ctx),
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
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[MessageGroupData](ctx),
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"message": schema.ListNestedBlock{
					Validators: []validator.List{
						listvalidator.SizeBetween(1, 1),
					},
					CustomType:   fwtypes.NewListNestedObjectTypeOf[MessageData](ctx),
					NestedObject: messageNBO,
				},
				"variations": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[MessageData](ctx),
					NestedObject: messageNBO,
				},
			},
		},
	}
}

type CustomPayloadData struct {
	Value types.String `tfsdk:"value"`
}

type ImageResponseCardData struct {
	Title    types.String                                `tfsdk:"title"`
	Button   fwtypes.ListNestedObjectValueOf[ButtonData] `tfsdk:"buttons"`
	ImageURL types.String                                `tfsdk:"image_url"`
	Subtitle types.String                                `tfsdk:"subtitle"`
}

type ButtonData struct {
	Text  types.String `tfsdk:"text"`
	Value types.String `tfsdk:"value"`
}

type PlainTextMessageData struct {
	Value types.String `tfsdk:"value"`
}

type SSMLMessageData struct {
	Value types.String `tfsdk:"value"`
}
type MessageGroupData struct {
	Message    fwtypes.ListNestedObjectValueOf[MessageData] `tfsdk:"message"`
	Variations fwtypes.ListNestedObjectValueOf[MessageData] `tfsdk:"variations"`
}

type MessageData struct {
	CustomPayload     fwtypes.ListNestedObjectValueOf[CustomPayloadData]     `tfsdk:"custom_payload"`
	ImageResponseCard fwtypes.ListNestedObjectValueOf[ImageResponseCardData] `tfsdk:"image_response_card"`
	PlainTextMessage  fwtypes.ListNestedObjectValueOf[PlainTextMessageData]  `tfsdk:"plain_text_message"`
	SSMLMessage       fwtypes.ListNestedObjectValueOf[SSMLMessageData]       `tfsdk:"ssml_message"`
}
