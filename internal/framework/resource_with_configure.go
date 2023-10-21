// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// ResourceWithConfigure is a structure to be embedded within a Resource that implements the ResourceWithConfigure interface.
type ResourceWithConfigure struct {
	withMeta
}

// Configure enables provider-level data or clients to be set in the
// provider-defined Resource type.
func (r *ResourceWithConfigure) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		r.meta = v
	}
}

// SetTagsAll calculates the new value for the `tags_all` attribute.
func (r *ResourceWithConfigure) SetTagsAll(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	// If the entire plan is null, the resource is planned for destruction.
	if request.Plan.Raw.IsNull() {
		return
	}

	defaultTagsConfig := r.Meta().DefaultTagsConfig
	ignoreTagsConfig := r.Meta().IgnoreTagsConfig

	var planTags types.Map

	response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)

	if response.Diagnostics.HasError() {
		return
	}

	if !planTags.IsUnknown() {
		if !mapHasUnknownElements(planTags) {
			resourceTags := tftags.New(ctx, planTags)
			allTags := defaultTagsConfig.MergeTags(resourceTags).IgnoreConfig(ignoreTagsConfig)

			response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrTagsAll), flex.FlattenFrameworkStringValueMapLegacy(ctx, allTags.Map()))...)
		} else {
			response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.Unknown)...)
		}
	} else {
		response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.Unknown)...)
	}
}

func mapHasUnknownElements(m types.Map) bool {
	for _, v := range m.Elements() {
		if v.IsUnknown() {
			return true
		}
	}

	return false
}
