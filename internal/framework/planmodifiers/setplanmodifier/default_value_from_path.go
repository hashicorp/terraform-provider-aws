// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package setplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// DefaultValueFromPath returns a plan modifier that sets a set's default value
// from the planned value at another path.
func DefaultValueFromPath[T basetypes.SetValuable](path path.Path) planmodifier.Set {
	return defaultValueFromPath[T]{
		path: path,
	}
}

type defaultValueFromPath[T basetypes.SetValuable] struct {
	path path.Path
}

func (m defaultValueFromPath[T]) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m defaultValueFromPath[T]) MarkdownDescription(context.Context) string {
	return fmt.Sprintf("The default value of this attribute is %[1]q's value.", m.path)
}

func (m defaultValueFromPath[T]) PlanModifySet(ctx context.Context, request planmodifier.SetRequest, response *planmodifier.SetResponse) {
	// Do nothing if there is a known planned value.
	if !request.PlanValue.IsUnknown() {
		return
	}

	var t T
	response.Diagnostics.Append(request.Plan.GetAttribute(ctx, m.path, &t)...)
	if response.Diagnostics.HasError() {
		return
	}

	v, diags := t.ToSetValue(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.PlanValue = v
}
