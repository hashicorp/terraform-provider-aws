// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// NewDataSourceComputedListOfObjectSchema returns a new schema.ListAttribute for objects of the specified type.
// The list is Computed-only.
func NewDataSourceComputedListOfObjectSchema[T any](ctx context.Context) schema.ListAttribute {
	return schema.ListAttribute{
		CustomType: fwtypes.NewListNestedObjectTypeOf[T](ctx),
		Computed:   true,
		ElementType: types.ObjectType{
			AttrTypes: fwtypes.AttributeTypesMust[T](ctx),
		},
	}
}
