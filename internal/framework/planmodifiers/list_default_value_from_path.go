// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ListDefaultValueFromPath returns a plan modifier that sets a list's default value
// from the planned value at another path.
func ListDefaultValueFromPath[T basetypes.ListValuable](path path.Path) planmodifier.List {
	return listDefaultValueFromPath[T]{
		path: path,
	}
}

type listDefaultValueFromPath[T basetypes.ListValuable] struct {
	path path.Path
}

func (m listDefaultValueFromPath[T]) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m listDefaultValueFromPath[T]) MarkdownDescription(context.Context) string {
	return "The default value of this attribute is another attribute's value."
}

func (m listDefaultValueFromPath[T]) PlanModifyList(ctx context.Context, request planmodifier.ListRequest, response *planmodifier.ListResponse) {
	// Do nothing if there is a known planned value.
	if !request.PlanValue.IsUnknown() {
		return
	}

	var t T
	response.Diagnostics.Append(request.Plan.GetAttribute(ctx, m.path, &t)...)
	if response.Diagnostics.HasError() {
		return
	}

	v, diags := t.ToListValue(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.PlanValue = v
}
