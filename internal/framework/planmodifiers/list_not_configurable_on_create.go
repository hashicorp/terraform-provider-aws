// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// ListNotConfigurableOnCreate returns a plan modifier that raises an error if:
//
//   - The resource is planned for create.
//   - The configuration value is not unknown.
//
// This plan modifier should be applied to an Optional+Computed list attribute.
// See e.g. framework.ResourceOptionalComputedListOfObjectAttribute.
func ListNotConfigurableOnCreate() planmodifier.List {
	return listNotConfigurableOnCreateModifier{}
}

type listNotConfigurableOnCreateModifier struct{}

func (m listNotConfigurableOnCreateModifier) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m listNotConfigurableOnCreateModifier) MarkdownDescription(context.Context) string {
	return "This attribute must not be configured when creating a new resource."
}

func (m listNotConfigurableOnCreateModifier) PlanModifyList(ctx context.Context, request planmodifier.ListRequest, response *planmodifier.ListResponse) {
	if request.State.Raw.IsNull() && !request.PlanValue.IsUnknown() {
		response.Diagnostics.AddAttributeError(request.Path, m.Description(ctx), "")
	}
}
