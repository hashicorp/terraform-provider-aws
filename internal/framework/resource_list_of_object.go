// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// NewResourceComputedListOfObjectSchema returns a new schema.ListAttribute for objects of the specified type.
// The list is Computed-only.
func ResourceComputedListOfObjectsAttribute[T any](ctx context.Context, planModifiers ...planmodifier.List) schema.ListAttribute {
	return schema.ListAttribute{
		CustomType:    fwtypes.NewListNestedObjectTypeOf[T](ctx),
		Computed:      true,
		PlanModifiers: planModifiers,
		ElementType: types.ObjectType{
			AttrTypes: fwtypes.AttributeTypesMust[T](ctx),
		},
	}
}

// ResourceOptionalComputedListOfObjectsAttribute returns a new schema.ListAttribute for objects of the specified type.
// The list is Optional+Computed.
func ResourceOptionalComputedListOfObjectsAttribute[T any](ctx context.Context, sizeAtMost int, planModifiers ...planmodifier.List) schema.ListAttribute {
	return schema.ListAttribute{
		CustomType:    fwtypes.NewListNestedObjectTypeOf[T](ctx),
		Optional:      true,
		Computed:      true,
		PlanModifiers: planModifiers,
		Validators: []validator.List{
			listvalidator.SizeAtMost(sizeAtMost),
		},
		ElementType: types.ObjectType{
			AttrTypes: fwtypes.AttributeTypesMust[T](ctx),
		},
	}
}
