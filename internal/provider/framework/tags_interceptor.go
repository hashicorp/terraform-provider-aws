// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"fmt"
	"slices"
	"unique"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
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

func (r tagsDataSourceInterceptor) read(ctx context.Context, opts interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]) {
	c := opts.c

	if !r.Enabled() {
		return
	}

	sp, serviceName, resourceName, _, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return
	}

	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		var configTags tftags.Map
		opts.response.Diagnostics.Append(request.Config.GetAttribute(ctx, path.Root(names.AttrTags), &configTags)...)
		if opts.response.Diagnostics.HasError() {
			return
		}

		tags := tftags.New(ctx, configTags)
		tagsInContext.TagsIn = option.Some(tags)
	case After:
		// If the R handler didn't set tags, try and read them from the service API.
		if tagsInContext.TagsOut.IsNone() {
			if identifier := r.GetIdentifierFramework(ctx, response.State); identifier != "" {
				if err := r.ListTags(ctx, sp, c, identifier); err != nil {
					opts.response.Diagnostics.AddError(fmt.Sprintf("listing tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())
					return
				}
			}
		}

		tags := tagsInContext.TagsOut.UnwrapOrDefault()
		// Remove any provider configured ignore_tags and system tags from those returned from the service API.
		stateTags := fwflex.FlattenFrameworkStringValueMapLegacy(ctx, tags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx)).Map())
		opts.response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTags), tftags.NewMapFromMapValue(stateTags))...)
		if opts.response.Diagnostics.HasError() {
			return
		}
	}
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

func (r tagsResourceInterceptor) create(ctx context.Context, opts interceptorOptions[resource.CreateRequest, resource.CreateResponse]) {
	c := opts.c

	if !r.Enabled() {
		return
	}

	sp, _, _, _, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return
	}

	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		var planTags tftags.Map
		opts.response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)
		if opts.response.Diagnostics.HasError() {
			return
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
		opts.response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.NewMapFromMapValue(stateTagsAll))...)
		if opts.response.Diagnostics.HasError() {
			return
		}
	}
}

func (r tagsResourceInterceptor) read(ctx context.Context, opts interceptorOptions[resource.ReadRequest, resource.ReadResponse]) {
	c := opts.c

	if !r.Enabled() {
		return
	}

	sp, serviceName, resourceName, _, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return
	}

	switch response, when := opts.response, opts.when; when {
	case After:
		// Will occur on a refresh when the resource does not exist in AWS and needs to be recreated, e.g. "_disappears" tests.
		if response.State.Raw.IsNull() {
			return
		}

		// If the R handler didn't set tags, try and read them from the service API.
		if tagsInContext.TagsOut.IsNone() {
			// Some old resources may not have the required attribute set after Read:
			// https://github.com/hashicorp/terraform-provider-aws/issues/31180
			if identifier := r.GetIdentifierFramework(ctx, response.State); identifier != "" {
				if err := r.ListTags(ctx, sp, c, identifier); err != nil {
					opts.response.Diagnostics.AddError(fmt.Sprintf("listing tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())

					return
				}
			}
		}

		apiTags := tagsInContext.TagsOut.UnwrapOrDefault()

		// AWS APIs often return empty lists of tags when none have been configured.
		var stateTags tftags.Map
		response.State.GetAttribute(ctx, path.Root(names.AttrTags), &stateTags)
		// Remove any provider configured ignore_tags and system tags from those returned from the service API.
		// The resource's configured tags do not include any provider configured default_tags.
		if v := apiTags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx)).ResolveDuplicatesFramework(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx), stateTags, &opts.response.Diagnostics).Map(); len(v) > 0 {
			stateTags = tftags.NewMapFromMapValue(fwflex.FlattenFrameworkStringValueMapLegacy(ctx, v))
		}
		opts.response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTags), &stateTags)...)
		if opts.response.Diagnostics.HasError() {
			return
		}

		// Computed tags_all do.
		stateTagsAll := fwflex.FlattenFrameworkStringValueMapLegacy(ctx, apiTags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx)).Map())
		opts.response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.NewMapFromMapValue(stateTagsAll))...)
		if opts.response.Diagnostics.HasError() {
			return
		}
	}
}

func (r tagsResourceInterceptor) update(ctx context.Context, opts interceptorOptions[resource.UpdateRequest, resource.UpdateResponse]) {
	c := opts.c

	if !r.Enabled() {
		return
	}

	sp, serviceName, resourceName, _, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return
	}

	switch request, when := opts.request, opts.when; when {
	case Before:
		var planTags tftags.Map
		opts.response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)
		if opts.response.Diagnostics.HasError() {
			return
		}

		// Merge the resource's configured tags with any provider configured default_tags.
		tags := c.DefaultTagsConfig(ctx).MergeTags(tftags.New(ctx, planTags))
		// Remove system tags.
		tags = tags.IgnoreSystem(sp.ServicePackageName())
		tagsInContext.TagsIn = option.Some(tags)

		var oldTagsAll, newTagsAll tftags.Map
		opts.response.Diagnostics.Append(request.State.GetAttribute(ctx, path.Root(names.AttrTagsAll), &oldTagsAll)...)
		if opts.response.Diagnostics.HasError() {
			return
		}
		opts.response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTagsAll), &newTagsAll)...)
		if opts.response.Diagnostics.HasError() {
			return
		}

		if !newTagsAll.Equal(oldTagsAll) {
			// Some old resources may not have the required attribute set after Read:
			// https://github.com/hashicorp/terraform-provider-aws/issues/31180
			if identifier := r.GetIdentifierFramework(ctx, request.Plan); identifier != "" {
				if err := r.UpdateTags(ctx, sp, c, identifier, oldTagsAll, newTagsAll); err != nil {
					opts.response.Diagnostics.AddError(fmt.Sprintf("updating tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())

					return
				}
			}
			// TODO If the only change was to tags it would be nice to not call the resource's U handler.
		}
	}
}

func (r tagsResourceInterceptor) modifyPlan(ctx context.Context, opts interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]) {
	c := opts.c

	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		// If the entire plan is null, the resource is planned for destruction.
		if request.Plan.Raw.IsNull() {
			return
		}

		// Calculate the new value for the `tags_all` attribute.
		var planTags tftags.Map
		opts.response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)
		if opts.response.Diagnostics.HasError() {
			return
		}

		if planTags.IsWhollyKnown() {
			allTags := c.DefaultTagsConfig(ctx).MergeTags(tftags.New(ctx, planTags)).IgnoreConfig(c.IgnoreTagsConfig(ctx))
			opts.response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrTagsAll), fwflex.FlattenFrameworkStringValueMapLegacy(ctx, allTags.Map()))...)
		} else {
			opts.response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.Unknown)...)
		}

		if opts.response.Diagnostics.HasError() {
			return
		}
	}
}

func resourceTransparentTagging(servicePackageResourceTags unique.Handle[inttypes.ServicePackageResourceTags]) interface {
	resourceCRUDInterceptor
	resourceModifyPlanInterceptor
} {
	return &tagsResourceInterceptor{
		HTags: interceptors.HTags(servicePackageResourceTags),
	}
}

// resourceValidateRequiredTags validates that required tags are present for a given resource type.
func resourceValidateRequiredTags() resourceModifyPlanInterceptor {
	return &resourceValidateRequiredTagsInterceptor{}
}

type resourceValidateRequiredTagsInterceptor struct{}

func (r resourceValidateRequiredTagsInterceptor) modifyPlan(ctx context.Context, opts interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]) {
	c := opts.c

	_, _, _, typeName, _, ok := interceptors.InfoFromContext(ctx, c) //nolint:dogsled // legitimate use as-is, signature to be refactored
	if !ok {
		return
	}

	policy := c.TagPolicyConfig(ctx)
	if policy == nil {
		return
	}
	reqTags, ok := policy.RequiredTags[typeName]
	if !ok {
		return
	}

	switch request, _, when := opts.request, opts.response, opts.when; when {
	case Before:
		// If the entire plan is null, the resource is planned for destruction.
		if request.Plan.Raw.IsNull() {
			return
		}

		var planTags, stateTags tftags.Map
		opts.response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)
		opts.response.Diagnostics.Append(request.State.GetAttribute(ctx, path.Root(names.AttrTags), &stateTags)...)
		if opts.response.Diagnostics.HasError() {
			return
		}

		if !planTags.IsWhollyKnown() {
			return
		}

		allPlanTags := c.DefaultTagsConfig(ctx).MergeTags(tftags.New(ctx, planTags))
		allStateTags := c.DefaultTagsConfig(ctx).MergeTags(tftags.New(ctx, stateTags))

		isCreate := request.State.Raw.IsNull()
		hasTagsChange := !allPlanTags.Equal(allStateTags)

		if !isCreate && !hasTagsChange {
			return
		}

		if allPlanTags.ContainsAllKeys(reqTags) {
			return
		}

		missing := reqTags.Removed(allPlanTags).Keys()
		slices.Sort(missing)

		summary := "Missing Required Tags"
		detail := fmt.Sprintf("An organizational tag policy requires the following tags for %s: %s", typeName, missing)

		switch policy.Severity {
		case "warning":
			opts.response.Diagnostics.AddAttributeWarning(path.Root(names.AttrTags), summary, detail)
		default:
			opts.response.Diagnostics.AddAttributeError(path.Root(names.AttrTags), summary, detail)
		}
	}
}
