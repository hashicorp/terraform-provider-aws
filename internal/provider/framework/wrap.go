// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
)

// Implemented by (Config|Plan|State).GetAttribute().
type getAttributeFunc func(context.Context, path.Path, any) diag.Diagnostics

// contextFunc augments Context.
type contextFunc func(context.Context, getAttributeFunc, *conns.AWSClient) (context.Context, diag.Diagnostics)

type wrappedDataSourceOptions struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	interceptors     interceptorInvocations
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
	ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	f := func(ctx context.Context, request *datasource.SchemaRequest, response *datasource.SchemaResponse) diag.Diagnostics {
		w.inner.Schema(ctx, *request, response)
		return response.Diagnostics
	}
	response.Diagnostics.Append(interceptedHandler(w.opts.interceptors.dataSourceSchema(), f, w.meta)(ctx, &request, response)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Validate the data source's model against the schema.
	if v, ok := w.inner.(framework.DataSourceValidateModel); ok {
		response.Diagnostics.Append(v.ValidateModel(ctx, &response.Schema)...)
		if response.Diagnostics.HasError() {
			response.Diagnostics.AddError("data source model validation error", w.opts.typeName)
			return
		}
	} else {
		response.Diagnostics.AddError("missing framework.DataSourceValidateModel", w.opts.typeName)
	}
}

func (w *wrappedDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	ctx, diags := w.opts.bootstrapContext(ctx, request.Config.GetAttribute, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	f := func(ctx context.Context, request *datasource.ReadRequest, response *datasource.ReadResponse) diag.Diagnostics {
		w.inner.Read(ctx, *request, response)
		return response.Diagnostics
	}
	response.Diagnostics.Append(interceptedHandler(w.opts.interceptors.dataSourceRead(), f, w.meta)(ctx, &request, response)...)
}

func (w *wrappedDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}

	ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	w.inner.Configure(ctx, request, response)
}

func (w *wrappedDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	if v, ok := w.inner.(datasource.DataSourceWithConfigValidators); ok {
		ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
		if diags.HasError() {
			tflog.Warn(ctx, "wrapping ConfigValidators", map[string]any{
				"data source":            w.opts.typeName,
				"bootstrapContext error": fwdiag.DiagnosticsString(diags),
			})

			return nil
		}

		return v.ConfigValidators(ctx)
	}

	return nil
}

func (w *wrappedDataSource) ValidateConfig(ctx context.Context, request datasource.ValidateConfigRequest, response *datasource.ValidateConfigResponse) {
	if v, ok := w.inner.(datasource.DataSourceWithValidateConfig); ok {
		ctx, diags := w.opts.bootstrapContext(ctx, request.Config.GetAttribute, w.meta)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		v.ValidateConfig(ctx, request, response)
	}
}

type wrappedEphemeralResourceOptions struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	interceptors     interceptorInvocations
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
	ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	f := func(ctx context.Context, request *ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) diag.Diagnostics {
		w.inner.Schema(ctx, *request, response)
		return response.Diagnostics
	}
	response.Diagnostics.Append(interceptedHandler(w.opts.interceptors.ephemeralResourceSchema(), f, w.meta)(ctx, &request, response)...)

	// Validate the ephemeral resource's model against the schema.
	if v, ok := w.inner.(framework.EphemeralResourceValidateModel); ok {
		response.Diagnostics.Append(v.ValidateModel(ctx, &response.Schema)...)
		if response.Diagnostics.HasError() {
			response.Diagnostics.AddError("ephemeral resource model validation error", w.opts.typeName)
			return
		}
	} else {
		response.Diagnostics.AddError("missing framework.EphemeralResourceValidateModel", w.opts.typeName)
	}
}

func (w *wrappedEphemeralResource) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	ctx, diags := w.opts.bootstrapContext(ctx, request.Config.GetAttribute, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	f := func(ctx context.Context, request *ephemeral.OpenRequest, response *ephemeral.OpenResponse) diag.Diagnostics {
		w.inner.Open(ctx, *request, response)
		return response.Diagnostics
	}
	response.Diagnostics.Append(interceptedHandler(w.opts.interceptors.ephemeralResourceOpen(), f, w.meta)(ctx, &request, response)...)
}

func (w *wrappedEphemeralResource) Configure(ctx context.Context, request ephemeral.ConfigureRequest, response *ephemeral.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}

	ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	w.inner.Configure(ctx, request, response)
}

func (w *wrappedEphemeralResource) Renew(ctx context.Context, request ephemeral.RenewRequest, response *ephemeral.RenewResponse) {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithRenew); ok {
		ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		f := func(ctx context.Context, request *ephemeral.RenewRequest, response *ephemeral.RenewResponse) diag.Diagnostics {
			v.Renew(ctx, *request, response)
			return response.Diagnostics
		}
		response.Diagnostics.Append(interceptedHandler(w.opts.interceptors.ephemeralResourceRenew(), f, w.meta)(ctx, &request, response)...)
	}
}

func (w *wrappedEphemeralResource) Close(ctx context.Context, request ephemeral.CloseRequest, response *ephemeral.CloseResponse) {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithClose); ok {
		ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		f := func(ctx context.Context, request *ephemeral.CloseRequest, response *ephemeral.CloseResponse) diag.Diagnostics {
			v.Close(ctx, *request, response)
			return response.Diagnostics
		}
		response.Diagnostics.Append(interceptedHandler(w.opts.interceptors.ephemeralResourceClose(), f, w.meta)(ctx, &request, response)...)
	}
}

func (w *wrappedEphemeralResource) ConfigValidators(ctx context.Context) []ephemeral.ConfigValidator {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithConfigValidators); ok {
		ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
		if diags.HasError() {
			tflog.Warn(ctx, "wrapping ConfigValidators", map[string]any{
				"ephemeral resource":     w.opts.typeName,
				"bootstrapContext error": fwdiag.DiagnosticsString(diags),
			})

			return nil
		}

		return v.ConfigValidators(ctx)
	}

	return nil
}

func (w *wrappedEphemeralResource) ValidateConfig(ctx context.Context, request ephemeral.ValidateConfigRequest, response *ephemeral.ValidateConfigResponse) {
	if v, ok := w.inner.(ephemeral.EphemeralResourceWithValidateConfig); ok {
		ctx, diags := w.opts.bootstrapContext(ctx, request.Config.GetAttribute, w.meta)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		v.ValidateConfig(ctx, request, response)
	}
}

type wrappedResourceOptions struct {
	// bootstrapContext is run on all wrapped methods before any interceptors.
	bootstrapContext contextFunc
	interceptors     interceptorInvocations
	typeName         string
	identity         types.Identity
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
	ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	f := func(ctx context.Context, request *resource.SchemaRequest, response *resource.SchemaResponse) diag.Diagnostics {
		w.inner.Schema(ctx, *request, response)
		return response.Diagnostics
	}
	response.Diagnostics.Append(interceptedHandler(w.opts.interceptors.resourceSchema(), f, w.meta)(ctx, &request, response)...)

	// Validate the resource's model against the schema.
	if v, ok := w.inner.(framework.ResourceValidateModel); ok {
		response.Diagnostics.Append(v.ValidateModel(ctx, &response.Schema)...)
		if response.Diagnostics.HasError() {
			response.Diagnostics.AddError("resource model validation error", w.opts.typeName)
			return
		}
	} else if w.opts.typeName != "aws_lexv2models_bot_version" { // Hacky yukkery caused by attribute of type map[string]Object.
		response.Diagnostics.AddError("missing framework.ResourceValidateModel", w.opts.typeName)
	}
}

func (w *wrappedResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	ctx, diags := w.opts.bootstrapContext(ctx, request.Plan.GetAttribute, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	f := func(ctx context.Context, request *resource.CreateRequest, response *resource.CreateResponse) diag.Diagnostics {
		w.inner.Create(ctx, *request, response)
		return response.Diagnostics
	}
	response.Diagnostics.Append(interceptedHandler(w.opts.interceptors.resourceCreate(), f, w.meta)(ctx, &request, response)...)
}

func (w *wrappedResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	ctx, diags := w.opts.bootstrapContext(ctx, request.State.GetAttribute, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	f := func(ctx context.Context, request *resource.ReadRequest, response *resource.ReadResponse) diag.Diagnostics {
		w.inner.Read(ctx, *request, response)
		return response.Diagnostics
	}
	response.Diagnostics.Append(interceptedHandler(w.opts.interceptors.resourceRead(), f, w.meta)(ctx, &request, response)...)
}

func (w *wrappedResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	ctx, diags := w.opts.bootstrapContext(ctx, request.Plan.GetAttribute, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	f := func(ctx context.Context, request *resource.UpdateRequest, response *resource.UpdateResponse) diag.Diagnostics {
		w.inner.Update(ctx, *request, response)
		return response.Diagnostics
	}
	response.Diagnostics.Append(interceptedHandler(w.opts.interceptors.resourceUpdate(), f, w.meta)(ctx, &request, response)...)
}

func (w *wrappedResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	ctx, diags := w.opts.bootstrapContext(ctx, request.State.GetAttribute, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	f := func(ctx context.Context, request *resource.DeleteRequest, response *resource.DeleteResponse) diag.Diagnostics {
		w.inner.Delete(ctx, *request, response)
		return response.Diagnostics
	}
	response.Diagnostics.Append(interceptedHandler(w.opts.interceptors.resourceDelete(), f, w.meta)(ctx, &request, response)...)
}

func (w *wrappedResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		w.meta = v
	}

	ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	w.inner.Configure(ctx, request, response)
}

func (w *wrappedResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	if v, ok := w.inner.(resource.ResourceWithImportState); ok {
		ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		ctx = importer.Context(ctx, w.meta)
		f := func(ctx context.Context, request *resource.ImportStateRequest, response *resource.ImportStateResponse) diag.Diagnostics {
			v.ImportState(ctx, *request, response)
			return response.Diagnostics
		}
		response.Diagnostics.Append(interceptedHandler(w.opts.interceptors.resourceImportState(), f, w.meta)(ctx, &request, response)...)

		return
	}

	response.Diagnostics.AddError(
		"Resource Import Not Implemented",
		"This resource does not support import. Please contact the provider developer for additional information.",
	)
}

func (w *wrappedResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	ctx, diags := w.opts.bootstrapContext(ctx, request.Config.GetAttribute, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// We run ModifyPlan interceptors even if the resource has not defined a ModifyPlan method.
	f := func(ctx context.Context, request *resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) diag.Diagnostics {
		return response.Diagnostics
	}
	if v, ok := w.inner.(resource.ResourceWithModifyPlan); ok {
		f = func(ctx context.Context, request *resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) diag.Diagnostics {
			v.ModifyPlan(ctx, *request, response)
			return response.Diagnostics
		}
	}
	response.Diagnostics.Append(interceptedHandler(w.opts.interceptors.resourceModifyPlan(), f, w.meta)(ctx, &request, response)...)
}

func (w *wrappedResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if v, ok := w.inner.(resource.ResourceWithConfigValidators); ok {
		ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
		if diags.HasError() {
			tflog.Warn(ctx, "wrapping ConfigValidators", map[string]any{
				"resource":               w.opts.typeName,
				"bootstrapContext error": fwdiag.DiagnosticsString(diags),
			})

			return nil
		}

		return v.ConfigValidators(ctx)
	}

	return nil
}

func (w *wrappedResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	ctx, diags := w.opts.bootstrapContext(ctx, request.Config.GetAttribute, w.meta)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if v, ok := w.inner.(resource.ResourceWithValidateConfig); ok {
		v.ValidateConfig(ctx, request, response)
	}
}

func (w *wrappedResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if v, ok := w.inner.(resource.ResourceWithUpgradeState); ok {
		ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
		if diags.HasError() {
			tflog.Warn(ctx, "wrapping UpgradeState", map[string]any{
				"resource":               w.opts.typeName,
				"bootstrapContext error": fwdiag.DiagnosticsString(diags),
			})

			return nil
		}

		return v.UpgradeState(ctx)
	}

	return nil
}

func (w *wrappedResource) MoveState(ctx context.Context) []resource.StateMover {
	if v, ok := w.inner.(resource.ResourceWithMoveState); ok {
		ctx, diags := w.opts.bootstrapContext(ctx, nil, w.meta)
		if diags.HasError() {
			tflog.Warn(ctx, "wrapping MoveState", map[string]any{
				"resource":               w.opts.typeName,
				"bootstrapContext error": fwdiag.DiagnosticsString(diags),
			})

			return nil
		}

		return v.MoveState(ctx)
	}

	return nil
}

func (w *wrappedResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	if len(w.opts.identity.Attributes) > 0 {
		resp.IdentitySchema = identity.NewIdentitySchema(w.opts.identity)
	}
}
