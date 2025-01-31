// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type tagsCRUDFunc func(context.Context, schemaResourceData, conns.ServicePackage, *types.ServicePackageResourceTags, string, string, any, diag.Diagnostics) (context.Context, diag.Diagnostics)

// tagsResourceInterceptor implements transparent tagging for resources.
type tagsResourceInterceptor struct {
	tagsInterceptor
	updateFunc tagsCRUDFunc
	readFunc   tagsCRUDFunc
}

func (r tagsResourceInterceptor) run(ctx context.Context, opts interceptorOptions) (context.Context, diag.Diagnostics) {
	c, diags := opts.c, opts.diags

	if r.tags == nil {
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

	switch d, when, why := opts.d, opts.when, opts.why; when {
	case Before:
		switch why {
		case Create, Update:
			// Merge the resource's configured tags with any provider configured default_tags.
			tags := c.DefaultTagsConfig(ctx).MergeTags(tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{})))
			// Remove system tags.
			tags = tags.IgnoreSystem(sp.ServicePackageName())

			tagsInContext.TagsIn = option.Some(tags)

			if why == Create {
				break
			}

			if d.GetRawPlan().GetAttr(names.AttrTagsAll).IsWhollyKnown() {
				if d.HasChange(names.AttrTagsAll) {
					// Some old resources may not have the required attribute set after Read:
					// https://github.com/hashicorp/terraform-provider-aws/issues/31180
					if identifier := r.getIdentifier(d); identifier != "" {
						o, n := d.GetChange(names.AttrTagsAll)

						if err := r.updateTags(ctx, sp, c, identifier, o, n); err != nil {
							return ctx, sdkdiag.AppendErrorf(diags, "updating tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
						}
					}
					// TODO If the only change was to tags it would be nice to not call the resource's U handler.
				}
			}
		}
	case After:
		// Set tags and tags_all in state after CRU.
		// C & U handlers are assumed to tail call the R handler.
		switch why {
		case Read:
			// Will occur on a refresh when the resource does not exist in AWS and needs to be recreated, e.g. "_disappears" tests.
			if d.Id() == "" {
				return ctx, diags
			}

			fallthrough
		case Create, Update:
			// If the R handler didn't set tags, try and read them from the service API.
			if tagsInContext.TagsOut.IsNone() {
				// Some old resources may not have the required attribute set after Read:
				// https://github.com/hashicorp/terraform-provider-aws/issues/31180
				if identifier := r.getIdentifier(d); identifier != "" {
					if err := r.listTags(ctx, sp, c, identifier); err != nil {
						return ctx, sdkdiag.AppendErrorf(diags, "listing tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
					}
				}
			}

			// Remove any provider configured ignore_tags and system tags from those returned from the service API.
			tags := tagsInContext.TagsOut.UnwrapOrDefault().IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx))

			// The resource's configured tags can now include duplicate tags that have been configured on the provider.
			if err := d.Set(names.AttrTags, tags.ResolveDuplicates(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx), d, names.AttrTags, nil).Map()); err != nil {
				return ctx, sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTags, err)
			}

			// Computed tags_all do.
			if err := d.Set(names.AttrTagsAll, tags.Map()); err != nil {
				return ctx, sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTagsAll, err)
			}
		}
	case Finally:
		switch why {
		case Update:
			if r.tags.IdentifierAttribute != "" && !d.GetRawPlan().GetAttr(names.AttrTagsAll).IsWhollyKnown() {
				ctx, diags = r.updateFunc(ctx, d, sp, r.tags, serviceName, resourceName, c, diags)
				ctx, diags = r.readFunc(ctx, d, sp, r.tags, serviceName, resourceName, c, diags)
			}
		}
	}

	return ctx, diags
}

// tagsResourceInterceptor implements transparent tagging for data sources.
type tagsDataSourceInterceptor struct {
	tagsInterceptor
}

func (r tagsDataSourceInterceptor) run(ctx context.Context, opts interceptorOptions) (context.Context, diag.Diagnostics) {
	c, diags := opts.c, opts.diags

	if r.tags == nil {
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

	switch d, when, why := opts.d, opts.when, opts.why; when {
	case Before:
		switch why {
		case Read:
			// Get the data source's configured tags.
			tags := tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{}))
			tagsInContext.TagsIn = option.Some(tags)
		}
	case After:
		// Set tags in state after R.
		switch why {
		case Read:
			// TODO: can this occur for a data source?
			if d.Id() == "" {
				return ctx, diags
			}

			// If the R handler didn't set tags, try and read them from the service API.
			if tagsInContext.TagsOut.IsNone() {
				// TODO: can this occur for a data source?
				// Some old resources may not have the required attribute set after Read:
				// https://github.com/hashicorp/terraform-provider-aws/issues/31180
				if identifier := r.getIdentifier(d); identifier != "" {
					if err := r.listTags(ctx, sp, c, identifier); err != nil {
						return ctx, sdkdiag.AppendErrorf(diags, "listing tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
					}
				}
			}

			// Remove any provider configured ignore_tags and system tags from those returned from the service API.
			tags := tagsInContext.TagsOut.UnwrapOrDefault().IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx))
			if err := d.Set(names.AttrTags, tags.Map()); err != nil {
				return ctx, sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTags, err)
			}
		}
	}

	return ctx, diags
}

func tagsUpdateFunc(ctx context.Context, d schemaResourceData, sp conns.ServicePackage, spt *types.ServicePackageResourceTags, serviceName, resourceName string, meta any, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	inContext, ok := conns.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	var identifier string
	if identifierAttribute := spt.IdentifierAttribute; identifierAttribute == "id" {
		identifier = d.Id()
	} else {
		identifier = d.Get(identifierAttribute).(string)
	}

	// Some old resources may not have the required attribute set after Read:
	// https://github.com/hashicorp/terraform-provider-aws/issues/31180
	if identifier == "" {
		return ctx, diags
	}

	configTags := make(map[string]string)
	if config := d.GetRawConfig(); !config.IsNull() && config.IsKnown() {
		c := config.GetAttr(names.AttrTags)
		if !c.IsNull() {
			for k, v := range c.AsValueMap() {
				if !v.IsNull() {
					configTags[k] = v.AsString()
				}
			}
		}
	}

	stateTags := make(map[string]string)
	if state := d.GetRawState(); !state.IsNull() && state.IsKnown() {
		s := state.GetAttr(names.AttrTagsAll)
		if !s.IsNull() {
			for k, v := range s.AsValueMap() {
				if !v.IsNull() {
					stateTags[k] = v.AsString()
				}
			}
		}
	}

	oldTags := tftags.New(ctx, stateTags)
	// if tags_all was computed because not wholly known
	// Merge the resource's configured tags with any provider configured default_tags.
	newTags := tagsInContext.DefaultConfig.MergeTags(tftags.New(ctx, configTags))
	// Remove system tags.
	newTags = newTags.IgnoreSystem(inContext.ServicePackageName)

	// If the service package has a generic resource update tags methods, call it.
	var err error

	if v, ok := sp.(tftags.ServiceTagUpdater); ok {
		err = v.UpdateTags(ctx, meta, identifier, oldTags, newTags)
	} else if v, ok := sp.(tftags.ResourceTypeTagUpdater); ok && spt.ResourceType != "" {
		err = v.UpdateTags(ctx, meta, identifier, spt.ResourceType, oldTags, newTags)
	} else {
		tflog.Warn(ctx, "No UpdateTags method found", map[string]interface{}{
			"ServicePackage": sp.ServicePackageName(),
			"ResourceType":   spt.ResourceType,
		})
	}

	// ISO partitions may not support tagging, giving error.
	if errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition(ctx), err) {
		return ctx, diags
	}

	if err != nil {
		return ctx, sdkdiag.AppendErrorf(diags, "updating tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
	}

	return ctx, diags
}

func tagsReadFunc(ctx context.Context, d schemaResourceData, sp conns.ServicePackage, spt *types.ServicePackageResourceTags, serviceName, resourceName string, meta any, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	inContext, ok := conns.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		return ctx, diags
	}
	var identifier string
	if identifierAttribute := spt.IdentifierAttribute; identifierAttribute == "id" {
		identifier = d.Id()
	} else {
		identifier = d.Get(identifierAttribute).(string)
	}

	// Some old resources may not have the required attribute set after Read:
	// https://github.com/hashicorp/terraform-provider-aws/issues/31180
	if identifier != "" {
		var err error

		if v, ok := sp.(tftags.ServiceTagLister); ok {
			err = v.ListTags(ctx, meta, identifier) // Sets tags in Context
		} else if v, ok := sp.(tftags.ResourceTypeTagLister); ok {
			if spt.ResourceType == "" {
				tflog.Error(ctx, "ListTags method requires ResourceType but none set", map[string]interface{}{
					"ServicePackage": sp.ServicePackageName(),
				})
			} else {
				err = v.ListTags(ctx, meta, identifier, spt.ResourceType) // Sets tags in Context
			}
		} else {
			tflog.Warn(ctx, "No ListTags method found", map[string]interface{}{
				"ServicePackage": sp.ServicePackageName(),
				"ResourceType":   spt.ResourceType,
			})
		}

		// ISO partitions may not support tagging, giving error.
		if errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition(ctx), err) {
			return ctx, diags
		}

		if inContext.ServicePackageName == names.DynamoDB && err != nil {
			// When a DynamoDB Table is `ARCHIVED`, ListTags returns `ResourceNotFoundException`.
			if tfresource.NotFound(err) || tfawserr.ErrMessageContains(err, "UnknownOperationException", "Tagging is not currently supported in DynamoDB Local.") {
				err = nil
			}
		}

		if err != nil {
			return ctx, sdkdiag.AppendErrorf(diags, "listing tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
		}
	}

	// Remove any provider configured ignore_tags and system tags from those returned from the service API.
	toAdd := tagsInContext.TagsOut.UnwrapOrDefault().IgnoreSystem(inContext.ServicePackageName).IgnoreConfig(tagsInContext.IgnoreConfig)

	// The resource's configured tags can now include duplicate tags that have been configured on the provider.
	if err := d.Set(names.AttrTags, toAdd.ResolveDuplicates(ctx, tagsInContext.DefaultConfig, tagsInContext.IgnoreConfig, d, names.AttrTags, nil).Map()); err != nil {
		return ctx, sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTags, err)
	}

	// Computed tags_all do.
	if err := d.Set(names.AttrTagsAll, toAdd.Map()); err != nil {
		return ctx, sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTagsAll, err)
	}

	return ctx, diags
}

type tagsInterceptor struct {
	tags *types.ServicePackageResourceTags
}

// getIdentifier returns the value of the identifier attribute used in AWS APIs.
func (r tagsInterceptor) getIdentifier(d schemaResourceData) string {
	var identifier string

	if identifierAttribute := r.tags.IdentifierAttribute; identifierAttribute != "" {
		if identifierAttribute == "id" {
			identifier = d.Id()
		} else {
			identifier = d.Get(identifierAttribute).(string)
		}
	}

	return identifier
}

// If the service package has a generic resource list tags methods, call it.
func (r tagsInterceptor) listTags(ctx context.Context, sp conns.ServicePackage, c *conns.AWSClient, identifier string) error {
	var err error

	if v, ok := sp.(tftags.ServiceTagLister); ok {
		err = v.ListTags(ctx, c, identifier) // Sets tags in Context
	} else if v, ok := sp.(tftags.ResourceTypeTagLister); ok {
		if r.tags.ResourceType == "" {
			tflog.Error(ctx, "ListTags method requires ResourceType but none set", map[string]interface{}{
				"ServicePackage": sp.ServicePackageName(),
			})
		} else {
			err = v.ListTags(ctx, c, identifier, r.tags.ResourceType) // Sets tags in Context
		}
	} else {
		tflog.Warn(ctx, "No ListTags method found", map[string]interface{}{
			"ServicePackage": sp.ServicePackageName(),
			"ResourceType":   r.tags.ResourceType,
		})
	}

	switch {
	// ISO partitions may not support tagging, giving error.
	case errs.IsUnsupportedOperationInPartitionError(c.Partition(ctx), err):
		err = nil
	case sp.ServicePackageName() == names.DynamoDB && err != nil:
		// When a DynamoDB Table is `ARCHIVED`, ListTags returns `ResourceNotFoundException`.
		if tfresource.NotFound(err) || tfawserr.ErrMessageContains(err, "UnknownOperationException", "Tagging is not currently supported in DynamoDB Local.") {
			err = nil
		}
	}

	return err
}

// If the service package has a generic resource update tags methods, call it.
func (r tagsInterceptor) updateTags(ctx context.Context, sp conns.ServicePackage, c *conns.AWSClient, identifier string, oldTags, newTags any) error {
	var err error

	if v, ok := sp.(tftags.ServiceTagUpdater); ok {
		err = v.UpdateTags(ctx, c, identifier, oldTags, newTags)
	} else if v, ok := sp.(tftags.ResourceTypeTagUpdater); ok {
		if r.tags.ResourceType == "" {
			tflog.Error(ctx, "UpdateTags method requires ResourceType but none set", map[string]interface{}{
				"ServicePackage": sp.ServicePackageName(),
			})
		} else {
			err = v.UpdateTags(ctx, c, identifier, r.tags.ResourceType, oldTags, newTags)
		}
	} else {
		tflog.Warn(ctx, "No UpdateTags method found", map[string]interface{}{
			"ServicePackage": sp.ServicePackageName(),
			"ResourceType":   r.tags.ResourceType,
		})
	}

	// ISO partitions may not support tagging, giving error.
	if errs.IsUnsupportedOperationInPartitionError(c.Partition(ctx), err) {
		err = nil
	}

	return err
}
