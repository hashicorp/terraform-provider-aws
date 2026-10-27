// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// SetDifference returns the difference between 2 Terraform Plugin Framework Sets:
// elements present in a but absent in b.
// Neither a nor b may be Unknown.
// Null sets are treated as empty sets.
func SetDifference(ctx context.Context, a, b types.Set) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	if a.IsUnknown() || b.IsUnknown() {
		diags.AddError("Invalid Value", "Unknown set value")
		return types.SetUnknown(a.ElementType(ctx)), diags
	}

	elemType := a.ElementType(ctx)
	if bElemType := b.ElementType(ctx); !elemType.Equal(bElemType) {
		diags.AddError("Invalid Value", "Mismatched set element types: "+elemType.String()+", "+bElemType.String())
		return types.SetUnknown(a.ElementType(ctx)), diags
	}

	aElems := a.Elements()
	if len(aElems) == 0 {
		r, d := types.SetValue(elemType, []attr.Value{})
		diags.Append(d...)
		return r, diags
	}

	bElems := b.Elements()
	if len(bElems) == 0 {
		r, d := types.SetValue(elemType, aElems)
		diags.Append(d...)
		return r, diags
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

	r, d := types.SetValue(elemType, diff)
	diags.Append(d...)
	return r, diags
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
