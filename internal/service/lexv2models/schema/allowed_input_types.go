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

func AllowedInputTypesBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
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
}

type AllowedInputTypes struct {
	AllowAudioInput types.Bool `tfsdk:"allow_audio_input"`
	AllowDTMFInput  types.Bool `tfsdk:"allow_dtmf_input"`
}
