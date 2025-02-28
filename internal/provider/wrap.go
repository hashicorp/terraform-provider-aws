// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// contextFunc augments Context.
type contextFunc func(context.Context, any) context.Context

type wrappedDataSourceOptions struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	interceptors     interceptorItems
	typeName         string
}

// wrappedDataSource represents an interceptor dispatcher for a Plugin SDK v2 data source.
type wrappedDataSource struct {
	opts wrappedDataSourceOptions
}

func wrapDataSource(r *schema.Resource, opts wrappedDataSourceOptions) {
	w := &wrappedDataSource{
		opts: opts,
	}

	if v := r.ReadWithoutTimeout; v != nil {
		r.ReadWithoutTimeout = w.read(v)
	}
}

func (w *wrappedDataSource) read(f schema.ReadContextFunc) schema.ReadContextFunc {
	return interceptedHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Read)
}

type wrappedResourceOptions struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext       contextFunc
	interceptors           interceptorItems
	typeName               string
	usesTransparentTagging bool
}

// wrappedResource represents an interceptor dispatcher for a Plugin SDK v2 resource.
type wrappedResource struct {
	opts wrappedResourceOptions
}

func wrapResource(r *schema.Resource, opts wrappedResourceOptions) {
	w := &wrappedResource{
		opts: opts,
	}

	r.CreateWithoutTimeout = w.create(r.CreateWithoutTimeout)
	r.ReadWithoutTimeout = w.read(r.ReadWithoutTimeout)
	r.UpdateWithoutTimeout = w.update(r.UpdateWithoutTimeout)
	r.DeleteWithoutTimeout = w.delete(r.DeleteWithoutTimeout)
	if v := r.Importer; v != nil {
		r.Importer.StateContext = w.state(v.StateContext)
	}
	r.CustomizeDiff = w.customizeDiff(r.CustomizeDiff)
	for i, v := range r.StateUpgraders {
		r.StateUpgraders[i].Upgrade = w.stateUpgrade(v.Upgrade)
	}
}

func (w *wrappedResource) create(f schema.CreateContextFunc) schema.CreateContextFunc {
	if f == nil {
		return nil
	}

	return interceptedHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Create)
}

func (w *wrappedResource) read(f schema.ReadContextFunc) schema.ReadContextFunc {
	if f == nil {
		return nil
	}

	return interceptedHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Read)
}

func (w *wrappedResource) update(f schema.UpdateContextFunc) schema.UpdateContextFunc {
	if f == nil {
		return nil
	}

	return interceptedHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Update)
}

func (w *wrappedResource) delete(f schema.DeleteContextFunc) schema.DeleteContextFunc {
	if f == nil {
		return nil
	}

	return interceptedHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Delete)
}

func (w *wrappedResource) state(f schema.StateContextFunc) schema.StateContextFunc {
	if f == nil {
		return nil
	}

	return func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
		ctx = w.opts.bootstrapContext(ctx, meta)
		return f(ctx, d, meta)
	}
}

func (w *wrappedResource) customizeDiff(f schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
	if w.opts.usesTransparentTagging {
		if f == nil {
			return w.customizeDiffWithBootstrappedContext(setTagsAll)
		} else {
			return w.customizeDiffWithBootstrappedContext(customdiff.Sequence(setTagsAll, f))
		}
	}

	if f == nil {
		return nil
	}

	return w.customizeDiffWithBootstrappedContext(f)
}

func (w *wrappedResource) customizeDiffWithBootstrappedContext(f schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
		ctx = w.opts.bootstrapContext(ctx, meta)
		return f(ctx, d, meta)
	}
}

func (w *wrappedResource) stateUpgrade(f schema.StateUpgradeFunc) schema.StateUpgradeFunc {
	if f == nil {
		return nil
	}

	return func(ctx context.Context, rawState map[string]interface{}, meta any) (map[string]interface{}, error) {
		ctx = w.opts.bootstrapContext(ctx, meta)
		return f(ctx, rawState, meta)
	}
}

// setTagsAll is a CustomizeDiff function that calculates the new value for the `tags_all` attribute.
func setTagsAll(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	c := meta.(*conns.AWSClient)

	if !d.GetRawPlan().GetAttr(names.AttrTags).IsWhollyKnown() {
		if err := d.SetNewComputed(names.AttrTagsAll); err != nil {
			return fmt.Errorf("setting tags_all to Computed: %w", err)
		}
		return nil
	}

	newTags := tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{}))
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
		if v, ok := d.Get(names.AttrTagsAll).(map[string]interface{}); ok {
			newTagsAll = tftags.New(ctx, v)
		}
		if len(allTags) > 0 && !newTagsAll.DeepEqual(allTags) && allTags.HasZeroValue() {
			if err := d.SetNewComputed(names.AttrTagsAll); err != nil {
				return fmt.Errorf("setting tags_all to Computed: %w", err)
			}
			return nil
		}
	}

	return nil
}
