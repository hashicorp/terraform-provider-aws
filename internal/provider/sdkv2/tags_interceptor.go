// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"
	"fmt"
	"unique"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/interceptors"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// tagsResourceCRUDInterceptor implements transparent tagging on CRUD operations for resources.
type tagsResourceCRUDInterceptor struct {
	interceptors.HTags
}

func resourceTransparentTagging(servicePackageResourceTags unique.Handle[inttypes.ServicePackageResourceTags]) crudInterceptor {
	return &tagsResourceCRUDInterceptor{
		HTags: interceptors.HTags(servicePackageResourceTags),
	}
}

func (r tagsResourceCRUDInterceptor) run(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	if !r.Enabled() {
		return diags
	}

	sp, serviceName, resourceName, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return diags
	}

	switch d, when, why := opts.d, opts.when, opts.why; when {
	case Before:
		switch why {
		case Create, Update:
			// Merge the resource's configured tags with any provider configured default_tags.
			tags := c.DefaultTagsConfig(ctx).MergeTags(tftags.New(ctx, d.Get(names.AttrTags).(map[string]any)))
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
					if identifier := r.GetIdentifierSDKv2(ctx, d); identifier != "" {
						o, n := d.GetChange(names.AttrTagsAll)

						if err := r.UpdateTags(ctx, sp, c, identifier, o, n); err != nil {
							return sdkdiag.AppendErrorf(diags, "updating tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
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
				return diags
			}

			fallthrough
		case Create, Update:
			// If the R handler didn't set tags, try and read them from the service API.
			if tagsInContext.TagsOut.IsNone() {
				// Some old resources may not have the required attribute set after Read:
				// https://github.com/hashicorp/terraform-provider-aws/issues/31180
				if identifier := r.GetIdentifierSDKv2(ctx, d); identifier != "" {
					if err := r.ListTags(ctx, sp, c, identifier); err != nil {
						return sdkdiag.AppendErrorf(diags, "listing tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
					}
				}
			}

			// Remove any provider configured ignore_tags and system tags from those returned from the service API.
			tags := tagsInContext.TagsOut.UnwrapOrDefault().IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx))

			// The resource's configured tags can now include duplicate tags that have been configured on the provider.
			if err := d.Set(names.AttrTags, tags.ResolveDuplicates(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx), d, names.AttrTags, nil).Map()); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTags, err)
			}

			// Computed tags_all do.
			if err := d.Set(names.AttrTagsAll, tags.Map()); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTagsAll, err)
			}
		}
	case Finally:
		switch why {
		case Update:
			// Some old resources may not have the required attribute set after Read:
			// https://github.com/hashicorp/terraform-provider-aws/issues/31180
			if identifier := r.GetIdentifierSDKv2(ctx, d); identifier != "" && !d.GetRawPlan().GetAttr(names.AttrTagsAll).IsWhollyKnown() {
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
					return sdkdiag.AppendErrorf(diags, "updating tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
				}

				if err := r.ListTags(ctx, sp, c, identifier); err != nil {
					return sdkdiag.AppendErrorf(diags, "listing tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
				}

				// Remove any provider configured ignore_tags and system tags from those returned from the service API.
				toAdd := tagsInContext.TagsOut.UnwrapOrDefault().IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx))

				// The resource's configured tags can now include duplicate tags that have been configured on the provider.
				if err := d.Set(names.AttrTags, toAdd.ResolveDuplicates(ctx, c.DefaultTagsConfig(ctx), c.IgnoreTagsConfig(ctx), d, names.AttrTags, nil).Map()); err != nil {
					return sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTags, err)
				}

				// Computed tags_all do.
				if err := d.Set(names.AttrTagsAll, toAdd.Map()); err != nil {
					return sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTagsAll, err)
				}
			}
		}
	}

	return diags
}

// tagsDataSourceCRUDInterceptor implements transparent tagging on CRUD operations for data sources.
type tagsDataSourceCRUDInterceptor struct {
	interceptors.HTags
}

func dataSourceTransparentTagging(servicePackageResourceTags unique.Handle[inttypes.ServicePackageResourceTags]) crudInterceptor {
	return &tagsDataSourceCRUDInterceptor{
		HTags: interceptors.HTags(servicePackageResourceTags),
	}
}

func (r tagsDataSourceCRUDInterceptor) run(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
	c := opts.c
	var diags diag.Diagnostics

	if !r.Enabled() {
		return diags
	}

	sp, serviceName, resourceName, tagsInContext, ok := interceptors.InfoFromContext(ctx, c)
	if !ok {
		return diags
	}

	switch d, when, why := opts.d, opts.when, opts.why; when {
	case Before:
		switch why {
		case Read:
			// Get the data source's configured tags.
			tags := tftags.New(ctx, d.Get(names.AttrTags).(map[string]any))
			tagsInContext.TagsIn = option.Some(tags)
		}
	case After:
		// Set tags in state after R.
		switch why {
		case Read:
			// TODO: can this occur for a data source?
			if d.Id() == "" {
				return diags
			}

			// If the R handler didn't set tags, try and read them from the service API.
			if tagsInContext.TagsOut.IsNone() {
				// TODO: can this occur for a data source?
				// Some old resources may not have the required attribute set after Read:
				// https://github.com/hashicorp/terraform-provider-aws/issues/31180
				if identifier := r.GetIdentifierSDKv2(ctx, d); identifier != "" {
					if err := r.ListTags(ctx, sp, c, identifier); err != nil {
						return sdkdiag.AppendErrorf(diags, "listing tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
					}
				}
			}

			// Remove any provider configured ignore_tags and system tags from those returned from the service API.
			tags := tagsInContext.TagsOut.UnwrapOrDefault().IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(c.IgnoreTagsConfig(ctx))
			if err := d.Set(names.AttrTags, tags.Map()); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTags, err)
			}
		}
	}

	return diags
}

func setTagsAll() customizeDiffInterceptor {
	return interceptorFunc1[*schema.ResourceDiff, error](func(ctx context.Context, opts customizeDiffInterceptorOptions) error {
		c := opts.c

		switch d, when, why := opts.d, opts.when, opts.why; when {
		case Before:
			switch why {
			case CustomizeDiff:
				// Calculate the new value for the `tags_all` attribute.
				if !d.GetRawPlan().GetAttr(names.AttrTags).IsWhollyKnown() {
					if err := d.SetNewComputed(names.AttrTagsAll); err != nil {
						return fmt.Errorf("setting tags_all to Computed: %w", err)
					}
					return nil
				}

				newTags := tftags.New(ctx, d.Get(names.AttrTags).(map[string]any))
				allTags := c.DefaultTagsConfig(ctx).MergeTags(newTags).IgnoreConfig(c.IgnoreTagsConfig(ctx))
				if d.HasChange(names.AttrTags) {
					if newTags.HasZeroValue() {
						if err := d.SetNewComputed(names.AttrTagsAll); err != nil {
							return fmt.Errorf("setting tags_all to Computed: %w", err)
						}
					}

					if len(allTags) > 0 && (!newTags.HasZeroValue() || !allTags.HasZeroValue()) {
						if err := d.SetNew(names.AttrTagsAll, allTags.Map()); err != nil {
							return fmt.Errorf("setting new tags_all diff: %w", err)
						}
					}

					if len(allTags) == 0 {
						if err := d.SetNew(names.AttrTagsAll, allTags.Map()); err != nil {
							return fmt.Errorf("setting new tags_all diff: %w", err)
						}
					}
				} else {
					if len(allTags) > 0 && !allTags.HasZeroValue() {
						if err := d.SetNew(names.AttrTagsAll, allTags.Map()); err != nil {
							return fmt.Errorf("setting new tags_all diff: %w", err)
						}
						return nil
					}

					var newTagsAll tftags.KeyValueTags
					if v, ok := d.Get(names.AttrTagsAll).(map[string]any); ok {
						newTagsAll = tftags.New(ctx, v)
					}
					if len(allTags) > 0 && !newTagsAll.DeepEqual(allTags) && allTags.HasZeroValue() {
						if err := d.SetNewComputed(names.AttrTagsAll); err != nil {
							return fmt.Errorf("setting tags_all to Computed: %w", err)
						}
						return nil
					}
				}
			}
		}

		return nil
	})
}
