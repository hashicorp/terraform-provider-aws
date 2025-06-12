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

// BoolFromFramework converts a Framework Bool value to a bool pointer.
// A null Bool is converted to a nil bool pointer.
func BoolFromFramework(ctx context.Context, v basetypes.BoolValuable) *bool {
	if v.IsUnknown() {
		return nil
	}
	val := fwdiag.Must(v.ToBoolValue(ctx))
	return val.ValueBoolPointer()
}

func BoolValueFromFramework(ctx context.Context, v basetypes.BoolValuable) bool {
	val := fwdiag.Must(v.ToBoolValue(ctx))
	return val.ValueBool()
}

// BoolToFramework converts a bool pointer to a Framework Bool value.
// A nil bool pointer is converted to a null Bool.
func BoolToFramework(ctx context.Context, v *bool) types.Bool {
	return types.BoolPointerValue(v)
}

// BoolValueToFramework converts a bool value to a Framework Bool value.
func BoolValueToFramework(ctx context.Context, v bool) types.Bool {
	return types.BoolValue(v)
}

// BoolToFrameworkLegacy converts a bool pointer to a Framework Bool value.
// A nil bool pointer is converted to a false Bool.
func BoolToFrameworkLegacy(_ context.Context, v *bool) types.Bool {
	return types.BoolValue(aws.ToBool(v))
}
