// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func AudioAndDTMFInputSpecificationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
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
				"audio_specification": AudioAndDTMFInputSpecificationBlock(ctx),
				"dtmf_specification":  DtmfSpecificationBlock(ctx),
			},
		},
	}
}

type AudioAndDTMFInputSpecification struct {
	StartTimeoutMs     types.Int64                                         `tfsdk:"start_timeout_ms"`
	AudioSpecification fwtypes.ListNestedObjectValueOf[AudioSpecification] `tfsdk:"audio_specification"`
	DTMFSpecification  fwtypes.ListNestedObjectValueOf[DTMFSpecification]  `tfsdk:"dtmf_specification"`
}
