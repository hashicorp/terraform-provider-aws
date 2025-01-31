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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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

func (r tagsDataSourceInterceptor) read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse, c *conns.AWSClient, when when, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	if !r.HasServicePackageResourceTags() {
		return ctx, diags
	}

	inContext, ok := conns.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	sp := c.ServicePackage(ctx, inContext.ServicePackageName)
	if sp == nil {
		return ctx, diags
	}

	serviceName, err := names.HumanFriendly(sp.ServicePackageName())
	if err != nil {
		serviceName = "<service>"
	}

	resourceName := inContext.ResourceName
	if resourceName == "" {
		resourceName = "<thing>"
	}

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	switch when {
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
	tags *types.ServicePackageResourceTags
}

func (r tagsResourceInterceptor) create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse, meta *conns.AWSClient, when when, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	if r.tags == nil {
		return ctx, diags
	}

	inContext, ok := conns.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	switch when {
	case Before:
		var planTags tftags.Map
		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)

		if diags.HasError() {
			return ctx, diags
		}

		// Merge the resource's configured tags with any provider configured default_tags.
		tags := tagsInContext.DefaultConfig.MergeTags(tftags.New(ctx, planTags))
		// Remove system tags.
		tags = tags.IgnoreSystem(inContext.ServicePackageName)

		tagsInContext.TagsIn = option.Some(tags)
	case After:
		// Set values for unknowns.
		// Remove any provider configured ignore_tags and system tags from those passed to the service API.
		// Computed tags_all include any provider configured default_tags.
		stateTagsAll := flex.FlattenFrameworkStringValueMapLegacy(ctx, tagsInContext.TagsIn.MustUnwrap().IgnoreSystem(inContext.ServicePackageName).IgnoreConfig(tagsInContext.IgnoreConfig).Map())
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.NewMapFromMapValue(stateTagsAll))...)

		if diags.HasError() {
			return ctx, diags
		}
	}

	return ctx, diags
}

func (r tagsResourceInterceptor) read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse, meta *conns.AWSClient, when when, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	if r.tags == nil {
		return ctx, diags
	}

	inContext, ok := conns.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	sp := meta.ServicePackage(ctx, inContext.ServicePackageName)
	if sp == nil {
		return ctx, diags
	}

	serviceName, err := names.HumanFriendly(sp.ServicePackageName())
	if err != nil {
		serviceName = "<service>"
	}

	resourceName := inContext.ResourceName
	if resourceName == "" {
		resourceName = "<thing>"
	}

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	switch when {
	case After:
		// Will occur on a refresh when the resource does not exist in AWS and needs to be recreated, e.g. "_disappears" tests.
		if response.State.Raw.IsNull() {
			return ctx, diags
		}

		// If the R handler didn't set tags, try and read them from the service API.
		if tagsInContext.TagsOut.IsNone() {
			if identifierAttribute := r.tags.IdentifierAttribute; identifierAttribute != "" {
				var identifier string

				diags.Append(response.State.GetAttribute(ctx, path.Root(identifierAttribute), &identifier)...)

				if diags.HasError() {
					return ctx, diags
				}

				// Some old resources may not have the required attribute set after Read:
				// https://github.com/hashicorp/terraform-provider-aws/issues/31180
				if identifier != "" {
					// If the service package has a generic resource list tags methods, call it.
					var err error

					if v, ok := sp.(tftags.ServiceTagLister); ok {
						err = v.ListTags(ctx, meta, identifier) // Sets tags in Context
					} else if v, ok := sp.(tftags.ResourceTypeTagLister); ok {
						if r.tags.ResourceType == "" {
							tflog.Error(ctx, "ListTags method requires ResourceType but none set", map[string]interface{}{
								"ServicePackage": sp.ServicePackageName(),
							})
						} else {
							err = v.ListTags(ctx, meta, identifier, r.tags.ResourceType) // Sets tags in Context
						}
					} else {
						tflog.Warn(ctx, "No ListTags method found", map[string]interface{}{
							"ServicePackage": sp.ServicePackageName(),
							"ResourceType":   r.tags.ResourceType,
						})
					}

					// ISO partitions may not support tagging, giving error.
					if errs.IsUnsupportedOperationInPartitionError(meta.Partition(ctx), err) {
						return ctx, diags
					}

					if err != nil {
						diags.AddError(fmt.Sprintf("listing tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())

						return ctx, diags
					}
				}
			}
		}

		apiTags := tagsInContext.TagsOut.UnwrapOrDefault()

		// AWS APIs often return empty lists of tags when none have been configured.
		var stateTags tftags.Map
		response.State.GetAttribute(ctx, path.Root(names.AttrTags), &stateTags)
		// Remove any provider configured ignore_tags and system tags from those returned from the service API.
		// The resource's configured tags do not include any provider configured default_tags.
		if v := apiTags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(tagsInContext.IgnoreConfig).ResolveDuplicatesFramework(ctx, tagsInContext.DefaultConfig, tagsInContext.IgnoreConfig, response, &diags).Map(); len(v) > 0 {
			stateTags = tftags.NewMapFromMapValue(flex.FlattenFrameworkStringValueMapLegacy(ctx, v))
		}
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTags), &stateTags)...)

		if diags.HasError() {
			return ctx, diags
		}

		// Computed tags_all do.
		stateTagsAll := flex.FlattenFrameworkStringValueMapLegacy(ctx, apiTags.IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(tagsInContext.IgnoreConfig).Map())
		diags.Append(response.State.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.NewMapFromMapValue(stateTagsAll))...)

		if diags.HasError() {
			return ctx, diags
		}
	}

	return ctx, diags
}

func (r tagsResourceInterceptor) update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse, meta *conns.AWSClient, when when, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	if r.tags == nil {
		return ctx, diags
	}

	inContext, ok := conns.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	sp := meta.ServicePackage(ctx, inContext.ServicePackageName)
	if sp == nil {
		return ctx, diags
	}

	serviceName, err := names.HumanFriendly(sp.ServicePackageName())
	if err != nil {
		serviceName = "<service>"
	}

	resourceName := inContext.ResourceName
	if resourceName == "" {
		resourceName = "<thing>"
	}

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	switch when {
	case Before:
		var planTags tftags.Map
		diags.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)

		if diags.HasError() {
			return ctx, diags
		}

		// Merge the resource's configured tags with any provider configured default_tags.
		tags := tagsInContext.DefaultConfig.MergeTags(tftags.New(ctx, planTags))
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
			if identifierAttribute := r.tags.IdentifierAttribute; identifierAttribute != "" {
				var identifier string

				diags.Append(request.Plan.GetAttribute(ctx, path.Root(identifierAttribute), &identifier)...)

				if diags.HasError() {
					return ctx, diags
				}

				// Some old resources may not have the required attribute set after Read:
				// https://github.com/hashicorp/terraform-provider-aws/issues/31180
				if identifier != "" {
					// If the service package has a generic resource update tags methods, call it.
					var err error

					if v, ok := sp.(tftags.ServiceTagUpdater); ok {
						err = v.UpdateTags(ctx, meta, identifier, oldTagsAll, newTagsAll)
					} else if v, ok := sp.(tftags.ResourceTypeTagUpdater); ok && r.tags.ResourceType != "" {
						err = v.UpdateTags(ctx, meta, identifier, r.tags.ResourceType, oldTagsAll, newTagsAll)
					} else {
						tflog.Warn(ctx, "No UpdateTags method found", map[string]interface{}{
							"ServicePackage": sp.ServicePackageName(),
							"ResourceType":   r.tags.ResourceType,
						})
					}

					// ISO partitions may not support tagging, giving error.
					if errs.IsUnsupportedOperationInPartitionError(meta.Partition(ctx), err) {
						return ctx, diags
					}

					if err != nil {
						diags.AddError(fmt.Sprintf("updating tags for %s %s (%s)", serviceName, resourceName, identifier), err.Error())

						return ctx, diags
					}
				}
			}
			// TODO If the only change was to tags it would be nice to not call the resource's U handler.
		}
	}

	return ctx, diags
}

func (r tagsResourceInterceptor) delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse, meta *conns.AWSClient, when when, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	return ctx, diags
}

type tagsInterceptor struct {
	interceptors.WithTaggingMethods
}

// getIdentifier returns the value of the identifier attribute used in AWS APIs.
func (r tagsInterceptor) getIdentifier(ctx context.Context, d attributeGetter) string {
	var identifier string

	if identifierAttribute := r.ServicePackageResourceTags.IdentifierAttribute; identifierAttribute != "" {
		d.GetAttribute(ctx, path.Root(identifierAttribute), &identifier)
	}

	return identifier
}
