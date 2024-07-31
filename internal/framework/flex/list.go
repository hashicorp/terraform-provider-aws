// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func ExpandFrameworkInt32List(ctx context.Context, v basetypes.ListValuable) []*int32 {
	var output []*int32

	must(Expand(ctx, v, &output))

	return output
}

func ExpandFrameworkInt32ValueList(ctx context.Context, v basetypes.ListValuable) []int32 {
	var output []int32

	must(Expand(ctx, v, &output))

	return output
}

func ExpandFrameworkInt64List(ctx context.Context, v basetypes.ListValuable) []*int64 {
	var output []*int64

	must(Expand(ctx, v, &output))

	return output
}

func ExpandFrameworkInt64ValueList(ctx context.Context, v basetypes.ListValuable) []int64 {
	var output []int64

	must(Expand(ctx, v, &output))

	return output
}

func ExpandFrameworkStringList(ctx context.Context, v basetypes.ListValuable) []*string {
	var output []*string

	must(Expand(ctx, v, &output))

	return output
}

func ExpandFrameworkStringValueList(ctx context.Context, v basetypes.ListValuable) []string {
	var output []string

	must(Expand(ctx, v, &output))

	return output
}

// FlattenFrameworkInt64List converts a slice of int32 pointers to a framework List value.
//
// A nil slice is converted to a null List.
// An empty slice is converted to a null List.
func FlattenFrameworkInt32List(ctx context.Context, v []*int32) types.List {
	if len(v) == 0 {
		return types.ListNull(types.Int64Type)
	}

	var output types.List

	must(Flatten(ctx, v, &output))

	return output
}

// FlattenFrameworkInt64ValueList converts a slice of int32 values to a framework List value.
//
// A nil slice is converted to a null List.
// An empty slice is converted to a null List.
func FlattenFrameworkInt32ValueList[T ~int32](ctx context.Context, v []T) types.List {
	if len(v) == 0 {
		return types.ListNull(types.Int64Type)
	}

	var output types.List

	must(Flatten(ctx, v, &output))

	return output
}

// FlattenFrameworkInt64List converts a slice of int64 pointers to a framework List value.
//
// A nil slice is converted to a null List.
// An empty slice is converted to a null List.
func FlattenFrameworkInt64List(ctx context.Context, v []*int64) types.List {
	if len(v) == 0 {
		return types.ListNull(types.Int64Type)
	}

	var output types.List

	must(Flatten(ctx, v, &output))

	return output
}

// FlattenFrameworkInt64ValueList converts a slice of int64 values to a framework List value.
//
// A nil slice is converted to a null List.
// An empty slice is converted to a null List.
func FlattenFrameworkInt64ValueList[T ~int64](ctx context.Context, v []T) types.List {
	if len(v) == 0 {
		return types.ListNull(types.Int64Type)
	}

	var output types.List

	must(Flatten(ctx, v, &output))

	return output
}

// FlattenFrameworkStringList converts a slice of string pointers to a framework List value.
//
// A nil slice is converted to a null List.
// An empty slice is converted to a null List.
func FlattenFrameworkStringList(ctx context.Context, v []*string) types.List {
	if len(v) == 0 {
		return types.ListNull(types.StringType)
	}

	var output types.List

	must(Flatten(ctx, v, &output))

	return output
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
func FlattenFrameworkStringValueList[T ~string](ctx context.Context, v []T) types.List {
	if len(v) == 0 {
		return types.ListNull(types.StringType)
	}

	var output types.List

	must(Flatten(ctx, v, &output))

	return output
}

func FlattenFrameworkStringValueListOfString(ctx context.Context, vs []string) fwtypes.ListValueOf[basetypes.StringValue] {
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
