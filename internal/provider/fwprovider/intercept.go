package fwprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
)

type resourceCRUDRequest interface {
	resource.CreateRequest | resource.ReadRequest | resource.UpdateRequest | resource.DeleteRequest
}
type resourceCRUDResponse interface {
	resource.CreateResponse | resource.ReadResponse | resource.UpdateResponse | resource.DeleteResponse
}

type resourceInterceptor interface {
	create(context.Context, resource.CreateRequest, *resource.CreateResponse, *conns.AWSClient, when, diag.Diagnostics) (context.Context, diag.Diagnostics)
	read(context.Context, resource.ReadRequest, *resource.ReadResponse, *conns.AWSClient, when, diag.Diagnostics) (context.Context, diag.Diagnostics)
	update(context.Context, resource.UpdateRequest, *resource.UpdateResponse, *conns.AWSClient, when, diag.Diagnostics) (context.Context, diag.Diagnostics)
	delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse, *conns.AWSClient, when, diag.Diagnostics) (context.Context, diag.Diagnostics)
}

type resourceInterceptorFunc[Request resourceCRUDRequest, Response resourceCRUDResponse] func(context.Context, Request, *Response, *conns.AWSClient, when, diag.Diagnostics) (context.Context, diag.Diagnostics)

type resourceInterceptors []resourceInterceptor

// create returns a slice of interceptors that run on resource Create.
func (s resourceInterceptors) create() []resourceInterceptorFunc[resource.CreateRequest, resource.CreateResponse] {
	return slices.ApplyToAll(s, func(e resourceInterceptor) resourceInterceptorFunc[resource.CreateRequest, resource.CreateResponse] {
		return e.create
	})
}

// read returns a slice of interceptors that run on resource Read.
func (s resourceInterceptors) read() []resourceInterceptorFunc[resource.ReadRequest, resource.ReadResponse] {
	return slices.ApplyToAll(s, func(e resourceInterceptor) resourceInterceptorFunc[resource.ReadRequest, resource.ReadResponse] {
		return e.read
	})
}

// update returns a slice of interceptors that run on resource Update.
func (s resourceInterceptors) update() []resourceInterceptorFunc[resource.UpdateRequest, resource.UpdateResponse] {
	return slices.ApplyToAll(s, func(e resourceInterceptor) resourceInterceptorFunc[resource.UpdateRequest, resource.UpdateResponse] {
		return e.update
	})
}

// delete returns a slice of interceptors that run on resource Delete.
func (s resourceInterceptors) delete() []resourceInterceptorFunc[resource.DeleteRequest, resource.DeleteResponse] {
	return slices.ApplyToAll(s, func(e resourceInterceptor) resourceInterceptorFunc[resource.DeleteRequest, resource.DeleteResponse] {
		return e.delete
	})
}

// interceptedHandler returns a handler that invokes the specified CRUD handler, running any interceptors.
func interceptedHandler[Request resourceCRUDRequest, Response resourceCRUDResponse](interceptors []resourceInterceptorFunc[Request, Response], f func(context.Context, Request, *Response) diag.Diagnostics, meta *conns.AWSClient) func(context.Context, Request, *Response) diag.Diagnostics {
	return func(ctx context.Context, request Request, response *Response) diag.Diagnostics {
		var diags diag.Diagnostics
		forward := interceptors

		when := Before
		for _, v := range forward {
			ctx, diags = v(ctx, request, response, meta, when, diags)

			// Short circuit if any Before interceptor errors.
			if diags.HasError() {
				return diags
			}
		}

		reverse := slices.Reverse(forward)
		diags = f(ctx, request, response)

		if diags.HasError() {
			when = OnError
		} else {
			when = After
		}
		for _, v := range reverse {
			ctx, diags = v(ctx, request, response, meta, when, diags)
		}

		when = Finally
		for _, v := range reverse {
			ctx, diags = v(ctx, request, response, meta, when, diags)
		}

		return diags
	}
}

// when represents the point in the CRUD request lifecycle that an interceptor is run.
// Multiple values can be ORed together.
type when uint16

const (
	Before  when = 1 << iota // Interceptor is invoked before call to method in schema
	After                    // Interceptor is invoked after successful call to method in schema
	OnError                  // Interceptor is invoked after unsuccessful call to method in schema
	Finally                  // Interceptor is invoked after After or OnError
)

type contextFunc func(context.Context, *conns.AWSClient) context.Context

// wrappedDataSource wraps a data source, adding common functionality.
type wrappedDataSource struct {
	bootstrapContext contextFunc
	inner            datasource.DataSourceWithConfigure
	meta             *conns.AWSClient
}

func newWrappedDataSource(bootstrapContext contextFunc, inner datasource.DataSourceWithConfigure) datasource.DataSourceWithConfigure {
	return &wrappedDataSource{
		bootstrapContext: bootstrapContext,
		inner:            inner,
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
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Read(ctx, request, response)
}

func (w *wrappedDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	ctx = w.bootstrapContext(ctx, w.meta)
	w.inner.Configure(ctx, request, response)
}

// wrappedResource wraps a resource, adding common functionality.
type wrappedResource struct {
	bootstrapContext contextFunc
	inner            resource.ResourceWithConfigure
	interceptors     resourceInterceptors
	meta             *conns.AWSClient
}

func newWrappedResource(bootstrapContext contextFunc, inner resource.ResourceWithConfigure) resource.ResourceWithConfigure {
	return &wrappedResource{
		bootstrapContext: bootstrapContext,
		inner:            inner,
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
	diags := interceptedHandler(w.interceptors.create(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	f := func(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) diag.Diagnostics {
		w.inner.Read(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	diags := interceptedHandler(w.interceptors.read(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	f := func(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) diag.Diagnostics {
		w.inner.Update(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	diags := interceptedHandler(w.interceptors.update(), f, w.meta)(ctx, request, response)
	response.Diagnostics = diags
}

func (w *wrappedResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	f := func(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) diag.Diagnostics {
		w.inner.Delete(ctx, request, response)
		return response.Diagnostics
	}
	ctx = w.bootstrapContext(ctx, w.meta)
	diags := interceptedHandler(w.interceptors.delete(), f, w.meta)(ctx, request, response)
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

		return
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
