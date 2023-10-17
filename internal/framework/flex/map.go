// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ExpandFrameworkStringMap(ctx context.Context, v types.Map) map[string]*string {
	var output map[string]*string

	panicOnError(Expand(ctx, v, &output))

	return output
}

func ExpandFrameworkStringValueMap(ctx context.Context, v types.Map) map[string]string {
	var output map[string]string

	panicOnError(Expand(ctx, v, &output))

	return output
}

// FlattenFrameworkStringMap converts a map of string pointers to a framework Map value.
//
// A nil map is converted to a null Map.
// An empty map is converted to a null Map.
func FlattenFrameworkStringMap(ctx context.Context, v map[string]*string) types.Map {
	if len(v) == 0 {
		return types.MapNull(types.StringType)
	}

	var output types.Map

	panicOnError(Flatten(ctx, v, &output))

	return output
}

// FlattenFrameworkStringValueMap converts a map of strings to a framework Map value.
//
// A nil map is converted to a null Map.
// An empty map is converted to a null Map.
func FlattenFrameworkStringValueMap(ctx context.Context, v map[string]string) types.Map {
	if len(v) == 0 {
		return types.MapNull(types.StringType)
	}

	var output types.Map

	panicOnError(Flatten(ctx, v, &output))

	return output
}

// FlattenFrameworkStringValueMapLegacy has no Plugin SDK equivalent as schema.ResourceData.Set can be passed string value maps directly.
// A nil map is converted to an empty (non-null) Map.
func FlattenFrameworkStringValueMapLegacy(_ context.Context, m map[string]string) types.Map {
	elems := make(map[string]attr.Value, len(m))

	for k, v := range m {
		elems[k] = types.StringValue(v)
	}

	return types.MapValueMust(types.StringType, elems)
}
