// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// contextFunc augments Context.
type contextFunc func(context.Context, *conns.AWSClient) context.Context

// wrappedDataSource represents an interceptor dispatcher for a Plugin Framework data source.
type wrappedDataSource struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	inner            datasource.DataSourceWithConfigure
	interceptors     dataSourceInterceptors
	meta             *conns.AWSClient
}

func newWrappedDataSource(bootstrapContext contextFunc, inner datasource.DataSourceWithConfigure, interceptors dataSourceInterceptors) datasource.DataSourceWithConfigure {
	return &wrappedDataSource{
		bootstrapContext: bootstrapContext,
		inner:            inner,
		interceptors:     interceptors,
	}
}

func (w *wrappedDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Metadata(ctx, request, response)
}

func (w *wrappedDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Schema(ctx, request, response)
}

func (w *wrappedDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	f := func(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) diag.Diagnostics {
		w.inner.Read(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	diags := interceptedDataSourceReadHandler(w.interceptors.read(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Configure(ctx, request, response)
}

func (w *wrappedDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	if v, ok := w.inner.(datasource.DataSourceWithConfigValidators); ok {
		ctx = w.bootstrapContext(ctx, w.meta)
		return v.ConfigValidators(ctx)
	}

	return nil
}

// wrappedResource represents an interceptor dispatcher for a Plugin Framework ephemeral resource.
type wrappedEphemeralResource struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	inner            ephemeral.EphemeralResourceWithConfigure
	meta             *conns.AWSClient
	interceptors     ephemeralResourceInterceptors
}

func newWrappedEphemeralResource(bootstrapContext contextFunc, inner ephemeral.EphemeralResourceWithConfigure, interceptors ephemeralResourceInterceptors) ephemeral.EphemeralResourceWithConfigure {
	return &wrappedEphemeralResource{
		bootstrapContext: bootstrapContext,
		inner:            inner,
		interceptors:     interceptors,
	}
}

func (w *wrappedEphemeralResource) Metadata(ctx context.Context, request ephemeral.MetadataRequest, response *ephemeral.MetadataResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Metadata(ctx, request, response)
}

func (w *wrappedEphemeralResource) Schema(ctx context.Context, request ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Schema(ctx, request, response)
}

func (w *wrappedEphemeralResource) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Open(ctx, request, response)
}

func (w *wrappedEphemeralResource) Configure(ctx context.Context, request ephemeral.ConfigureRequest, response *ephemeral.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Configure(ctx, request, response)
}

func (w *wrappedEphemeralResource) Renew(ctx context.Context, request ephemeral.RenewRequest, response *ephemeral.RenewResponse) {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithRenew); ok {
		ctx = w.bootstrapContext(ctx, w.meta)
		v.Renew(ctx, request, response)
	}
}

func (w *wrappedEphemeralResource) Close(ctx context.Context, request ephemeral.CloseRequest, response *ephemeral.CloseResponse) {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithClose); ok {
		ctx = w.bootstrapContext(ctx, w.meta)
		v.Close(ctx, request, response)
	}
}

func (w *wrappedEphemeralResource) ConfigValidators(ctx context.Context) []ephemeral.ConfigValidator {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithConfigValidators); ok {
		ctx = w.bootstrapContext(ctx, w.meta)
		return v.ConfigValidators(ctx)
	}

	return nil
}

func (w *wrappedEphemeralResource) ValidateConfig(ctx context.Context, request ephemeral.ValidateConfigRequest, response *ephemeral.ValidateConfigResponse) {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithValidateConfig); ok {
		ctx = w.bootstrapContext(ctx, w.meta)
		v.ValidateConfig(ctx, request, response)
	}
}

// wrappedResource represents an interceptor dispatcher for a Plugin Framework resource.
type wrappedResource struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	inner            resource.ResourceWithConfigure
	interceptors     resourceInterceptors
	meta             *conns.AWSClient
}

func newWrappedResource(bootstrapContext contextFunc, inner resource.ResourceWithConfigure, interceptors resourceInterceptors) resource.ResourceWithConfigure {
	return &wrappedResource{
		bootstrapContext: bootstrapContext,
		inner:            inner,
		interceptors:     interceptors,
	}
}

func (w *wrappedResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Metadata(ctx, request, response)
}

func (w *wrappedResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Schema(ctx, request, response)
}

func (w *wrappedResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	f := func(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) diag.Diagnostics {
		w.inner.Create(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	diags := interceptedResourceHandler(w.interceptors.create(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	f := func(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) diag.Diagnostics {
		w.inner.Read(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	diags := interceptedResourceHandler(w.interceptors.read(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	f := func(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) diag.Diagnostics {
		w.inner.Update(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	diags := interceptedResourceHandler(w.interceptors.update(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	f := func(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) diag.Diagnostics {
		w.inner.Delete(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	diags := interceptedResourceHandler(w.interceptors.delete(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Configure(ctx, request, response)
}

func (w *wrappedResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	if v, ok := w.inner.(resource.ResourceWithImportState); ok {
		ctx = w.bootstrapContext(ctx, w.meta)
		v.ImportState(ctx, request, response)

		return
	}

	response.Diagnostics.AddError(
		"Resource Import Not Implemented",
		"This resource does not support import. Please contact the provider developer for additional information.",
	)
}

func (w *wrappedResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if v, ok := w.inner.(resource.ResourceWithModifyPlan); ok {
		ctx = w.bootstrapContext(ctx, w.meta)
		v.ModifyPlan(ctx, request, response)
	}
}

func (w *wrappedResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if v, ok := w.inner.(resource.ResourceWithConfigValidators); ok {
		ctx = w.bootstrapContext(ctx, w.meta)
		return v.ConfigValidators(ctx)
	}

	return nil
}

func (w *wrappedResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	if v, ok := w.inner.(resource.ResourceWithValidateConfig); ok {
		ctx = w.bootstrapContext(ctx, w.meta)
		v.ValidateConfig(ctx, request, response)
	}
}

func (w *wrappedResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if v, ok := w.inner.(resource.ResourceWithUpgradeState); ok {
		ctx = w.bootstrapContext(ctx, w.meta)

		return v.UpgradeState(ctx)
	}

	return nil
}

func (w *wrappedResource) MoveState(ctx context.Context) []resource.StateMover {
	if v, ok := w.inner.(resource.ResourceWithMoveState); ok {
		ctx = w.bootstrapContext(ctx, w.meta)

		return v.MoveState(ctx)
	}

	return nil
}
