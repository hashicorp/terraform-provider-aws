// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// NullValueOf returns a null attr.Value for the specified `v`.
func NullValueOf(ctx context.Context, v any) (attr.Value, diag.Diagnostics) {
	var attrType attr.Type
	var tfType tftypes.Type

	switch v := v.(type) {
	case basetypes.BoolValuable:
		attrType = v.Type(ctx)
		toType := attrType.(basetypes.BoolTypable)
		return toType.ValueFromBool(ctx, types.BoolNull())

	case basetypes.Float32Valuable:
		attrType = v.Type(ctx)
		toType := attrType.(basetypes.Float32Typable)
		return toType.ValueFromFloat32(ctx, types.Float32Null())

	case basetypes.Float64Valuable:
		attrType = v.Type(ctx)
		toType := attrType.(basetypes.Float64Typable)
		return toType.ValueFromFloat64(ctx, types.Float64Null())

	case basetypes.Int32Valuable:
		attrType = v.Type(ctx)
		toType := attrType.(basetypes.Int32Typable)
		return toType.ValueFromInt32(ctx, types.Int32Null())

	case basetypes.Int64Valuable:
		attrType = v.Type(ctx)
		toType := attrType.(basetypes.Int64Typable)
		return toType.ValueFromInt64(ctx, types.Int64Null())

	case basetypes.StringValuable:
		attrType = v.Type(ctx)
		toType := attrType.(basetypes.StringTypable)
		return toType.ValueFromString(ctx, types.StringNull())

	case basetypes.ListValuable:
		attrType = v.Type(ctx)
		if v, ok := attrType.(attr.TypeWithElementType); ok {
			toType := attrType.(basetypes.ListTypable)
			return toType.ValueFromList(ctx, types.ListNull(v.ElementType()))
		} else {
			tfType = tftypes.List{}
		}

	case basetypes.SetValuable:
		attrType = v.Type(ctx)
		if v, ok := attrType.(attr.TypeWithElementType); ok {
			toType := attrType.(basetypes.SetTypable)
			return toType.ValueFromSet(ctx, types.SetNull(v.ElementType()))
		} else {
			tfType = tftypes.Set{}
		}

	case basetypes.MapValuable:
		attrType = v.Type(ctx)
		if v, ok := attrType.(attr.TypeWithElementType); ok {
			toType := attrType.(basetypes.MapTypable)
			return toType.ValueFromMap(ctx, types.MapNull(v.ElementType()))
		} else {
			tfType = tftypes.Map{}
		}

	case basetypes.ObjectValuable:
		attrType = v.Type(ctx)
		if v, ok := attrType.(attr.TypeWithAttributeTypes); ok {
			toType := attrType.(basetypes.ObjectTypable)
			return toType.ValueFromObject(ctx, types.ObjectNull(v.AttributeTypes()))
		} else {
			tfType = tftypes.Object{}
		}

	default:
		return nil, nil
	}

	result, err := attrType.ValueFromTerraform(ctx, tftypes.NewValue(tfType, nil))
	return result, diag.Diagnostics{diag.NewErrorDiagnostic(
		"Incompatible Types",
		err.Error(),
	)}
}
