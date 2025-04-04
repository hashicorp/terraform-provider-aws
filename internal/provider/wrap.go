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
	crudInterceptors crudInterceptorItems
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
	return interceptedCRUDHandler(w.opts.bootstrapContext, w.opts.crudInterceptors, f, Read)
}

type wrappedResourceOptions struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext          contextFunc
	crudInterceptors          crudInterceptorItems
	customizeDiffInterceptors customizeDiffInterceptorItems
	customizeDiffFuncs        []schema.CustomizeDiffFunc
	// importsFuncs are called before bootstrapContext.
	importFuncs []schema.StateContextFunc
	typeName    string
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
	if f == nil {
		return nil
	}

	return interceptedCRUDHandler(w.opts.bootstrapContext, w.opts.crudInterceptors, f, Create)
}

func (w *wrappedResource) read(f schema.ReadContextFunc) schema.ReadContextFunc {
	if f == nil {
		return nil
	}

	return interceptedCRUDHandler(w.opts.bootstrapContext, w.opts.crudInterceptors, f, Read)
}

func (w *wrappedResource) update(f schema.UpdateContextFunc) schema.UpdateContextFunc {
	if f == nil {
		return nil
	}

	return interceptedCRUDHandler(w.opts.bootstrapContext, w.opts.crudInterceptors, f, Update)
}

func (w *wrappedResource) delete(f schema.DeleteContextFunc) schema.DeleteContextFunc {
	if f == nil {
		return nil
	}

	return interceptedCRUDHandler(w.opts.bootstrapContext, w.opts.crudInterceptors, f, Delete)
}

func (w *wrappedResource) import_(f schema.StateContextFunc) schema.StateContextFunc {
	if f == nil {
		return nil
	}

	return func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
		for _, f := range w.opts.importFuncs {
			if _, err := f(ctx, d, meta); err != nil {
				return nil, err
			}
		}

		ctx, diags := w.opts.bootstrapContext(ctx, d.GetOk, meta)
		if diags.HasError() {
			return nil, sdkdiag.DiagnosticsError(diags)
		}

		return f(ctx, d, meta)
	}
}

func (w *wrappedResource) customizeDiff(f schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
	// if f == nil {
	// 	f = func(context.Context, *schema.ResourceDiff, any) error {
	// 		return nil
	// 	}
	// }

	// return interceptedCustomizeDiffHandler(w.opts.bootstrapContext, w.opts.customizeDiffInterceptors, f)

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
