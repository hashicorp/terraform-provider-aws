// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Implemented by (schema.ResourceData|schema.ResourceDiff).GetOk().
type getAttributeFunc func(string) (any, bool)

// contextFunc augments Context.
type contextFunc func(context.Context, getAttributeFunc, any) (context.Context, error)

type wrappedDataSourceOptions struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	interceptors     interceptorInvocations
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
	return interceptedCRUDHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Read)
}

type wrappedResourceOptions struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	interceptors     interceptorInvocations
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

	r.CreateWithoutTimeout = w.create(r.CreateWithoutTimeout)
	r.ReadWithoutTimeout = w.read(r.ReadWithoutTimeout)
	r.UpdateWithoutTimeout = w.update(r.UpdateWithoutTimeout)
	r.DeleteWithoutTimeout = w.delete(r.DeleteWithoutTimeout)
	if v := r.Importer; v != nil {
		r.Importer.StateContext = w.import_(v.StateContext)
	}
	r.CustomizeDiff = w.customizeDiff(r.CustomizeDiff)
	for i, v := range r.StateUpgraders {
		r.StateUpgraders[i].Upgrade = w.stateUpgrade(v.Upgrade)
	}
}

func (w *wrappedResource) create(f schema.CreateContextFunc) schema.CreateContextFunc {
	return interceptedCRUDHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Create)
}

func (w *wrappedResource) read(f schema.ReadContextFunc) schema.ReadContextFunc {
	return interceptedCRUDHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Read)
}

func (w *wrappedResource) update(f schema.UpdateContextFunc) schema.UpdateContextFunc {
	return interceptedCRUDHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Update)
}

func (w *wrappedResource) delete(f schema.DeleteContextFunc) schema.DeleteContextFunc {
	return interceptedCRUDHandler(w.opts.bootstrapContext, w.opts.interceptors, f, Delete)
}

func (w *wrappedResource) import_(f schema.StateContextFunc) schema.StateContextFunc {
	return interceptedImportHandler(w.opts.bootstrapContext, w.opts.interceptors, f)
}

func (w *wrappedResource) customizeDiff(f schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
	return interceptedCustomizeDiffHandler(w.opts.bootstrapContext, w.opts.interceptors, f)
}

func (w *wrappedResource) stateUpgrade(f schema.StateUpgradeFunc) schema.StateUpgradeFunc {
	if f == nil {
		return nil
	}

	return func(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
		ctx, err := w.opts.bootstrapContext(ctx, func(key string) (any, bool) { v, ok := rawState[key]; return v, ok }, meta)
		if err != nil {
			return nil, err
		}

		return f(ctx, rawState, meta)
	}
}
