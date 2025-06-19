// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"fmt"
	"unique"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/interceptors"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// tagsDataSourceInterceptor implements transparent tagging for data sources.
type tagsDataSourceInterceptor struct {
	interceptors.HTags
}

func (r tagsDataSourceInterceptor) read(ctx context.Context, opts interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	if !r.Enabled() {
		return diags
	}

	sp, serviceName, resourceName, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return diags
	}

	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		var configTags tftags.Map
		diags.Append(request.Config.GetAttribute(ctx, path.Root(names.AttrTags), &configTags)...)
		if diags.HasError() {
			return diags
		}

		tags := tftags.New(ctx, configTags)
		tagsInContext.TagsIn = option.Some(tags)
	case After:
		// If the R handler didn't set tags, try and read them from the service API.
		if tagsInContext.TagsOut.IsNone() {
			if identifier := r.GetIdentifierFramework(ctx, response.State); identifier != "" {
				if err := r.ListTags(ctx, sp, c, identifier); err != nil {
					diags.AddError(fmt.Sprintf("listing tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())
					return diags
				}
			}
		}

		tags := tagsInContext.TagsOut.UnwrapOrDefault()
		// Remove any provider configured ignore_tags and system tags from those returned from the service API.
		stateTags := fwflex.FlattenFrameworkStringValueMapLegacy(ctx, tags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx)).Map())
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTags), tftags.NewMapFromMapValue(stateTags))...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

func dataSourceTransparentTagging(servicePackageResourceTags unique.Handle[inttypes.ServicePackageResourceTags]) dataSourceCRUDInterceptor {
	return &tagsDataSourceInterceptor{
		HTags: interceptors.HTags(servicePackageResourceTags),
	}
}

// tagsResourceInterceptor implements transparent tagging for resources.
type tagsResourceInterceptor struct {
	resourceNoOpCRUDInterceptor
	interceptors.HTags
}

func (r tagsResourceInterceptor) create(ctx context.Context, opts interceptorOptions[resource.CreateRequest, resource.CreateResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	if !r.Enabled() {
		return diags
	}

	sp, _, _, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return diags
	}

	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		var planTags tftags.Map
		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)
		if diags.HasError() {
			return diags
		}

		// Merge the resource's configured tags with any provider configured default_tags.
		tags := c.DefaultTagsConfig(ctx).MergeTags(tftags.New(ctx, planTags))
		// Remove system tags.
		tags = tags.IgnoreSystem(sp.ServicePackageName())
		tagsInContext.TagsIn = option.Some(tags)
	case After:
		// Set values for unknowns.
		// Remove any provider configured ignore_tags and system tags from those passed to the service API.
		// Computed tags_all include any provider configured default_tags.
		stateTagsAll := fwflex.FlattenFrameworkStringValueMapLegacy(ctx, tagsInContext.TagsIn.MustUnwrap().IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx)).Map())
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.NewMapFromMapValue(stateTagsAll))...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

func (r tagsResourceInterceptor) read(ctx context.Context, opts interceptorOptions[resource.ReadRequest, resource.ReadResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	if !r.Enabled() {
		return diags
	}

	sp, serviceName, resourceName, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return diags
	}

	switch response, when := opts.response, opts.when; when {
	case After:
		// Will occur on a refresh when the resource does not exist in AWS and needs to be recreated, e.g. "_disappears" tests.
		if response.State.Raw.IsNull() {
			return diags
		}

		// If the R handler didn't set tags, try and read them from the service API.
		if tagsInContext.TagsOut.IsNone() {
			// Some old resources may not have the required attribute set after Read:
			// https://github.com/hashicorp/terraform-provider-aws/issues/31180
			if identifier := r.GetIdentifierFramework(ctx, response.State); identifier != "" {
				if err := r.ListTags(ctx, sp, c, identifier); err != nil {
					diags.AddError(fmt.Sprintf("listing tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())

					return diags
				}
			}
		}

		apiTags := tagsInContext.TagsOut.UnwrapOrDefault()

		// AWS APIs often return empty lists of tags when none have been configured.
		var stateTags tftags.Map
		response.State.GetAttribute(ctx, path.Root(names.AttrTags), &stateTags)
		// Remove any provider configured ignore_tags and system tags from those returned from the service API.
		// The resource's configured tags do not include any provider configured default_tags.
		if v := apiTags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx)).ResolveDuplicatesFramework(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx), response, &diags).Map(); len(v) > 0 {
			stateTags = tftags.NewMapFromMapValue(fwflex.FlattenFrameworkStringValueMapLegacy(ctx, v))
		}
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTags), &stateTags)...)
		if diags.HasError() {
			return diags
		}

		// Computed tags_all do.
		stateTagsAll := fwflex.FlattenFrameworkStringValueMapLegacy(ctx, apiTags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx)).Map())
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.NewMapFromMapValue(stateTagsAll))...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

func (r tagsResourceInterceptor) update(ctx context.Context, opts interceptorOptions[resource.UpdateRequest, resource.UpdateResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	if !r.Enabled() {
		return diags
	}

	sp, serviceName, resourceName, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return diags
	}

	switch request, when := opts.request, opts.when; when {
	case Before:
		var planTags tftags.Map
		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)
		if diags.HasError() {
			return diags
		}

		// Merge the resource's configured tags with any provider configured default_tags.
		tags := c.DefaultTagsConfig(ctx).MergeTags(tftags.New(ctx, planTags))
		// Remove system tags.
		tags = tags.IgnoreSystem(sp.ServicePackageName())
		tagsInContext.TagsIn = option.Some(tags)

		var oldTagsAll, newTagsAll tftags.Map
		diags.Append(request.State.GetAttribute(ctx, path.Root(names.AttrTagsAll), &oldTagsAll)...)
		if diags.HasError() {
			return diags
		}
		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTagsAll), &newTagsAll)...)
		if diags.HasError() {
			return diags
		}

		if !newTagsAll.Equal(oldTagsAll) {
			// Some old resources may not have the required attribute set after Read:
			// https://github.com/hashicorp/terraform-provider-aws/issues/31180
			if identifier := r.GetIdentifierFramework(ctx, request.Plan); identifier != "" {
				if err := r.UpdateTags(ctx, sp, c, identifier, oldTagsAll, newTagsAll); err != nil {
					diags.AddError(fmt.Sprintf("updating tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())

					return diags
				}
			}
			// TODO If the only change was to tags it would be nice to not call the resource's U handler.
		}
	}

	return diags
}

func (r tagsResourceInterceptor) modifyPlan(ctx context.Context, opts interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		// If the entire plan is null, the resource is planned for destruction.
		if request.Plan.Raw.IsNull() {
			return diags
		}

		// Calculate the new value for the `tags_all` attribute.
		var planTags tftags.Map
		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)
		if diags.HasError() {
			return diags
		}

		if planTags.IsWhollyKnown() {
			allTags := c.DefaultTagsConfig(ctx).MergeTags(tftags.New(ctx, planTags)).IgnoreConfig(c.IgnoreTagsConfig(ctx))
			diags.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrTagsAll), fwflex.FlattenFrameworkStringValueMapLegacy(ctx, allTags.Map()))...)
		} else {
			diags.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.Unknown)...)
		}

		if diags.HasError() {
			return diags
		}
	}

	return diags
}

func resourceTransparentTagging(servicePackageResourceTags unique.Handle[inttypes.ServicePackageResourceTags]) interface {
	resourceCRUDInterceptor
	resourceModifyPlanInterceptor
} {
	return &tagsResourceInterceptor{
		HTags: interceptors.HTags(servicePackageResourceTags),
	}
}
