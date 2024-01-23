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

func TextInputSpecificationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[TextInputSpecification](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"start_timeout_ms": schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
			},
		},
	}
}

type TextInputSpecification struct {
	StartTimeoutMs types.Int64 `tfsdk:"start_timeout_ms"`
}
