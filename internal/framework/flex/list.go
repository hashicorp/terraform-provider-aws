// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
)

func ExpandFrameworkStringList(ctx context.Context, v types.List) []*string {
	var output []*string

	panicOnError(Expand(ctx, v, &output))

	return output
}

func ExpandFrameworkStringValueList(ctx context.Context, v types.List) []string {
	var output []string

	panicOnError(Expand(ctx, v, &output))

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

	panicOnError(Flatten(ctx, v, &output))

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
func FlattenFrameworkStringValueList(ctx context.Context, v []string) types.List {
	if len(v) == 0 {
		return types.ListNull(types.StringType)
	}

	var output types.List

	panicOnError(Flatten(ctx, v, &output))

	return output
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

type FrameworkElementExpanderFunc[T any, U any] func(context.Context, T) U

func ExpandFrameworkListNestedBlock[T any, U any](ctx context.Context, tfList types.List, f FrameworkElementExpanderFunc[T, U]) []U {
	if tfList.IsNull() || tfList.IsUnknown() {
		return nil
	}

	var data []T

	_ = fwdiag.Must(0, tfList.ElementsAs(ctx, &data, false))

	return slices.ApplyToAll(data, func(t T) U {
		return f(ctx, t)
	})
}

func ExpandFrameworkListNestedBlockPtr[T any, U any](ctx context.Context, tfList types.List, f FrameworkElementExpanderFunc[T, *U]) *U {
	if tfList.IsNull() || tfList.IsUnknown() {
		return nil
	}

	var data []T

	_ = fwdiag.Must(0, tfList.ElementsAs(ctx, &data, false))

	if len(data) == 0 {
		return nil
	}

	return f(ctx, data[0])
}

type FrameworkElementFlattenerFunc[T any, U any] func(context.Context, U) T

func FlattenFrameworkListNestedBlock[T any, U any](ctx context.Context, apiObjects []U, f FrameworkElementFlattenerFunc[T, U]) types.List {
	attributeTypes := AttributeTypesMust[T](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	if len(apiObjects) == 0 {
		return types.ListNull(elementType)
	}

	data := slices.ApplyToAll(apiObjects, func(apiObject U) T {
		return f(ctx, apiObject)
	})

	return fwdiag.Must(types.ListValueFrom(ctx, elementType, data))
}
