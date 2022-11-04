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
		elems[i] = types.String{Value: aws.ToString(v)}
	}

	return types.List{ElemType: types.StringType, Elems: elems}
}

// FlattenFrameworkStringValueList is the Plugin Framework variant of FlattenStringValueList.
// In particular, a nil slice is converted to an empty (non-null) List.
func FlattenFrameworkStringValueList(_ context.Context, vs []string) types.List {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.String{Value: v}
	}

	return types.List{ElemType: types.StringType, Elems: elems}
}

// FlattenFrameworkStringValueSet is the Plugin Framework variant of FlattenStringValueSet.
// In particular, a nil slice is converted to an empty (non-null) Set.
func FlattenFrameworkStringValueSet(_ context.Context, vs []string) types.Set {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.String{Value: v}
	}

	return types.Set{ElemType: types.StringType, Elems: elems}
}

// FlattenFrameworkStringValueMap has no Plugin SDK equivalent as schema.ResourceData.Set can be passed string value maps directly.
// In particular, a nil map is converted to an empty (non-null) Map.
func FlattenFrameworkStringValueMap(_ context.Context, m map[string]string) types.Map {
	elems := make(map[string]attr.Value, len(m))

	for k, v := range m {
		elems[k] = types.String{Value: v}
	}

	return types.Map{ElemType: types.StringType, Elems: elems}
}

// ToFrameworkStringValue converts a string pointer to a Framework String value.
// A nil string pointer is converted to a null String.
func ToFrameworkStringValue(_ context.Context, v *string) types.String {
	if v == nil {
		return types.String{Null: true}
	}

	return types.String{Value: *v}
}
