// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// contextFunc augments Context.
type contextFunc func(context.Context, *conns.AWSClient) context.Context

type wrappedDataSourceOptions struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	interceptors     dataSourceInterceptors
	typeName         string
}

// wrappedDataSource represents an interceptor dispatcher for a Plugin Framework data source.
type wrappedDataSource struct {
	inner datasource.DataSourceWithConfigure
	meta  *conns.AWSClient
	opts  wrappedDataSourceOptions
}

func newWrappedDataSource(inner datasource.DataSourceWithConfigure, opts wrappedDataSourceOptions) datasource.DataSourceWithConfigure {
	return &wrappedDataSource{
		inner: inner,
		opts:  opts,
	}
}

func (w *wrappedDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	// This method does not call down to the inner data source.
	response.TypeName = w.opts.typeName
}

func (w *wrappedDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	ctx = w.opts.bootstrapContext(ctx, w.meta)
	w.inner.Schema(ctx, request, response)
}

func (w *wrappedDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	f := func(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) diag.Diagnostics {
		w.inner.Read(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.opts.bootstrapContext(ctx, w.meta)
	diags := interceptedHandler(w.opts.interceptors.read(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}
	ctx = w.opts.bootstrapContext(ctx, w.meta)
	w.inner.Configure(ctx, request, response)
}

func (w *wrappedDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	if v, ok := w.inner.(datasource.DataSourceWithConfigValidators); ok {
		ctx = w.opts.bootstrapContext(ctx, w.meta)
		return v.ConfigValidators(ctx)
	}

	return nil
}

type wrappedEphemeralResourceOptions struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	interceptors     ephemeralResourceInterceptors
	typeName         string
}

// wrappedEphemeralResource represents an interceptor dispatcher for a Plugin Framework ephemeral resource.
type wrappedEphemeralResource struct {
	inner ephemeral.EphemeralResourceWithConfigure
	meta  *conns.AWSClient
	opts  wrappedEphemeralResourceOptions
}

func newWrappedEphemeralResource(inner ephemeral.EphemeralResourceWithConfigure, opts wrappedEphemeralResourceOptions) ephemeral.EphemeralResourceWithConfigure {
	return &wrappedEphemeralResource{
		inner: inner,
		opts:  opts,
	}
}

func (w *wrappedEphemeralResource) Metadata(ctx context.Context, request ephemeral.MetadataRequest, response *ephemeral.MetadataResponse) {
	// This method does not call down to the inner ephemeral resource.
	response.TypeName = w.opts.typeName
}

func (w *wrappedEphemeralResource) Schema(ctx context.Context, request ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	ctx = w.opts.bootstrapContext(ctx, w.meta)
	w.inner.Schema(ctx, request, response)
}

func (w *wrappedEphemeralResource) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	f := func(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) diag.Diagnostics {
		w.inner.Open(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.opts.bootstrapContext(ctx, w.meta)
	diags := interceptedHandler(w.opts.interceptors.open(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedEphemeralResource) Configure(ctx context.Context, request ephemeral.ConfigureRequest, response *ephemeral.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}
	ctx = w.opts.bootstrapContext(ctx, w.meta)
	w.inner.Configure(ctx, request, response)
}

func (w *wrappedEphemeralResource) Renew(ctx context.Context, request ephemeral.RenewRequest, response *ephemeral.RenewResponse) {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithRenew); ok {
		f := func(ctx context.Context, request ephemeral.RenewRequest, response *ephemeral.RenewResponse) diag.Diagnostics {
			v.Renew(ctx, request, response)
			return response.Diagnostics
		}
		ctx = w.opts.bootstrapContext(ctx, w.meta)
		diags := interceptedHandler(w.opts.interceptors.renew(), f, w.meta)(ctx, request, response)
		response.Diagnostics = diags
	}
}

func (w *wrappedEphemeralResource) Close(ctx context.Context, request ephemeral.CloseRequest, response *ephemeral.CloseResponse) {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithClose); ok {
		f := func(ctx context.Context, request ephemeral.CloseRequest, response *ephemeral.CloseResponse) diag.Diagnostics {
			v.Close(ctx, request, response)
			return response.Diagnostics
		}
		ctx = w.opts.bootstrapContext(ctx, w.meta)
		diags := interceptedHandler(w.opts.interceptors.close(), f, w.meta)(ctx, request, response)
		response.Diagnostics = diags
	}
}

func (w *wrappedEphemeralResource) ConfigValidators(ctx context.Context) []ephemeral.ConfigValidator {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithConfigValidators); ok {
		ctx = w.opts.bootstrapContext(ctx, w.meta)
		return v.ConfigValidators(ctx)
	}

	return nil
}

func (w *wrappedEphemeralResource) ValidateConfig(ctx context.Context, request ephemeral.ValidateConfigRequest, response *ephemeral.ValidateConfigResponse) {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithValidateConfig); ok {
		ctx = w.opts.bootstrapContext(ctx, w.meta)
		v.ValidateConfig(ctx, request, response)
	}
}

type wrappedResourceOptions struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext       contextFunc
	interceptors           resourceInterceptors
	typeName               string
	usesTransparentTagging bool
}

// wrappedResource represents an interceptor dispatcher for a Plugin Framework resource.
type wrappedResource struct {
	inner resource.ResourceWithConfigure
	meta  *conns.AWSClient
	opts  wrappedResourceOptions
}

func newWrappedResource(inner resource.ResourceWithConfigure, opts wrappedResourceOptions) resource.ResourceWithConfigure {
	return &wrappedResource{
		inner: inner,
		opts:  opts,
	}
}

func (w *wrappedResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	// This method does not call down to the inner resource.
	response.TypeName = w.opts.typeName
}

func (w *wrappedResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	ctx = w.opts.bootstrapContext(ctx, w.meta)
	w.inner.Schema(ctx, request, response)
}

func (w *wrappedResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	f := func(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) diag.Diagnostics {
		w.inner.Create(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.opts.bootstrapContext(ctx, w.meta)
	diags := interceptedHandler(w.opts.interceptors.create(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	f := func(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) diag.Diagnostics {
		w.inner.Read(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.opts.bootstrapContext(ctx, w.meta)
	diags := interceptedHandler(w.opts.interceptors.read(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	f := func(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) diag.Diagnostics {
		w.inner.Update(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.opts.bootstrapContext(ctx, w.meta)
	diags := interceptedHandler(w.opts.interceptors.update(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	f := func(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) diag.Diagnostics {
		w.inner.Delete(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.opts.bootstrapContext(ctx, w.meta)
	diags := interceptedHandler(w.opts.interceptors.delete(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}
	ctx = w.opts.bootstrapContext(ctx, w.meta)
	w.inner.Configure(ctx, request, response)
}

func (w *wrappedResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	if v, ok := w.inner.(resource.ResourceWithImportState); ok {
		ctx = w.opts.bootstrapContext(ctx, w.meta)
		v.ImportState(ctx, request, response)

		return
	}

	response.Diagnostics.AddError(
		"Resource Import Not Implemented",
		"This resource does not support import. Please contact the provider developer for additional information.",
	)
}

func (w *wrappedResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	ctx = w.opts.bootstrapContext(ctx, w.meta)

	if w.opts.usesTransparentTagging {
		w.setTagsAll(ctx, request, response)
		if response.Diagnostics.HasError() {
			return
		}
	}

	if v, ok := w.inner.(resource.ResourceWithModifyPlan); ok {
		v.ModifyPlan(ctx, request, response)
	}
}

func (w *wrappedResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if v, ok := w.inner.(resource.ResourceWithConfigValidators); ok {
		ctx = w.opts.bootstrapContext(ctx, w.meta)
		return v.ConfigValidators(ctx)
	}

	return nil
}

func (w *wrappedResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	if v, ok := w.inner.(resource.ResourceWithValidateConfig); ok {
		ctx = w.opts.bootstrapContext(ctx, w.meta)
		v.ValidateConfig(ctx, request, response)
	}
}

func (w *wrappedResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if v, ok := w.inner.(resource.ResourceWithUpgradeState); ok {
		ctx = w.opts.bootstrapContext(ctx, w.meta)

		return v.UpgradeState(ctx)
	}

	return nil
}

func (w *wrappedResource) MoveState(ctx context.Context) []resource.StateMover {
	if v, ok := w.inner.(resource.ResourceWithMoveState); ok {
		ctx = w.opts.bootstrapContext(ctx, w.meta)

		return v.MoveState(ctx)
	}

	return nil
}

// setTagsAll is a plan modifier that calculates the new value for the `tags_all` attribute.
func (w *wrappedResource) setTagsAll(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	// If the entire plan is null, the resource is planned for destruction.
	if request.Plan.Raw.IsNull() {
		return
	}

	var planTags tftags.Map
	response.Diagnostics.Append(request.Plan.GetAttribute(ctx, path.Root(names.AttrTags), &planTags)...)
	if response.Diagnostics.HasError() {
		return
	}

	if planTags.IsWhollyKnown() {
		allTags := w.meta.DefaultTagsConfig(ctx).MergeTags(tftags.New(ctx, planTags)).IgnoreConfig(w.meta.IgnoreTagsConfig(ctx))
		response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrTagsAll), fwflex.FlattenFrameworkStringValueMapLegacy(ctx, allTags.Map()))...)
	} else {
		response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root(names.AttrTagsAll), tftags.Unknown)...)
	}
}
