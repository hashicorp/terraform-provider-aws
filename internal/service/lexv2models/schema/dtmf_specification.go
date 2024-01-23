// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func DtmfSpecificationBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		CustomType: fwtypes.NewListNestedObjectTypeOf[DTMFSpecification](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"deletion_character": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.RegexMatches(
							regexache.MustCompile(`^[A-D0-9#*]{1}$`),
							"alphanumeric characters",
						),
					},
				},
				"end_character": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.RegexMatches(
							regexache.MustCompile(`^[A-D0-9#*]{1}$`),
							"alphanumeric characters",
						),
					},
				},
				"end_timeout_ms": schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						int64validator.AtLeast(1),
					},
				},
				"max_length": schema.Int64Attribute{
					Required: true,
					Validators: []validator.Int64{
						int64validator.Between(1, 1024),
					},
				},
			},
		},
	}
}

type DTMFSpecification struct {
	DeletionCharacter types.String `tfsdk:"deletion_character"`
	EndCharacter      types.String `tfsdk:"end_character"`
	EndTimeoutMs      types.Int64  `tfsdk:"end_timeout_ms"`
	MaxLength         types.Int64  `tfsdk:"max_length"`
}
