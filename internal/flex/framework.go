package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Terraform Plugin Framework variants of standard flatteners and expanders.

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

// FlattenFrameworkStringList is the Plugin Framework variant of FlattenStringList.
// In particular, a nil slice is converted to an empty (non-null) List.
func FlattenFrameworkStringList(_ context.Context, vs []*string) types.List {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.StringValue(aws.ToString(v))
	}

	return types.ListValueMust(types.StringType, elems)
}

// FlattenFrameworkStringValueList is the Plugin Framework variant of FlattenStringValueList.
// In particular, a nil slice is converted to an empty (non-null) List.
func FlattenFrameworkStringValueList(_ context.Context, vs []string) types.List {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.StringValue(v)
	}

	return types.ListValueMust(types.StringType, elems)
}

// FlattenFrameworkStringValueSet is the Plugin Framework variant of FlattenStringValueSet.
// In particular, a nil slice is converted to an empty (non-null) Set.
func FlattenFrameworkStringValueSet(_ context.Context, vs []string) types.Set {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.StringValue(v)
	}

	return types.SetValueMust(types.StringType, elems)
}

// FlattenFrameworkStringValueMap has no Plugin SDK equivalent as schema.ResourceData.Set can be passed string value maps directly.
// In particular, a nil map is converted to an empty (non-null) Map.
func FlattenFrameworkStringValueMap(_ context.Context, m map[string]string) types.Map {
	elems := make(map[string]attr.Value, len(m))

	for k, v := range m {
		elems[k] = types.StringValue(v)
	}

	return types.MapValueMust(types.StringType, elems)
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

// Int64ToFramework converts an int64 pointer to a Framework Int64 value.
// A nil int64 pointer is converted to a null Int64.
func Int64ToFramework(_ context.Context, v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}

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

// StringToFrameworkWithTransform converts a string pointer to a Framework String value.
// A nil string pointer is converted to a null String.
// A non-nil string pointer has its value transformed by `f`.
func StringToFrameworkWithTransform(_ context.Context, v *string, f func(string) string) types.String {
	if v == nil {
		return types.StringNull()
	}

	return types.StringValue(f(aws.ToString(v)))
}
