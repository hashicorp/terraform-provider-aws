// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func ExpandFrameworkStringValueSet(ctx context.Context, v basetypes.SetValuable) inttypes.Set[string] {
	var output []string

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
