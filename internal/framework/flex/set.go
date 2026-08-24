// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// DiffSets returns the set difference a\b: elements present in a but absent in b.
// Neither a nor b may be Unknown. Null/Unknown values are treated as empty.
// The returned set uses the element type of a.
func DiffSets(ctx context.Context, a, b types.Set) types.Set {
	elemType := a.ElementType(ctx)

	aElems := a.Elements()
	if len(aElems) == 0 {
		return types.SetValueMust(elemType, []attr.Value{})
	}

	bElems := b.Elements()
	if len(bElems) == 0 {
		return types.SetValueMust(elemType, aElems)
	}

	var diff []attr.Value
	for _, av := range aElems {
		if inB := slices.ContainsFunc(bElems, av.Equal); !inB {
			diff = append(diff, av)
		}
	}
	if diff == nil {
		diff = []attr.Value{}
	}

	return types.SetValueMust(elemType, diff)
}

func ExpandFrameworkStringValueSet(ctx context.Context, v basetypes.SetValuable) inttypes.Set[string] {
	return ExpandFrameworkStringyValueSet[string](ctx, v)
}

func ExpandFrameworkStringyValueSet[E ~string](ctx context.Context, v basetypes.SetValuable) inttypes.Set[E] {
	var output []E

	must(Expand(ctx, v, &output))

	return output
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

func FlattenFrameworkStringValueSetOfString(ctx context.Context, vs []string) fwtypes.SetOfString {
	return fwtypes.SetValueOf[basetypes.StringValue]{SetValue: FlattenFrameworkStringValueSet(ctx, vs)}
}

func FlattenFrameworkStringyValueSetOfStringEnum[T enum.Valueser[T]](ctx context.Context, vs []T) fwtypes.SetOfStringEnum[T] {
	return fwtypes.SetValueOf[fwtypes.StringEnum[T]]{SetValue: FlattenFrameworkStringValueSet(ctx, vs)}
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

func FlattenFrameworkStringValueSetOfStringLegacy(ctx context.Context, vs []string) fwtypes.SetOfString {
	return fwtypes.SetValueOf[basetypes.StringValue]{SetValue: FlattenFrameworkStringValueSetLegacy(ctx, vs)}
}
