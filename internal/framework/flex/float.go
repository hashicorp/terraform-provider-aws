// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Float64ToFramework converts a float64 pointer to a Framework Float64 value.
// A nil float64 pointer is converted to a null Float64.
func Float64ToFramework(ctx context.Context, v *float64) types.Float64 {
	return types.Float64PointerValue(v)
}

// Float64ToFrameworkLegacy converts a float64 pointer to a Framework Float64 value.
// A nil float64 pointer is converted to a zero float64.
func Float64ToFrameworkLegacy(_ context.Context, v *float64) types.Float64 {
	return types.Float64Value(aws.ToFloat64(v))
}

// Float32ToFrameworkFloat64 converts a float32 pointer to a Framework Float64 value.
// A nil float32 pointer is converted to a null Float64.
func Float32ToFrameworkFloat64(_ context.Context, v *float32) types.Float64 {
	if v == nil {
		return types.Float64Null()
	}
	return types.Float64Value(float64(aws.ToFloat32(v)))
}

// Float32ToFrameworkFloat64Legacy converts a float32 pointer to a Framework Float64 value.
// A nil float32 pointer is converted to a zero float64.
func Float32ToFrameworkFloat64Legacy(_ context.Context, v *float32) types.Float64 {
	return types.Float64Value(float64(aws.ToFloat32(v)))
}
