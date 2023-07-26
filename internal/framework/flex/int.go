// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Int64FromFramework converts a Framework Int64 value to an int64 pointer.
// A null Int64 is converted to a nil int64 pointer.
func Int64FromFramework(_ context.Context, v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return aws.Int64(v.ValueInt64())
}

// Int64ToFramework converts an int64 pointer to a Framework Int64 value.
// A nil int64 pointer is converted to a null Int64.
func Int64ToFramework(_ context.Context, v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}

	return types.Int64Value(aws.ToInt64(v))
}

// Int64ToFrameworkLegacy converts an int64 pointer to a Framework Int64 value.
// A nil int64 pointer is converted to a zero Int64.
func Int64ToFrameworkLegacy(_ context.Context, v *int64) types.Int64 {
	return types.Int64Value(aws.ToInt64(v))
}

func Int64FromFrameworkLegacy(_ context.Context, v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	i := v.ValueInt64()
	if i == 0 {
		return nil
	}

	return aws.Int64(i)
}
