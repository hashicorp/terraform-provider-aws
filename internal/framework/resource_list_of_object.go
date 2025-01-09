// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// NewResourceComputedListOfObjectSchema returns a new schema.ListAttribute for objects of the specified type.
// The list is Computed-only.
func ResourceComputedListOfObjectsAttribute[T any](ctx context.Context) schema.ListAttribute {
	return schema.ListAttribute{
		CustomType: fwtypes.NewListNestedObjectTypeOf[T](ctx),
		Computed:   true,
		ElementType: types.ObjectType{
			AttrTypes: fwtypes.AttributeTypesMust[T](ctx),
		},
	}
}

// ResourceOptionalComputedListOfSingleObjectAttribute returns a new schema.ListAttribute for a single object of the specified type.
// The list is Optional+Computed.
func ResourceOptionalComputedListOfSingleObjectAttribute[T any](ctx context.Context) schema.ListAttribute {
	return schema.ListAttribute{
		CustomType: fwtypes.NewListNestedObjectTypeOf[T](ctx),
		Optional:   true,
		Computed:   true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		ElementType: types.ObjectType{
			AttrTypes: fwtypes.AttributeTypesMust[T](ctx),
		},
	}
}
