// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/interceptors"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// tagsResourceInterceptor implements transparent tagging for resources.
type tagsResourceInterceptor struct {
	tagsInterceptor
}

func newTagsResourceInterceptor(servicePackageResourceTags *types.ServicePackageResourceTags) interceptor {
	return &tagsResourceInterceptor{
		tagsInterceptor: tagsInterceptor{
			WithTaggingMethods: interceptors.WithTaggingMethods{
				ServicePackageResourceTags: servicePackageResourceTags,
			},
		},
	}
}

func (r tagsResourceInterceptor) run(ctx context.Context, opts interceptorOptions) (context.Context, diag.Diagnostics) {
	c, diags := opts.c, opts.diags

	if !r.HasServicePackageResourceTags() {
		return ctx, diags
	}

	sp, serviceName, resourceName, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
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

						if err := r.UpdateTags(ctx, sp, c, identifier, o, n); err != nil {
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
					if err := r.ListTags(ctx, sp, c, identifier); err != nil {
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
			// Some old resources may not have the required attribute set after Read:
			// https://github.com/hashicorp/terraform-provider-aws/issues/31180
			if identifier := r.getIdentifier(d); identifier != "" && !d.GetRawPlan().GetAttr(names.AttrTagsAll).IsWhollyKnown() {
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
				newTags := c.DefaultTagsConfig(ctx).MergeTags(tftags.New(ctx, configTags))
				// Remove system tags.
				newTags = newTags.IgnoreSystem(sp.ServicePackageName())

				if err := r.UpdateTags(ctx, sp, c, identifier, oldTags, newTags); err != nil {
					return ctx, sdkdiag.AppendErrorf(diags, "updating tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
				}

				if err := r.ListTags(ctx, sp, c, identifier); err != nil {
					return ctx, sdkdiag.AppendErrorf(diags, "listing tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
				}

				// Remove any provider configured ignore_tags and system tags from those returned from the service API.
				toAdd := tagsInContext.TagsOut.UnwrapOrDefault().IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx))

				// The resource's configured tags can now include duplicate tags that have been configured on the provider.
				if err := d.Set(names.AttrTags, toAdd.ResolveDuplicates(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx), d, names.AttrTags, nil).Map()); err != nil {
					return ctx, sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTags, err)
				}

				// Computed tags_all do.
				if err := d.Set(names.AttrTagsAll, toAdd.Map()); err != nil {
					return ctx, sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTagsAll, err)
				}
			}
		}
	}

	return ctx, diags
}

// tagsResourceInterceptor implements transparent tagging for data sources.
type tagsDataSourceInterceptor struct {
	tagsInterceptor
}

func newTagsDataSourceInterceptor(servicePackageResourceTags *types.ServicePackageResourceTags) interceptor {
	return &tagsDataSourceInterceptor{
		tagsInterceptor: tagsInterceptor{
			WithTaggingMethods: interceptors.WithTaggingMethods{
				ServicePackageResourceTags: servicePackageResourceTags,
			},
		},
	}
}

func (r tagsDataSourceInterceptor) run(ctx context.Context, opts interceptorOptions) (context.Context, diag.Diagnostics) {
	c, diags := opts.c, opts.diags

	if !r.HasServicePackageResourceTags() {
		return ctx, diags
	}

	sp, serviceName, resourceName, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
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
					if err := r.ListTags(ctx, sp, c, identifier); err != nil {
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

type tagsInterceptor struct {
	interceptors.WithTaggingMethods
}

// getIdentifier returns the value of the identifier attribute used in AWS APIs.
func (r tagsInterceptor) getIdentifier(d schemaResourceData) string {
	var identifier string

	if identifierAttribute := r.ServicePackageResourceTags.IdentifierAttribute; identifierAttribute != "" {
		if identifierAttribute == "id" {
			identifier = d.Id()
		} else {
			identifier = d.Get(identifierAttribute).(string)
		}
	}

	return identifier
}
