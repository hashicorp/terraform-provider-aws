// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ExpandFrameworkStringMap(ctx context.Context, set types.Map) map[string]*string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}

	var m map[string]*string

	if set.ElementsAs(ctx, &m, false).HasError() {
		return nil
	}

	return m
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

// FlattenFrameworkStringMap converts a map of string pointers to a framework Map value.
//
// A nil map is converted to a null Map.
// An empty map is converted to a null Map.
func FlattenFrameworkStringMap(_ context.Context, m map[string]*string) types.Map {
	if len(m) == 0 {
		return types.MapNull(types.StringType)
	}

	elems := make(map[string]attr.Value, len(m))

	for k, v := range m {
		elems[k] = types.StringValue(aws.ToString(v))
	}

	return types.MapValueMust(types.StringType, elems)
}

// FlattenFrameworkStringValueMap converts a map of strings to a framework Map value.
//
// A nil map is converted to a null Map.
// An empty map is converted to a null Map.
func FlattenFrameworkStringValueMap(_ context.Context, m map[string]string) types.Map {
	if len(m) == 0 {
		return types.MapNull(types.StringType)
	}

	elems := make(map[string]attr.Value, len(m))

	for k, v := range m {
		elems[k] = types.StringValue(v)
	}

	return types.MapValueMust(types.StringType, elems)
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
