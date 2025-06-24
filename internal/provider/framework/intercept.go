// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

type interceptorOptions[Request, Response any] struct {
	c        *conns.AWSClient
	request  *Request
	response *Response
	when     when
}

type interceptorFunc[Request, Response any] func(context.Context, interceptorOptions[Request, Response]) diag.Diagnostics

type interceptorInvocations []any

// A data source interceptor is functionality invoked during the data source's CRUD request lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type dataSourceCRUDInterceptor interface {
	// read is invoked for a Read call.
	read(context.Context, interceptorOptions[datasource.ReadRequest, datasource.ReadResponse]) diag.Diagnostics
}

// dataSourceRead returns a slice of interceptors that run on data source Read.
func (s interceptorInvocations) dataSourceRead() []interceptorFunc[datasource.ReadRequest, datasource.ReadResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(dataSourceCRUDInterceptor)
		return ok
	}), func(e any) interceptorFunc[datasource.ReadRequest, datasource.ReadResponse] {
		return e.(dataSourceCRUDInterceptor).read
	})
}

type dataSourceSchemaInterceptor interface {
	// schema is invoked for a Schema call.
	schema(context.Context, interceptorOptions[datasource.SchemaRequest, datasource.SchemaResponse]) diag.Diagnostics
}

// dataSourceSchema returns a slice of interceptors that run on data source Schema.
func (s interceptorInvocations) dataSourceSchema() []interceptorFunc[datasource.SchemaRequest, datasource.SchemaResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(dataSourceSchemaInterceptor)
		return ok
	}), func(e any) interceptorFunc[datasource.SchemaRequest, datasource.SchemaResponse] {
		return e.(dataSourceSchemaInterceptor).schema
	})
}

type ephemeralResourceORCInterceptor interface {
	// open is invoked for an Open call.
	open(context.Context, interceptorOptions[ephemeral.OpenRequest, ephemeral.OpenResponse]) diag.Diagnostics
	// renew is invoked for a Renew call.
	renew(context.Context, interceptorOptions[ephemeral.RenewRequest, ephemeral.RenewResponse]) diag.Diagnostics
	// close is invoked for a Close call.
	close(context.Context, interceptorOptions[ephemeral.CloseRequest, ephemeral.CloseResponse]) diag.Diagnostics
}

// ephemeralResourceOpen returns a slice of interceptors that run on ephemeral resource Open.
func (s interceptorInvocations) ephemeralResourceOpen() []interceptorFunc[ephemeral.OpenRequest, ephemeral.OpenResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(ephemeralResourceORCInterceptor)
		return ok
	}), func(e any) interceptorFunc[ephemeral.OpenRequest, ephemeral.OpenResponse] {
		return e.(ephemeralResourceORCInterceptor).open
	})
}

// ephemeralResourceRenew returns a slice of interceptors that run on ephemeral resource Renew.
func (s interceptorInvocations) ephemeralResourceRenew() []interceptorFunc[ephemeral.RenewRequest, ephemeral.RenewResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(ephemeralResourceORCInterceptor)
		return ok
	}), func(e any) interceptorFunc[ephemeral.RenewRequest, ephemeral.RenewResponse] {
		return e.(ephemeralResourceORCInterceptor).renew
	})
}

// ephemeralResourceRenew returns a slice of interceptors that run on ephemeral resource Close.
func (s interceptorInvocations) ephemeralResourceClose() []interceptorFunc[ephemeral.CloseRequest, ephemeral.CloseResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(ephemeralResourceORCInterceptor)
		return ok
	}), func(e any) interceptorFunc[ephemeral.CloseRequest, ephemeral.CloseResponse] {
		return e.(ephemeralResourceORCInterceptor).close
	})
}

// ephemeralResourceNoOpORCInterceptor is a no-op implementation of the ephemeralResourceORCInterceptor interface.
// It can be embedded into a struct to provide default behavior for the open, renew, and close methods.
type ephemeralResourceNoOpORCInterceptor struct{}

func (r ephemeralResourceNoOpORCInterceptor) open(ctx context.Context, opts interceptorOptions[ephemeral.OpenRequest, ephemeral.OpenResponse]) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func (r ephemeralResourceNoOpORCInterceptor) renew(ctx context.Context, opts interceptorOptions[ephemeral.RenewRequest, ephemeral.RenewResponse]) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func (r ephemeralResourceNoOpORCInterceptor) close(ctx context.Context, opts interceptorOptions[ephemeral.CloseRequest, ephemeral.CloseResponse]) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

type ephemeralResourceSchemaInterceptor interface {
	// schema is invoked for a Schema call.
	schema(context.Context, interceptorOptions[ephemeral.SchemaRequest, ephemeral.SchemaResponse]) diag.Diagnostics
}

// ephemeralResourceSchema returns a slice of interceptors that run on ephemeral resource Schema.
func (s interceptorInvocations) ephemeralResourceSchema() []interceptorFunc[ephemeral.SchemaRequest, ephemeral.SchemaResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(ephemeralResourceSchemaInterceptor)
		return ok
	}), func(e any) interceptorFunc[ephemeral.SchemaRequest, ephemeral.SchemaResponse] {
		return e.(ephemeralResourceSchemaInterceptor).schema
	})
}

// A resource interceptor is functionality invoked during the resource's CRUD request lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type resourceCRUDInterceptor interface {
	// create is invoked for a Create call.
	create(context.Context, interceptorOptions[resource.CreateRequest, resource.CreateResponse]) diag.Diagnostics
	// read is invoked for a Read call.
	read(context.Context, interceptorOptions[resource.ReadRequest, resource.ReadResponse]) diag.Diagnostics
	// update is invoked for an Update call.
	update(context.Context, interceptorOptions[resource.UpdateRequest, resource.UpdateResponse]) diag.Diagnostics
	// delete is invoked for a Delete call.
	delete(context.Context, interceptorOptions[resource.DeleteRequest, resource.DeleteResponse]) diag.Diagnostics
}

// resourceCreate returns a slice of interceptors that run on resource Create.
func (s interceptorInvocations) resourceCreate() []interceptorFunc[resource.CreateRequest, resource.CreateResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(resourceCRUDInterceptor)
		return ok
	}), func(e any) interceptorFunc[resource.CreateRequest, resource.CreateResponse] {
		return e.(resourceCRUDInterceptor).create
	})
}

// resourceRead returns a slice of interceptors that run on resource Read.
func (s interceptorInvocations) resourceRead() []interceptorFunc[resource.ReadRequest, resource.ReadResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(resourceCRUDInterceptor)
		return ok
	}), func(e any) interceptorFunc[resource.ReadRequest, resource.ReadResponse] {
		return e.(resourceCRUDInterceptor).read
	})
}

// resourceUpdate returns a slice of interceptors that run on resource Update.
func (s interceptorInvocations) resourceUpdate() []interceptorFunc[resource.UpdateRequest, resource.UpdateResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(resourceCRUDInterceptor)
		return ok
	}), func(e any) interceptorFunc[resource.UpdateRequest, resource.UpdateResponse] {
		return e.(resourceCRUDInterceptor).update
	})
}

// resourceDelete returns a slice of interceptors that run on resource Delete.
func (s interceptorInvocations) resourceDelete() []interceptorFunc[resource.DeleteRequest, resource.DeleteResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(resourceCRUDInterceptor)
		return ok
	}), func(e any) interceptorFunc[resource.DeleteRequest, resource.DeleteResponse] {
		return e.(resourceCRUDInterceptor).delete
	})
}

// resourceNoOpCRUDInterceptor is a no-op implementation of the resourceCRUDInterceptor interface.
// It can be embedded into a struct to provide default behavior for the create, read, update, and delete methods.
type resourceNoOpCRUDInterceptor struct{}

func (r resourceNoOpCRUDInterceptor) create(ctx context.Context, opts interceptorOptions[resource.CreateRequest, resource.CreateResponse]) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func (r resourceNoOpCRUDInterceptor) read(ctx context.Context, opts interceptorOptions[resource.ReadRequest, resource.ReadResponse]) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func (r resourceNoOpCRUDInterceptor) update(ctx context.Context, opts interceptorOptions[resource.UpdateRequest, resource.UpdateResponse]) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func (r resourceNoOpCRUDInterceptor) delete(ctx context.Context, opts interceptorOptions[resource.DeleteRequest, resource.DeleteResponse]) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

type resourceSchemaInterceptor interface {
	// schema is invoked for a Schema call.
	schema(context.Context, interceptorOptions[resource.SchemaRequest, resource.SchemaResponse]) diag.Diagnostics
}

// resourceSchema returns a slice of interceptors that run on resource Schema.
func (s interceptorInvocations) resourceSchema() []interceptorFunc[resource.SchemaRequest, resource.SchemaResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(resourceSchemaInterceptor)
		return ok
	}), func(e any) interceptorFunc[resource.SchemaRequest, resource.SchemaResponse] {
		return e.(resourceSchemaInterceptor).schema
	})
}

type resourceModifyPlanInterceptor interface {
	// modifyPlan is invoked for a ModifyPlan call.
	modifyPlan(context.Context, interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse]) diag.Diagnostics
}

// resourceModifyPlan returns a slice of interceptors that run on resource ModifyPlan.
func (s interceptorInvocations) resourceModifyPlan() []interceptorFunc[resource.ModifyPlanRequest, resource.ModifyPlanResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(resourceModifyPlanInterceptor)
		return ok
	}), func(e any) interceptorFunc[resource.ModifyPlanRequest, resource.ModifyPlanResponse] {
		return e.(resourceModifyPlanInterceptor).modifyPlan
	})
}

type resourceImportStateInterceptor interface {
	// importState is invoked for an ImportState call.
	importState(context.Context, interceptorOptions[resource.ImportStateRequest, resource.ImportStateResponse]) diag.Diagnostics
}

// resourceSchema returns a slice of interceptors that run on resource Schema.
func (s interceptorInvocations) resourceImportState() []interceptorFunc[resource.ImportStateRequest, resource.ImportStateResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(resourceImportStateInterceptor)
		return ok
	}), func(e any) interceptorFunc[resource.ImportStateRequest, resource.ImportStateResponse] {
		return e.(resourceImportStateInterceptor).importState
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
	datasource.SchemaRequest |
		datasource.ReadRequest |
		ephemeral.SchemaRequest |
		ephemeral.OpenRequest |
		ephemeral.RenewRequest |
		ephemeral.CloseRequest |
		resource.SchemaRequest |
		resource.CreateRequest |
		resource.ReadRequest |
		resource.UpdateRequest |
		resource.DeleteRequest |
		resource.ModifyPlanRequest |
		resource.ImportStateRequest
}

// interceptedResponse represents a Plugin Framework response type that can be intercepted.
type interceptedResponse interface {
	datasource.SchemaResponse |
		datasource.ReadResponse |
		ephemeral.SchemaResponse |
		ephemeral.OpenResponse |
		ephemeral.RenewResponse |
		ephemeral.CloseResponse |
		resource.SchemaResponse |
		resource.CreateResponse |
		resource.ReadResponse |
		resource.UpdateResponse |
		resource.DeleteResponse |
		resource.ModifyPlanResponse |
		resource.ImportStateResponse
}

// interceptedHandler returns a handler that runs any interceptors.
func interceptedHandler[Request interceptedRequest, Response interceptedResponse](interceptors []interceptorFunc[Request, Response], f func(context.Context, *Request, *Response) diag.Diagnostics, c *conns.AWSClient) func(context.Context, *Request, *Response) diag.Diagnostics {
	return func(ctx context.Context, request *Request, response *Response) diag.Diagnostics {
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
		reverse := tfslices.Reverse(forward)
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
