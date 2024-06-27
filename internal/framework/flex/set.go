// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func ExpandFrameworkInt32Set(ctx context.Context, v basetypes.SetValuable) []*int32 {
	var output []*int32

	must(Expand(ctx, v, &output))

	return output
}

func ExpandFrameworkInt32ValueSet(ctx context.Context, v basetypes.SetValuable) []int32 {
	var output []int32

	must(Expand(ctx, v, &output))

	return output
}

func ExpandFrameworkInt64Set(ctx context.Context, v basetypes.SetValuable) []*int64 {
	var output []*int64

	must(Expand(ctx, v, &output))

	return output
}

func ExpandFrameworkInt64ValueSet(ctx context.Context, v basetypes.SetValuable) []int64 {
	var output []int64

	must(Expand(ctx, v, &output))

	return output
}

func ExpandFrameworkStringSet(ctx context.Context, v basetypes.SetValuable) []*string {
	var output []*string

	must(Expand(ctx, v, &output))

	return output
}

func ExpandFrameworkStringValueSet(ctx context.Context, v basetypes.SetValuable) itypes.Set[string] {
	var output []string

	must(Expand(ctx, v, &output))

	return output
}

func ExpandFrameworkStringyValueSet[T ~string](ctx context.Context, v basetypes.SetValuable) itypes.Set[T] {
	vs := ExpandFrameworkStringValueSet(ctx, v)
	if vs == nil {
		return nil
	}
	return tfslices.ApplyToAll(vs, func(s string) T { return T(s) })
}

// FlattenFrameworkInt64Set converts a slice of int32 pointers to a framework Set value.
//
// A nil slice is converted to a null Set.
// An empty slice is converted to a null Set.
func FlattenFrameworkInt32Set(ctx context.Context, v []*int32) types.Set {
	if len(v) == 0 {
		return types.SetNull(types.Int64Type)
	}

	var output types.Set

	must(Flatten(ctx, v, &output))

	return output
}

// FlattenFrameworkInt64ValueSet converts a slice of int32 values to a framework Set value.
//
// A nil slice is converted to a null Set.
// An empty slice is converted to a null Set.
func FlattenFrameworkInt32ValueSet[T ~int32](ctx context.Context, v []T) types.Set {
	if len(v) == 0 {
		return types.SetNull(types.Int64Type)
	}

	var output types.Set

	must(Flatten(ctx, v, &output))

	return output
}

// FlattenFrameworkInt64Set converts a slice of int64 pointers to a framework Set value.
//
// A nil slice is converted to a null Set.
// An empty slice is converted to a null Set.
func FlattenFrameworkInt64Set(ctx context.Context, v []*int64) types.Set {
	if len(v) == 0 {
		return types.SetNull(types.Int64Type)
	}

	var output types.Set

	must(Flatten(ctx, v, &output))

	return output
}

// FlattenFrameworkInt64ValueSet converts a slice of int64 values to a framework Set value.
//
// A nil slice is converted to a null Set.
// An empty slice is converted to a null Set.
func FlattenFrameworkInt64ValueSet[T ~int64](ctx context.Context, v []T) types.Set {
	if len(v) == 0 {
		return types.SetNull(types.Int64Type)
	}

	var output types.Set

	must(Flatten(ctx, v, &output))

	return output
}

// FlattenFrameworkStringSet converts a slice of string pointers to a framework Set value.
//
// A nil slice is converted to a null Set.
// An empty slice is converted to a null Set.
func FlattenFrameworkStringSet(ctx context.Context, v []*string) types.Set {
	if len(v) == 0 {
		return types.SetNull(types.StringType)
	}

	var output types.Set

	must(Flatten(ctx, v, &output))

	return output
}

// FlattenFrameworkStringSetLegacy converts a slice of string pointers to a framework Set value.
//
// A nil slice is converted to an empty (non-null) Set.
func FlattenFrameworkStringSetLegacy(_ context.Context, vs []*string) types.Set {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.StringValue(aws.ToString(v))
	}

	return types.SetValueMust(types.StringType, elems)
}

// FlattenFrameworkStringValueSet converts a slice of string values to a framework Set value.
//
// A nil slice is converted to a null Set.
// An empty slice is converted to a null Set.
func FlattenFrameworkStringValueSet[T ~string](ctx context.Context, v []T) types.Set {
	if len(v) == 0 {
		return types.SetNull(types.StringType)
	}

	var output types.Set

	must(Flatten(ctx, v, &output))

	return output
}

// FlattenFrameworkStringValueSetLegacy is the Plugin Framework variant of FlattenStringValueSet.
// A nil slice is converted to an empty (non-null) Set.
func FlattenFrameworkStringValueSetLegacy[T ~string](_ context.Context, vs []T) types.Set {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.StringValue(string(v))
	}

	return types.SetValueMust(types.StringType, elems)
}
