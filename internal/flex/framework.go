package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Terraform Plugin Framework variants of standard flatteners and expanders.

func ExpandFrameworkStringList(ctx context.Context, list types.List) []*string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var vl []*string

	if list.ElementsAs(ctx, &vl, false).HasError() {
		return nil
	}

	return vl
}

func ExpandFrameworkStringValueList(ctx context.Context, list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var vl []string

	if list.ElementsAs(ctx, &vl, false).HasError() {
		return nil
	}

	return vl
}

func ExpandFrameworkStringSet(ctx context.Context, set types.Set) []*string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}

	var vs []*string

	if set.ElementsAs(ctx, &vs, false).HasError() {
		return nil
	}

	return vs
}

func ExpandFrameworkStringValueSet(ctx context.Context, set types.Set) []string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}

	var vs []string

	if set.ElementsAs(ctx, &vs, false).HasError() {
		return nil
	}

	return vs
}

func ExpandFrameworkStringValueMap(ctx context.Context, set types.Map) map[string]string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}

	var m map[string]string

	if set.ElementsAs(ctx, &m, false).HasError() {
		return nil
	}

	return m
}

// FlattenFrameworkStringList converts a slice of string pointers to a framework List value.
//
// A nil slice is converted to a null List.
// An empty slice is converted to a null List.
func FlattenFrameworkStringList(_ context.Context, vs []*string) types.List {
	if len(vs) == 0 {
		return types.ListNull(types.StringType)
	}

	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.StringValue(aws.ToString(v))
	}

	return types.ListValueMust(types.StringType, elems)
}

// FlattenFrameworkStringListLegacy is the Plugin Framework variant of FlattenStringList.
// A nil slice is converted to an empty (non-null) List.
func FlattenFrameworkStringListLegacy(_ context.Context, vs []*string) types.List {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.StringValue(aws.ToString(v))
	}

	return types.ListValueMust(types.StringType, elems)
}

// FlattenFrameworkStringValueList converts a slice of string values to a framework List value.
//
// A nil slice is converted to a null List.
// An empty slice is converted to a null List.
func FlattenFrameworkStringValueList(_ context.Context, vs []string) types.List {
	if len(vs) == 0 {
		return types.ListNull(types.StringType)
	}

	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.StringValue(v)
	}

	return types.ListValueMust(types.StringType, elems)
}

// FlattenFrameworkStringValueListLegacy is the Plugin Framework variant of FlattenStringValueList.
// A nil slice is converted to an empty (non-null) List.
func FlattenFrameworkStringValueListLegacy(_ context.Context, vs []string) types.List {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.StringValue(v)
	}

	return types.ListValueMust(types.StringType, elems)
}

// FlattenFrameworkStringValueSet converts a slice of string values to a framework Set value.
//
// A nil slice is converted to a null Set.
// An empty slice is converted to a null Set.
func FlattenFrameworkStringValueSet(_ context.Context, vs []string) types.Set {
	if len(vs) == 0 {
		return types.SetNull(types.StringType)
	}

	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.StringValue(v)
	}

	return types.SetValueMust(types.StringType, elems)
}

// FlattenFrameworkStringValueSetLegacy is the Plugin Framework variant of FlattenStringValueSet.
// A nil slice is converted to an empty (non-null) Set.
func FlattenFrameworkStringValueSetLegacy(_ context.Context, vs []string) types.Set {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.StringValue(v)
	}

	return types.SetValueMust(types.StringType, elems)
}

// FlattenFrameworkStringValueMapLegacy has no Plugin SDK equivalent as schema.ResourceData.Set can be passed string value maps directly.
// A nil map is converted to an empty (non-null) Map.
func FlattenFrameworkStringValueMapLegacy(_ context.Context, m map[string]string) types.Map {
	elems := make(map[string]attr.Value, len(m))

	for k, v := range m {
		elems[k] = types.StringValue(v)
	}

	return types.MapValueMust(types.StringType, elems)
}

// BoolFromFramework converts a Framework Bool value to a bool pointer.
// A null Bool is converted to a nil bool pointer.
func BoolFromFramework(_ context.Context, v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return aws.Bool(v.ValueBool())
}

// Int64FromFramework converts a Framework Int64 value to an int64 pointer.
// A null Int64 is converted to a nil int64 pointer.
func Int64FromFramework(_ context.Context, v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return aws.Int64(v.ValueInt64())
}

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

// BoolToFramework converts a bool pointer to a Framework Bool value.
// A nil bool pointer is converted to a null Bool.
func BoolToFramework(_ context.Context, v *bool) types.Bool {
	if v == nil {
		return types.BoolNull()
	}

	return types.BoolValue(aws.ToBool(v))
}

// BoolToFrameworkLegacy converts a bool pointer to a Framework Bool value.
// A nil bool pointer is converted to a false Bool.
func BoolToFrameworkLegacy(_ context.Context, v *bool) types.Bool {
	return types.BoolValue(aws.ToBool(v))
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
