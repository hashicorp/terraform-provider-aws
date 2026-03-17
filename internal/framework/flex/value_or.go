// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// *ValueOr returns the value if it is not null or unknown, otherwise it returns the provided default value.
// Useful for filling in default values when no plan is available.

// Useful for filling in default int64 values when no plan is available.
func Int64ValueOr(ctx context.Context, v types.Int64, defaultValue int64) int64 {
	if v.IsNull() || v.IsUnknown() {
		return defaultValue
	}
	return v.ValueInt64()
}

// Useful for filling in default string values when no plan is available.
func StringValueOr(ctx context.Context, v types.String, defaultValue string) string {
	if v.IsNull() || v.IsUnknown() {
		return defaultValue
	}
	return v.ValueString()
}

// Useful for filling in default string enum values when no plan is available.
func StringEnumValueOr[T enum.Valueser[T]](ctx context.Context, v fwtypes.StringEnum[T], defaultValue T) T {
	if v.IsNull() || v.IsUnknown() {
		return defaultValue
	}
	return v.ValueEnum()
}
