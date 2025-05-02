// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// StringFromFramework converts a Framework String value to a string pointer.
// A null String is converted to a nil string pointer.
func StringFromFramework(ctx context.Context, v basetypes.StringValuable) *string {
	if v.IsUnknown() {
		return nil
	}
	val := fwdiag.Must(v.ToStringValue(ctx))
	return val.ValueStringPointer()
}

// StringValueFromFramework converts a Framework String value to a string.
// A null String is converted to an empty string.
func StringValueFromFramework(ctx context.Context, v basetypes.StringValuable) string {
	val := fwdiag.Must(v.ToStringValue(ctx))
	return val.ValueString()
}

// StringSliceValueFromFramework converts a single Framework String value to a string slice.
// A null String is converted to a nil slice.
func StringSliceValueFromFramework(ctx context.Context, v basetypes.StringValuable) []string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return []string{StringValueFromFramework(ctx, v)}
}

// StringToFramework converts a string pointer to a Framework String value.
// A nil string pointer is converted to a null String.
func StringToFramework(ctx context.Context, v *string) types.String {
	return types.StringPointerValue(v)
}

// StringValueToFramework converts a string value to a Framework String value.
// An empty string is converted to a null String.
func StringValueToFramework[T ~string](ctx context.Context, v T) types.String {
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

// StringToFrameworkLegacy converts a string pointer to a Framework String value.
// A nil string pointer is converted to an empty String.
func StringToFrameworkLegacy(_ context.Context, v *string) types.String {
	return types.StringValue(aws.ToString(v))
}

// StringToFrameworkARN converts a string pointer to a Framework custom ARN value.
// A nil string pointer is converted to a null ARN.
func StringToFrameworkARN(ctx context.Context, v *string) fwtypes.ARN {
	return StringToFrameworkValuable[fwtypes.ARN](ctx, v)
}

// StringToFrameworkValuable converts a string pointer to a Framework StringValuable value.
// A nil string pointer is converted to a null StringValuable.
func StringToFrameworkValuable[T basetypes.StringValuable](ctx context.Context, v *string) T {
	var output T
	typ := output.Type(ctx)
	styp := typ.(basetypes.StringTypable)

	sv := fwdiag.Must(styp.ValueFromString(ctx, types.StringPointerValue(v)))
	return sv.(T)
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

func EmptyStringAsNull(v types.String) types.String {
	if v.IsNull() || v.IsUnknown() {
		return v
	}

	if v.ValueString() == "" {
		return types.StringNull()
	}

	return v
}
