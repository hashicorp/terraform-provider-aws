// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

// AttributeTypes returns a map of attribute types for the specified type T.
// T must be a struct and reflection is used to find exported fields of T with the `tfsdk` tag.
func AttributeTypes[T any](ctx context.Context) (map[string]attr.Type, diag.Diagnostics) {
	var diags diag.Diagnostics
	var t T
	attributeTypes := make(map[string]attr.Type)

	diags.Append(walkStruct(ctx, t, func(_ context.Context, tag string, val reflect.Value) diag.Diagnostics {
		var diags diag.Diagnostics

		if v, ok := val.Interface().(attr.Value); ok {
			attributeTypes[tag] = v.Type(ctx)
		}

		return diags
	})...)

	return attributeTypes, diags
}

func AttributeTypesMust[T any](ctx context.Context) map[string]attr.Type {
	return fwdiag.Must(AttributeTypes[T](ctx))
}

func newAttrTypeOf[T attr.Value](ctx context.Context) attr.Type {
	var zero T
	return zero.Type(ctx)
}
