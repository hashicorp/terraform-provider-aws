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

func AudioSpecificationnBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
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
}

type AudioSpecification struct {
	EndTimeoutMs types.Int64 `tfsdk:"end_timeout_ms"`
	MaxLengthMs  types.Int64 `tfsdk:"max_length_ms"`
}
