// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// StringFromFramework converts a Framework String value to a string pointer.
// A null String is converted to a nil string pointer.
func StringFromFramework(_ context.Context, v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return aws.String(v.ValueString())
}

// StringFromFramework converts a single Framework String value to a string pointer slice.
// A null String is converted to a nil slice.
func StringSliceFromFramework(_ context.Context, v types.String) []*string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return aws.StringSlice([]string{v.ValueString()})
}

// StringValueToFramework converts a string value to a Framework String value.
// An empty string is converted to a null String.
func StringValueToFramework[T ~string](_ context.Context, v T) types.String {
	if v == "" {
		return types.StringNull()
	}
	return types.StringValue(string(v))
}

// StringValueToFrameworkLegacy converts a string value to a Framework String value.
// An empty string is left as an empty String.
func StringValueToFrameworkLegacy[T ~string](_ context.Context, v T) types.String {
	return types.StringValue(string(v))
}

// StringToFramework converts a string pointer to a Framework String value.
// A nil string pointer is converted to a null String.
func StringToFramework(_ context.Context, v *string) types.String {
	if v == nil {
		return types.StringNull()
	}

	return types.StringValue(aws.ToString(v))
}

// StringToFrameworkLegacy converts a string pointer to a Framework String value.
// A nil string pointer is converted to an empty String.
func StringToFrameworkLegacy(_ context.Context, v *string) types.String {
	return types.StringValue(aws.ToString(v))
}

// StringToFrameworkWithTransform converts a string pointer to a Framework String value.
// A nil string pointer is converted to a null String.
// A non-nil string pointer has its value transformed by `f`.
func StringToFrameworkWithTransform(_ context.Context, v *string, f func(string) string) types.String {
	if v == nil {
		return types.StringNull()
	}

	return types.StringValue(f(aws.ToString(v)))
}

func ARNStringFromFramework(_ context.Context, v fwtypes.ARN) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return aws.String(v.ValueARN().String())
}

func StringToFrameworkARN(ctx context.Context, v *string, diags *diag.Diagnostics) fwtypes.ARN {
	if v == nil {
		return fwtypes.ARNNull()
	}

	a, err := arn.Parse(aws.ToString(v))
	if err != nil {
		diags.AddError(
			"Parsing Error",
			fmt.Sprintf("String %s cannot be parsed as an ARN.", aws.ToString(v)),
		)
	}

	return fwtypes.ARNValue(a)
}

func StringFromFrameworkLegacy(_ context.Context, v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	s := v.ValueString()
	if s == "" {
		return nil
	}

	return aws.String(s)
}
