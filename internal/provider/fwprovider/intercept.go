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
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
)

type interceptorOptions[Request, Response any] struct {
	c        *conns.AWSClient
	request  Request
	response *Response
	when     when
}

type interceptorFunc[Request, Response any] func(context.Context, interceptorOptions[Request, Response]) diag.Diagnostics

// A data source interceptor is functionality invoked during the data source's CRUD request lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type dataSourceInterceptor interface {
	// read is invoked for a Read call.
	read(context.Context, interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]) diag.Diagnostics
}

type dataSourceInterceptors []dataSourceInterceptor

// read returns a slice of interceptors that run on data source Read.
func (s dataSourceInterceptors) read() []interceptorFunc[datasource.ReadRequest, datasource.ReadResponse] {
	return slices.ApplyToAll(s, func(e dataSourceInterceptor) interceptorFunc[datasource.ReadRequest, datasource.ReadResponse] {
		return e.read
	})
}

type ephemeralResourceInterceptor interface {
	// open is invoked for an Open call.
	open(context.Context, interceptorOptions[ephemeral.OpenRequest, ephemeral.OpenResponse]) diag.Diagnostics
	// renew is invoked for a Renew call.
	renew(context.Context, interceptorOptions[ephemeral.RenewRequest, ephemeral.RenewResponse]) diag.Diagnostics
	// close is invoked for a Close call.
	close(context.Context, interceptorOptions[ephemeral.CloseRequest, ephemeral.CloseResponse]) diag.Diagnostics
}

type ephemeralResourceInterceptors []ephemeralResourceInterceptor

// open returns a slice of interceptors that run on ephemeral resource Open.
func (s ephemeralResourceInterceptors) open() []interceptorFunc[ephemeral.OpenRequest, ephemeral.OpenResponse] {
	return slices.ApplyToAll(s, func(e ephemeralResourceInterceptor) interceptorFunc[ephemeral.OpenRequest, ephemeral.OpenResponse] {
		return e.open
	})
}

// renew returns a slice of interceptors that run on ephemeral resource Renew.
func (s ephemeralResourceInterceptors) renew() []interceptorFunc[ephemeral.RenewRequest, ephemeral.RenewResponse] {
	return slices.ApplyToAll(s, func(e ephemeralResourceInterceptor) interceptorFunc[ephemeral.RenewRequest, ephemeral.RenewResponse] {
		return e.renew
	})
}

// close returns a slice of interceptors that run on ephemeral resource Renew.
func (s ephemeralResourceInterceptors) close() []interceptorFunc[ephemeral.CloseRequest, ephemeral.CloseResponse] {
	return slices.ApplyToAll(s, func(e ephemeralResourceInterceptor) interceptorFunc[ephemeral.CloseRequest, ephemeral.CloseResponse] {
		return e.close
	})
}

// A resource interceptor is functionality invoked during the resource's CRUD request lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type resourceInterceptor interface {
	// create is invoked for a Create call.
	create(context.Context, interceptorOptions[resource.CreateRequest, resource.CreateResponse]) diag.Diagnostics
	// read is invoked for a Read call.
	read(context.Context, interceptorOptions[resource.ReadRequest, resource.ReadResponse]) diag.Diagnostics
	// update is invoked for an Update call.
	update(context.Context, interceptorOptions[resource.UpdateRequest, resource.UpdateResponse]) diag.Diagnostics
	// delete is invoked for a Delete call.
	delete(context.Context, interceptorOptions[resource.DeleteRequest, resource.DeleteResponse]) diag.Diagnostics
}

type resourceInterceptors []resourceInterceptor

// create returns a slice of interceptors that run on resource Create.
func (s resourceInterceptors) create() []interceptorFunc[resource.CreateRequest, resource.CreateResponse] {
	return slices.ApplyToAll(s, func(e resourceInterceptor) interceptorFunc[resource.CreateRequest, resource.CreateResponse] {
		return e.create
	})
}

// read returns a slice of interceptors that run on resource Read.
func (s resourceInterceptors) read() []interceptorFunc[resource.ReadRequest, resource.ReadResponse] {
	return slices.ApplyToAll(s, func(e resourceInterceptor) interceptorFunc[resource.ReadRequest, resource.ReadResponse] {
		return e.read
	})
}

// update returns a slice of interceptors that run on resource Update.
func (s resourceInterceptors) update() []interceptorFunc[resource.UpdateRequest, resource.UpdateResponse] {
	return slices.ApplyToAll(s, func(e resourceInterceptor) interceptorFunc[resource.UpdateRequest, resource.UpdateResponse] {
		return e.update
	})
}

// delete returns a slice of interceptors that run on resource Delete.
func (s resourceInterceptors) delete() []interceptorFunc[resource.DeleteRequest, resource.DeleteResponse] {
	return slices.ApplyToAll(s, func(e resourceInterceptor) interceptorFunc[resource.DeleteRequest, resource.DeleteResponse] {
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

// interceptedRequest represents a Plugin Framework request type that can be intercepted.
type interceptedRequest interface {
	datasource.ReadRequest | ephemeral.OpenRequest | ephemeral.RenewRequest | ephemeral.CloseRequest | resource.CreateRequest | resource.ReadRequest | resource.UpdateRequest | resource.DeleteRequest
}

// interceptedResponse represents a Plugin Framework response type that can be intercepted.
type interceptedResponse interface {
	datasource.ReadResponse | ephemeral.OpenResponse | ephemeral.RenewResponse | ephemeral.CloseResponse | resource.CreateResponse | resource.ReadResponse | resource.UpdateResponse | resource.DeleteResponse
}

// interceptedHandler returns a handler that runs any interceptors.
func interceptedHandler[Request interceptedRequest, Response interceptedResponse](interceptors []interceptorFunc[Request, Response], f func(context.Context, Request, *Response) diag.Diagnostics, c *conns.AWSClient) func(context.Context, Request, *Response) diag.Diagnostics {
	return func(ctx context.Context, request Request, response *Response) diag.Diagnostics {
		var diags diag.Diagnostics
		// Before interceptors are run first to last.
		forward := interceptors

		when := Before
		for _, v := range forward {
			opts := interceptorOptions[Request, Response]{
				c:        c,
				request:  request,
				response: response,
				when:     when,
			}
			diags.Append(v(ctx, opts)...)

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
				request:  request,
				response: response,
				when:     when,
			}
			diags.Append(v(ctx, opts)...)
		}

		when = Finally
		for _, v := range reverse {
			opts := interceptorOptions[Request, Response]{
				c:        c,
				request:  request,
				response: response,
				when:     when,
			}
			diags.Append(v(ctx, opts)...)
		}

		return diags
	}
}
