// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwprovider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/interceptors"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// tagsDataSourceInterceptor implements transparent tagging for data sources.
type tagsDataSourceInterceptor struct {
	tagsInterceptor
}

func newTagsDataSourceInterceptor(servicePackageResourceTags *types.ServicePackageResourceTags) dataSourceInterceptor {
	return &tagsDataSourceInterceptor{
		tagsInterceptor: tagsInterceptor{
			WithTaggingMethods: interceptors.WithTaggingMethods{
				ServicePackageResourceTags: servicePackageResourceTags,
			},
		},
	}
}

func (r tagsDataSourceInterceptor) read(ctx context.Context, opts interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]) (context.Context, diag.Diagnostics) {
	c, diags := opts.c, opts.diags

	if !r.HasServicePackageResourceTags() {
		return ctx, diags
	}

	sp, serviceName, resourceName, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return ctx, diags
	}

	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		var configTags tftags.Map
		diags.Append(request.Config.GetAttribute(ctx, path.Root(names.AttrTags), &configTags)...)
		if diags.HasError() {
			return ctx, diags
		}

		tags := tftags.New(ctx, configTags)
		tagsInContext.TagsIn = option.Some(tags)
	case After:
		// If the R handler didn't set tags, try and read them from the service API.
		if tagsInContext.TagsOut.IsNone() {
			if identifier := r.getIdentifier(ctx, response.State); identifier != "" {
				if err := r.ListTags(ctx, sp, c, identifier); err != nil {
					diags.AddError(fmt.Sprintf("listing tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())
					return ctx, diags
				}
			}
		}

		tags := tagsInContext.TagsOut.UnwrapOrDefault()
		// Remove any provider configured ignore_tags and system tags from those returned from the service API.
		stateTags := flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx)).Map())
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTags), tftags.NewMapFromMapValue(stateTags))...)
		if diags.HasError() {
			return ctx, diags
		}
	}

	return ctx, diags
}

// tagsResourceInterceptor implements transparent tagging for resources.
type tagsResourceInterceptor struct {
	tagsInterceptor
}

func newTagsResourceInterceptor(servicePackageResourceTags *types.ServicePackageResourceTags) resourceInterceptor {
	return &tagsResourceInterceptor{
		tagsInterceptor: tagsInterceptor{
			WithTaggingMethods: interceptors.WithTaggingMethods{
				ServicePackageResourceTags: servicePackageResourceTags,
			},
		},
	}
}

func (r tagsResourceInterceptor) create(ctx context.Context, opts interceptorOptions[resource.CreateRequest, resource.CreateResponse]) (context.Context, diag.Diagnostics) {
	c, diags := opts.c, opts.diags

	if !r.HasServicePackageResourceTags() {
		return ctx, diags
	}

	sp, _, _, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return ctx, diags
	}

	switch request, response, when := opts.request, opts.response, opts.when; when {
	case Before:
		var planTags tftags.Map
		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)
		if diags.HasError() {
			return ctx, diags
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
		stateTagsAll := flex.FlattenFrameworkStringValueMapLegacy(ctx, tagsInContext.TagsIn.MustUnwrap().IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx)).Map())
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.NewMapFromMapValue(stateTagsAll))...)
		if diags.HasError() {
			return ctx, diags
		}
	}

	return ctx, diags
}

func (r tagsResourceInterceptor) read(ctx context.Context, opts interceptorOptions[resource.ReadRequest, resource.ReadResponse]) (context.Context, diag.Diagnostics) {
	c, diags := opts.c, opts.diags

	if !r.HasServicePackageResourceTags() {
		return ctx, diags
	}

	sp, serviceName, resourceName, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return ctx, diags
	}

	switch response, when := opts.response, opts.when; when {
	case After:
		// Will occur on a refresh when the resource does not exist in AWS and needs to be recreated, e.g. "_disappears" tests.
		if response.State.Raw.IsNull() {
			return ctx, diags
		}

		// If the R handler didn't set tags, try and read them from the service API.
		if tagsInContext.TagsOut.IsNone() {
			// Some old resources may not have the required attribute set after Read:
			// https://github.com/hashicorp/terraform-provider-aws/issues/31180
			if identifier := r.getIdentifier(ctx, response.State); identifier != "" {
				if err := r.ListTags(ctx, sp, c, identifier); err != nil {
					diags.AddError(fmt.Sprintf("listing tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())

					return ctx, diags
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
			stateTags = tftags.NewMapFromMapValue(flex.FlattenFrameworkStringValueMapLegacy(ctx, v))
		}
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTags), &stateTags)...)
		if diags.HasError() {
			return ctx, diags
		}

		// Computed tags_all do.
		stateTagsAll := flex.FlattenFrameworkStringValueMapLegacy(ctx, apiTags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx)).Map())
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.NewMapFromMapValue(stateTagsAll))...)
		if diags.HasError() {
			return ctx, diags
		}
	}

	return ctx, diags
}

func (r tagsResourceInterceptor) update(ctx context.Context, opts interceptorOptions[resource.UpdateRequest, resource.UpdateResponse]) (context.Context, diag.Diagnostics) {
	c, diags := opts.c, opts.diags

	if !r.HasServicePackageResourceTags() {
		return ctx, diags
	}

	sp, serviceName, resourceName, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return ctx, diags
	}

	switch request, when := opts.request, opts.when; when {
	case Before:
		var planTags tftags.Map
		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)
		if diags.HasError() {
			return ctx, diags
		}

		// Merge the resource's configured tags with any provider configured default_tags.
		tags := c.DefaultTagsConfig(ctx).MergeTags(tftags.New(ctx, planTags))
		// Remove system tags.
		tags = tags.IgnoreSystem(sp.ServicePackageName())
		tagsInContext.TagsIn = option.Some(tags)

		var oldTagsAll, newTagsAll tftags.Map
		diags.Append(request.State.GetAttribute(ctx, path.Root(names.AttrTagsAll), &oldTagsAll)...)
		if diags.HasError() {
			return ctx, diags
		}
		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTagsAll), &newTagsAll)...)
		if diags.HasError() {
			return ctx, diags
		}

		if !newTagsAll.Equal(oldTagsAll) {
			// Some old resources may not have the required attribute set after Read:
			// https://github.com/hashicorp/terraform-provider-aws/issues/31180
			if identifier := r.getIdentifier(ctx, request.Plan); identifier != "" {
				if err := r.UpdateTags(ctx, sp, c, identifier, oldTagsAll, newTagsAll); err != nil {
					diags.AddError(fmt.Sprintf("updating tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())

					return ctx, diags
				}
			}
			// TODO If the only change was to tags it would be nice to not call the resource's U handler.
		}
	}

	return ctx, diags
}

func (r tagsResourceInterceptor) delete(ctx context.Context, opts interceptorOptions[resource.DeleteRequest, resource.DeleteResponse]) (context.Context, diag.Diagnostics) {
	return ctx, opts.diags
}

type tagsInterceptor struct {
	interceptors.WithTaggingMethods
}

// getIdentifier returns the value of the identifier attribute used in AWS APIs.
func (r tagsInterceptor) getIdentifier(ctx context.Context, d interface {
	GetAttribute(context.Context, path.Path, any) diag.Diagnostics
}) string {
	var identifier string

	if identifierAttribute := r.ServicePackageResourceTags.IdentifierAttribute; identifierAttribute != "" {
		d.GetAttribute(ctx, path.Root(identifierAttribute), &identifier)
	}

	return identifier
}
