// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// BoolFromFramework converts a Framework Bool value to a bool pointer.
// A null Bool is converted to a nil bool pointer.
func BoolFromFramework(ctx context.Context, v basetypes.BoolValuable) *bool {
	var output *bool

	must(Expand(ctx, v, &output))

	return output
}

func BoolValueFromFramework(ctx context.Context, v basetypes.BoolValuable) bool {
	var output bool

	must(Expand(ctx, v, &output))

	return output
}

// BoolToFramework converts a bool pointer to a Framework Bool value.
// A nil bool pointer is converted to a null Bool.
func BoolToFramework(ctx context.Context, v *bool) types.Bool {
	var output types.Bool

	must(Flatten(ctx, v, &output))

	return output
}

// BoolToFrameworkLegacy converts a bool pointer to a Framework Bool value.
// A nil bool pointer is converted to a false Bool.
func BoolToFrameworkLegacy(_ context.Context, v *bool) types.Bool {
	return types.BoolValue(aws.ToBool(v))
}
