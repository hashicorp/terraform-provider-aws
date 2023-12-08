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

func DefaultValueSpecificationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[DefaultValueSpecificationData](ctx),
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"default_value_list": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[DefaultValueData](ctx),
					Validators: []validator.List{
						listvalidator.IsRequired(),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"default_value": schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
		},
	}
}

type DefaultValueSpecificationData struct {
	DefaultValueList fwtypes.ListNestedObjectValueOf[DefaultValueData] `tfsdk:"default_value_list"`
}

type DefaultValueData struct {
	DefaultValue types.String `tfsdk:"default_value"`
}
