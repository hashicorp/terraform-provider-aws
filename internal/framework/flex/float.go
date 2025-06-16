// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

// Float64FromFramework converts a Framework Int64 value to an int64 pointer.
// A null Int64 is converted to a nil int64 pointer.
func Float64FromFramework(ctx context.Context, v basetypes.Float64Valuable) *float64 {
	if v.IsUnknown() {
		return nil
	}
	val := fwdiag.Must(v.ToFloat64Value(ctx))
	return val.ValueFloat64Pointer()
}

// Float32ToFrameworkFloat64Legacy converts a float32 pointer to a Framework Float64 value.
// A nil float32 pointer is converted to a zero float64.
func Float32ToFrameworkFloat64Legacy(_ context.Context, v *float32) types.Float64 {
	return types.Float64Value(float64(aws.ToFloat32(v)))
}
