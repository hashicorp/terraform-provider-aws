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
	return schema.ListNestedBlock{
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
				"variations": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[Message](ctx),
					NestedObject: messageNBO,
				},
			},
		},
	}
}

type CustomPayload struct {
	Value types.String `tfsdk:"value"`
}

type ImageResponseCard struct {
	Title    types.String                            `tfsdk:"title"`
	Button   fwtypes.ListNestedObjectValueOf[Button] `tfsdk:"buttons"`
	ImageURL types.String                            `tfsdk:"image_url"`
	Subtitle types.String                            `tfsdk:"subtitle"`
}

type Button struct {
	Text  types.String `tfsdk:"text"`
	Value types.String `tfsdk:"value"`
}

type PlainTextMessage struct {
	Value types.String `tfsdk:"value"`
}

type SSMLMessage struct {
	Value types.String `tfsdk:"value"`
}
type MessageGroup struct {
	Message    fwtypes.ListNestedObjectValueOf[Message] `tfsdk:"message"`
	Variations fwtypes.ListNestedObjectValueOf[Message] `tfsdk:"variations"`
}

type Message struct {
	CustomPayload     fwtypes.ListNestedObjectValueOf[CustomPayload]     `tfsdk:"custom_payload"`
	ImageResponseCard fwtypes.ListNestedObjectValueOf[ImageResponseCard] `tfsdk:"image_response_card"`
	PlainTextMessage  fwtypes.ListNestedObjectValueOf[PlainTextMessage]  `tfsdk:"plain_text_message"`
	SSMLMessage       fwtypes.ListNestedObjectValueOf[SSMLMessage]       `tfsdk:"ssml_message"`
}
