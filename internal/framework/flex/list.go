package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
