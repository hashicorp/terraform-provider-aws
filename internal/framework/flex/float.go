// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

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
