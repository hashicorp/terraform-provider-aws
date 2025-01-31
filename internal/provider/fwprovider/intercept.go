// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
)

type attributeGetter interface {
	GetAttribute(context.Context, path.Path, any) diag.Diagnostics
}

type interceptorOptions[Request, Response any] struct {
	c        *conns.AWSClient
	diags    diag.Diagnostics
	request  Request
	response *Response
	when     when
}

type interceptorFunc[Request, Response any] func(context.Context, interceptorOptions[Request, Response]) (context.Context, diag.Diagnostics)

// A data source interceptor is functionality invoked during the data source's CRUD request lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type dataSourceInterceptor interface {
	// read is invoke for a Read call.
	read(context.Context, interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]) (context.Context, diag.Diagnostics)
}

type dataSourceInterceptors []dataSourceInterceptor

type dataSourceInterceptorReadFunc interceptorFunc[datasource.ReadRequest, datasource.ReadResponse]

// read returns a slice of interceptors that run on data source Read.
func (s dataSourceInterceptors) read() []dataSourceInterceptorReadFunc {
	return slices.ApplyToAll(s, func(e dataSourceInterceptor) dataSourceInterceptorReadFunc {
		return e.read
	})
}

type ephemeralResourceInterceptor interface {
	// TODO implement me
}

type ephemeralResourceInterceptors []ephemeralResourceInterceptor

type resourceCRUDRequest interface {
	resource.CreateRequest | resource.ReadRequest | resource.UpdateRequest | resource.DeleteRequest
}
type resourceCRUDResponse interface {
	resource.CreateResponse | resource.ReadResponse | resource.UpdateResponse | resource.DeleteResponse
}

// A resource interceptor is functionality invoked during the resource's CRUD request lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type resourceInterceptor interface {
	// create is invoked for a Create call.
	create(context.Context, interceptorOptions[resource.CreateRequest, resource.CreateResponse]) (context.Context, diag.Diagnostics)
	// read is invoked for a Read call.
	read(context.Context, interceptorOptions[resource.ReadRequest, resource.ReadResponse]) (context.Context, diag.Diagnostics)
	// update is invoked for an Update call.
	update(context.Context, interceptorOptions[resource.UpdateRequest, resource.UpdateResponse]) (context.Context, diag.Diagnostics)
	// delete is invoked for a Delete call.
	delete(context.Context, interceptorOptions[resource.DeleteRequest, resource.DeleteResponse]) (context.Context, diag.Diagnostics)
}

type resourceInterceptors []resourceInterceptor

type resourceInterceptorFunc[Request resourceCRUDRequest, Response resourceCRUDResponse] interceptorFunc[Request, Response]

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

// when represents the point in the CRUD request lifecycle that an interceptor is run.
// Multiple values can be ORed together.
type when uint16

const (
	Before  when = 1 << iota // Interceptor is invoked before call to method in schema
	After                    // Interceptor is invoked after successful call to method in schema
	OnError                  // Interceptor is invoked after unsuccessful call to method in schema
	Finally                  // Interceptor is invoked after After or OnError
)

// TODO Share the intercepted handler logic between data sources and resources..

// interceptedDataSourceHandler returns a handler that invokes the specified data source Read handler, running any interceptors.
func interceptedDataSourceReadHandler(interceptors []dataSourceInterceptorReadFunc, f func(context.Context, datasource.ReadRequest, *datasource.ReadResponse) diag.Diagnostics, c *conns.AWSClient) func(context.Context, datasource.ReadRequest, *datasource.ReadResponse) diag.Diagnostics {
	return func(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) diag.Diagnostics {
		var diags diag.Diagnostics
		// Before interceptors are run first to last.
		forward := interceptors

		when := Before
		for _, v := range forward {
			opts := interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]{
				c:        c,
				diags:    diags,
				request:  request,
				response: response,
				when:     when,
			}
			ctx, diags = v(ctx, opts)

			// Short circuit if any Before interceptor errors.
			if diags.HasError() {
				return diags
			}
		}

		// All other interceptors are run last to first.
		reverse := slices.Reverse(forward)
		diags = f(ctx, request, response)

		if diags.HasError() {
			when = OnError
		} else {
			when = After
		}
		for _, v := range reverse {
			opts := interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]{
				c:        c,
				diags:    diags,
				request:  request,
				response: response,
				when:     when,
			}
			ctx, diags = v(ctx, opts)
		}

		when = Finally
		for _, v := range reverse {
			opts := interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]{
				c:        c,
				diags:    diags,
				request:  request,
				response: response,
				when:     when,
			}
			ctx, diags = v(ctx, opts)
		}

		return diags
	}
}

// interceptedResourceHandler returns a handler that invokes the specified resource CRUD handler, running any interceptors.
func interceptedResourceHandler[Request resourceCRUDRequest, Response resourceCRUDResponse](interceptors []resourceInterceptorFunc[Request, Response], f func(context.Context, Request, *Response) diag.Diagnostics, c *conns.AWSClient) func(context.Context, Request, *Response) diag.Diagnostics {
	return func(ctx context.Context, request Request, response *Response) diag.Diagnostics {
		var diags diag.Diagnostics
		// Before interceptors are run first to last.
		forward := interceptors

		when := Before
		for _, v := range forward {
			opts := interceptorOptions[Request, Response]{
				c:        c,
				diags:    diags,
				request:  request,
				response: response,
				when:     when,
			}
			ctx, diags = v(ctx, opts)

			// Short circuit if any Before interceptor errors.
			if diags.HasError() {
				return diags
			}
		}

		// All other interceptors are run last to first.
		reverse := slices.Reverse(forward)
		diags = f(ctx, request, response)

		if diags.HasError() {
			when = OnError
		} else {
			when = After
		}
		for _, v := range reverse {
			opts := interceptorOptions[Request, Response]{
				c:        c,
				diags:    diags,
				request:  request,
				response: response,
				when:     when,
			}
			ctx, diags = v(ctx, opts)
		}

		when = Finally
		for _, v := range reverse {
			opts := interceptorOptions[Request, Response]{
				c:        c,
				diags:    diags,
				request:  request,
				response: response,
				when:     when,
			}
			ctx, diags = v(ctx, opts)
		}

		return diags
	}
}
