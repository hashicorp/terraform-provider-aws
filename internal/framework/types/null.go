// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
)

// NullValueOf returns a null attr.Value for the specified `v`.
func NullValueOf(ctx context.Context, v any) (attr.Value, error) {
	var attrType attr.Type
	var tfType tftypes.Type

	switch v := v.(type) {
	case basetypes.BoolValuable:
		attrType = v.Type(ctx)
		tfType = tftypes.Bool
	case basetypes.Float64Valuable:
		attrType = v.Type(ctx)
		tfType = tftypes.Number
	case basetypes.Int64Valuable:
		attrType = v.Type(ctx)
		tfType = tftypes.Number
	case basetypes.StringValuable:
		attrType = v.Type(ctx)
		tfType = tftypes.String
	case basetypes.ListValuable:
		attrType = v.Type(ctx)
		if v, ok := attrType.(attr.TypeWithElementType); ok {
			tfType = tftypes.List{ElementType: v.ElementType().TerraformType(ctx)}
		} else {
			tfType = tftypes.List{}
		}
	case basetypes.SetValuable:
		attrType = v.Type(ctx)
		if v, ok := attrType.(attr.TypeWithElementType); ok {
			tfType = tftypes.Set{ElementType: v.ElementType().TerraformType(ctx)}
		} else {
			tfType = tftypes.Set{}
		}
	case basetypes.MapValuable:
		attrType = v.Type(ctx)
		if v, ok := attrType.(attr.TypeWithElementType); ok {
			tfType = tftypes.Map{ElementType: v.ElementType().TerraformType(ctx)}
		} else {
			tfType = tftypes.Map{}
		}
	case basetypes.ObjectValuable:
		attrType = v.Type(ctx)
		if v, ok := attrType.(attr.TypeWithAttributeTypes); ok {
			tfType = tftypes.Object{AttributeTypes: tfmaps.ApplyToAllValues(v.AttributeTypes(), func(attrType attr.Type) tftypes.Type {
				return attrType.TerraformType(ctx)
			})}
		} else {
			tfType = tftypes.Object{}
		}
	default:
		return nil, nil
	}

	return attrType.ValueFromTerraform(ctx, tftypes.NewValue(tfType, nil))
}
