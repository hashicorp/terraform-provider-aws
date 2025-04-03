// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// Implemented by (schema.ResourceData|schema.ResourceDiff).GetOk().
type getAttributeFunc func(string) (any, bool)

// contextFunc augments Context.
type contextFunc func(context.Context, getAttributeFunc, any) (context.Context, diag.Diagnostics)

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
	bootstrapContext   contextFunc
	customizeDiffFuncs []schema.CustomizeDiffFunc
	interceptors       interceptorItems
	typeName           string
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
		ctx, diags := w.opts.bootstrapContext(ctx, d.GetOk, meta)
		if diags.HasError() {
			return nil, sdkdiag.DiagnosticsError(diags)
		}

		return f(ctx, d, meta)
	}
}

func (w *wrappedResource) customizeDiff(f schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
	if len(w.opts.customizeDiffFuncs) > 0 {
		customizeDiffFuncs := slices.Clone(w.opts.customizeDiffFuncs)
		if f != nil {
			customizeDiffFuncs = append(customizeDiffFuncs, f)
		}
		return w.customizeDiffWithBootstrappedContext(customdiff.Sequence(customizeDiffFuncs...))
	}

	if f == nil {
		return nil
	}

	return w.customizeDiffWithBootstrappedContext(f)
}

func (w *wrappedResource) customizeDiffWithBootstrappedContext(f schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
		ctx, diags := w.opts.bootstrapContext(ctx, d.GetOk, meta)
		if diags.HasError() {
			return sdkdiag.DiagnosticsError(diags)
		}

		return f(ctx, d, meta)
	}
}

func (w *wrappedResource) stateUpgrade(f schema.StateUpgradeFunc) schema.StateUpgradeFunc {
	if f == nil {
		return nil
	}

	return func(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
		ctx, diags := w.opts.bootstrapContext(ctx, func(key string) (any, bool) { v, ok := rawState[key]; return v, ok }, meta)
		if diags.HasError() {
			return nil, sdkdiag.DiagnosticsError(diags)
		}

		return f(ctx, rawState, meta)
	}
}
