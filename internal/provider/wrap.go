// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	bootstrapContext contextFunc
	interceptors     interceptorItems
	typeName         string
}

// wrappedResource represents an interceptor dispatcher for a Plugin SDK v2 resource.
type wrappedResource struct {
	opts wrappedResourceOptions
}

func wrapResource(r *schema.Resource, opts wrappedResourceOptions) {
	w := &wrappedResource{
		opts: opts,
	}

	if v := r.CreateWithoutTimeout; v != nil {
		r.CreateWithoutTimeout = w.create(v)
	}
	if v := r.ReadWithoutTimeout; v != nil {
		r.ReadWithoutTimeout = w.read(v)
	}
	if v := r.UpdateWithoutTimeout; v != nil {
		r.UpdateWithoutTimeout = w.update(v)
	}
	if v := r.DeleteWithoutTimeout; v != nil {
		r.DeleteWithoutTimeout = w.delete(v)
	}
	if v := r.Importer; v != nil {
		if v := v.StateContext; v != nil {
			r.Importer.StateContext = w.state(v)
		}
	}
	if v := r.CustomizeDiff; v != nil {
		r.CustomizeDiff = w.customizeDiff(v)
	}
	for i, v := range r.StateUpgraders {
		if v := v.Upgrade; v != nil {
			r.StateUpgraders[i].Upgrade = w.stateUpgrade(v)
		}
	}
}

func (w *wrappedResource) create(f schema.CreateContextFunc) schema.CreateContextFunc {
	return interceptedHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Create)
}

func (w *wrappedResource) read(f schema.ReadContextFunc) schema.ReadContextFunc {
	return interceptedHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Read)
}

func (w *wrappedResource) update(f schema.UpdateContextFunc) schema.UpdateContextFunc {
	return interceptedHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Update)
}

func (w *wrappedResource) delete(f schema.DeleteContextFunc) schema.DeleteContextFunc {
	return interceptedHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Delete)
}

func (w *wrappedResource) state(f schema.StateContextFunc) schema.StateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
		ctx = w.opts.bootstrapContext(ctx, meta)
		return f(ctx, d, meta)
	}
}

func (w *wrappedResource) customizeDiff(f schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
		ctx = w.opts.bootstrapContext(ctx, meta)
		return f(ctx, d, meta)
	}
}

func (w *wrappedResource) stateUpgrade(f schema.StateUpgradeFunc) schema.StateUpgradeFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta any) (map[string]interface{}, error) {
		ctx = w.opts.bootstrapContext(ctx, meta)
		return f(ctx, rawState, meta)
	}
}
