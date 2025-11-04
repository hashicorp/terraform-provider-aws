// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func ExpandFrameworkStringValueList(ctx context.Context, v basetypes.ListValuable) []string {
	var output []string

	must(Expand(ctx, v, &output))

	return output
}

// FlattenFrameworkStringValueList converts a slice of string values to a framework List value.
//
// A nil slice is converted to a null List.
// An empty slice is converted to a null List.
func FlattenFrameworkStringValueList[T ~string](ctx context.Context, v []T) types.List {
	if len(v) == 0 {
		return types.ListNull(types.StringType)
	}

	var output types.List

	must(Flatten(ctx, v, &output))

	return output
}

func FlattenFrameworkStringValueListOfString(ctx context.Context, vs []string) fwtypes.ListOfString {
	return fwtypes.ListValueOf[basetypes.StringValue]{ListValue: FlattenFrameworkStringValueList(ctx, vs)}
}

// FlattenFrameworkStringValueListLegacy is the Plugin Framework variant of FlattenStringValueList.
// A nil slice is converted to an empty (non-null) List.
func FlattenFrameworkStringValueListLegacy[T ~string](_ context.Context, vs []T) types.List {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.StringValue(string(v))
	}

	return types.ListValueMust(types.StringType, elems)
}

func FlattenFrameworkStringValueListOfStringLegacy(ctx context.Context, vs []string) fwtypes.ListOfString {
	return fwtypes.ListValueOf[basetypes.StringValue]{ListValue: FlattenFrameworkStringValueListLegacy(ctx, vs)}
}
