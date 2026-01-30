// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

type awsClient interface {
	AccountID(context.Context) string
	Region(context.Context) string
	DefaultTagsConfig(ctx context.Context) *tftags.DefaultConfig
	IgnoreTagsConfig(ctx context.Context) *tftags.IgnoreConfig
	Partition(context.Context) string
	ServicePackage(_ context.Context, name string) conns.ServicePackage
	TagPolicyConfig(ctx context.Context) *tftags.TagPolicyConfig
	ValidateInContextRegionInPartition(ctx context.Context) error
	AwsConfig(context.Context) aws.Config
}

type interceptorOptions[Request, Response any] struct {
	c        awsClient
	request  *Request
	response *Response
	when     when
}

type interceptorFunc[Request, Response any] func(context.Context, interceptorOptions[Request, Response])

type interceptorInvocations []any

// A data source interceptor is functionality invoked during the data source's CRUD request lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type dataSourceCRUDInterceptor interface {
	// read is invoked for a Read call.
	read(context.Context, interceptorOptions[datasource.ReadRequest, datasource.ReadResponse])
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
	schema(context.Context, interceptorOptions[datasource.SchemaRequest, datasource.SchemaResponse])
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
	open(context.Context, interceptorOptions[ephemeral.OpenRequest, ephemeral.OpenResponse])
	// renew is invoked for a Renew call.
	renew(context.Context, interceptorOptions[ephemeral.RenewRequest, ephemeral.RenewResponse])
	// close is invoked for a Close call.
	close(context.Context, interceptorOptions[ephemeral.CloseRequest, ephemeral.CloseResponse])
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

func (r ephemeralResourceNoOpORCInterceptor) open(ctx context.Context, opts interceptorOptions[ephemeral.OpenRequest, ephemeral.OpenResponse]) {
}

func (r ephemeralResourceNoOpORCInterceptor) renew(ctx context.Context, opts interceptorOptions[ephemeral.RenewRequest, ephemeral.RenewResponse]) {
}

func (r ephemeralResourceNoOpORCInterceptor) close(ctx context.Context, opts interceptorOptions[ephemeral.CloseRequest, ephemeral.CloseResponse]) {
}

type ephemeralResourceSchemaInterceptor interface {
	// schema is invoked for a Schema call.
	schema(context.Context, interceptorOptions[ephemeral.SchemaRequest, ephemeral.SchemaResponse])
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
	create(context.Context, interceptorOptions[resource.CreateRequest, resource.CreateResponse])
	// read is invoked for a Read call.
	read(context.Context, interceptorOptions[resource.ReadRequest, resource.ReadResponse])
	// update is invoked for an Update call.
	update(context.Context, interceptorOptions[resource.UpdateRequest, resource.UpdateResponse])
	// delete is invoked for a Delete call.
	delete(context.Context, interceptorOptions[resource.DeleteRequest, resource.DeleteResponse])
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

func (r resourceNoOpCRUDInterceptor) create(ctx context.Context, opts interceptorOptions[resource.CreateRequest, resource.CreateResponse]) {
}

func (r resourceNoOpCRUDInterceptor) read(ctx context.Context, opts interceptorOptions[resource.ReadRequest, resource.ReadResponse]) {
}

func (r resourceNoOpCRUDInterceptor) update(ctx context.Context, opts interceptorOptions[resource.UpdateRequest, resource.UpdateResponse]) {
}

func (r resourceNoOpCRUDInterceptor) delete(ctx context.Context, opts interceptorOptions[resource.DeleteRequest, resource.DeleteResponse]) {
}

type resourceSchemaInterceptor interface {
	// schema is invoked for a Schema call.
	schema(context.Context, interceptorOptions[resource.SchemaRequest, resource.SchemaResponse])
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
	modifyPlan(context.Context, interceptorOptions[resource.ModifyPlanRequest, resource.ModifyPlanResponse])
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
	importState(context.Context, interceptorOptions[resource.ImportStateRequest, resource.ImportStateResponse])
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

type listInterceptorFunc[Request, Response any] func(context.Context, interceptorOptions[Request, Response]) diag.Diagnostics

type listResourceListInterceptor interface {
	list(context.Context, interceptorOptions[list.ListRequest, list.ListResultsStream]) diag.Diagnostics
}

// resourceList returns a slice of interceptors that run on resource List.
func (s interceptorInvocations) resourceList() []listInterceptorFunc[list.ListRequest, list.ListResultsStream] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(listResourceListInterceptor)
		return ok
	}), func(e any) listInterceptorFunc[list.ListRequest, list.ListResultsStream] {
		return e.(listResourceListInterceptor).list
	})
}

type listResourceSchemaInterceptor interface {
	schema(context.Context, interceptorOptions[list.ListResourceSchemaRequest, list.ListResourceSchemaResponse])
}

// resourceListResourceConfigSchema returns a slice of interceptors that run on resource ListResourceConfigSchema.
func (s interceptorInvocations) resourceListResourceConfigSchema() []interceptorFunc[list.ListResourceSchemaRequest, list.ListResourceSchemaResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(listResourceSchemaInterceptor)
		return ok
	}), func(e any) interceptorFunc[list.ListResourceSchemaRequest, list.ListResourceSchemaResponse] {
		return e.(listResourceSchemaInterceptor).schema
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

// Only generate strings for use in tests
//go:generate stringer -type=when -output=when_string_test.go

// An action interceptor is functionality invoked during the action's lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type actionInvokeInterceptor interface {
	// invoke is invoked for an Invoke call.
	invoke(context.Context, interceptorOptions[action.InvokeRequest, action.InvokeResponse])
}

// actionInvoke returns a slice of interceptors that run on action Invoke.
func (s interceptorInvocations) actionInvoke() []interceptorFunc[action.InvokeRequest, action.InvokeResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(actionInvokeInterceptor)
		return ok
	}), func(e any) interceptorFunc[action.InvokeRequest, action.InvokeResponse] {
		return e.(actionInvokeInterceptor).invoke
	})
}

type actionSchemaInterceptor interface {
	// schema is invoked for a Schema call.
	schema(context.Context, interceptorOptions[action.SchemaRequest, action.SchemaResponse])
}

// actionSchema returns a slice of interceptors that run on action Schema.
func (s interceptorInvocations) actionSchema() []interceptorFunc[action.SchemaRequest, action.SchemaResponse] {
	return tfslices.ApplyToAll(tfslices.Filter(s, func(e any) bool {
		_, ok := e.(actionSchemaInterceptor)
		return ok
	}), func(e any) interceptorFunc[action.SchemaRequest, action.SchemaResponse] {
		return e.(actionSchemaInterceptor).schema
	})
}

// interceptedRequest represents a Plugin Framework request type that can be intercepted.
type interceptedRequest interface {
	action.SchemaRequest |
		action.InvokeRequest |
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
		resource.ImportStateRequest |
		list.ListResourceSchemaRequest
}

// interceptedResponse represents a Plugin Framework response type that can be intercepted.
type interceptedResponse interface {
	action.SchemaResponse |
		action.InvokeResponse |
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
		resource.ImportStateResponse |
		list.ListResourceSchemaResponse
}

type innerFunc[Request, Response any] func(ctx context.Context, request Request, response *Response)

// interceptedHandler returns a handler that runs any interceptors.
func interceptedHandler[Request interceptedRequest, Response interceptedResponse](interceptors []interceptorFunc[Request, Response], f innerFunc[Request, Response], hasError hasErrorFn[Response], c awsClient) func(context.Context, Request, *Response) {
	return func(ctx context.Context, request Request, response *Response) {
		opts := interceptorOptions[Request, Response]{
			c:        c,
			request:  &request,
			response: response,
		}

		// Before interceptors are run first to last.
		opts.when = Before
		for v := range slices.Values(interceptors) {
			v(ctx, opts)

			// Short circuit if any Before interceptor errors.
			if hasError(response) {
				return
			}
		}

		f(ctx, request, response)

		// All other interceptors are run last to first.
		if hasError(response) {
			opts.when = OnError
		} else {
			opts.when = After
		}
		for v := range tfslices.BackwardValues(interceptors) {
			v(ctx, opts)
		}

		opts.when = Finally
		for v := range tfslices.BackwardValues(interceptors) {
			v(ctx, opts)
		}
	}
}

type hasErrorFn[Response interceptedResponse] func(response *Response) bool

func dataSourceSchemaHasError(response *datasource.SchemaResponse) bool {
	return response.Diagnostics.HasError()
}

func dataSourceReadHasError(response *datasource.ReadResponse) bool {
	return response.Diagnostics.HasError()
}

func ephemeralSchemaHasError(response *ephemeral.SchemaResponse) bool {
	return response.Diagnostics.HasError()
}

func ephemeralOpenHasError(response *ephemeral.OpenResponse) bool {
	return response.Diagnostics.HasError()
}

func ephemeralRenewHasError(response *ephemeral.RenewResponse) bool {
	return response.Diagnostics.HasError()
}

func ephemeralCloseHasError(response *ephemeral.CloseResponse) bool {
	return response.Diagnostics.HasError()
}

func resourceSchemaHasError(response *resource.SchemaResponse) bool {
	return response.Diagnostics.HasError()
}

func resourceCreateHasError(response *resource.CreateResponse) bool {
	return response.Diagnostics.HasError()
}

func resourceReadHasError(response *resource.ReadResponse) bool {
	return response.Diagnostics.HasError()
}

func resourceUpdateHasError(response *resource.UpdateResponse) bool {
	return response.Diagnostics.HasError()
}

func resourceDeleteHasError(response *resource.DeleteResponse) bool {
	return response.Diagnostics.HasError()
}

func resourceModifyPlanHasError(response *resource.ModifyPlanResponse) bool {
	return response.Diagnostics.HasError()
}

func resourceImportStateHasError(response *resource.ImportStateResponse) bool {
	return response.Diagnostics.HasError()
}

func actionSchemaHasError(response *action.SchemaResponse) bool {
	return response.Diagnostics.HasError()
}

func actionInvokeHasError(response *action.InvokeResponse) bool {
	return response.Diagnostics.HasError()
}

func listResourceConfigSchemaHasError(response *list.ListResourceSchemaResponse) bool {
	return response.Diagnostics.HasError()
}

func interceptedListHandler(interceptors []listInterceptorFunc[list.ListRequest, list.ListResultsStream], f func(context.Context, list.ListRequest, *list.ListResultsStream), c awsClient) func(context.Context, list.ListRequest, *list.ListResultsStream) {
	return func(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
		opts := interceptorOptions[list.ListRequest, list.ListResultsStream]{
			c:        c,
			request:  &request,
			response: stream,
		}

		// Before interceptors are run first to last.
		opts.when = Before
		for v := range slices.Values(interceptors) {
			diags := v(ctx, opts)
			if len(diags) > 0 {
				stream.Results = tfiter.Concat(stream.Results, list.ListResultsStreamDiagnostics(diags))
			}
			if diags.HasError() {
				return
			}
		}

		// Stash `stream.Results` so that inner function can be unaware of interceptors.
		resultStream := stream.Results
		stream.Results = nil

		f(ctx, request, stream)
		innerResultStream := stream.Results

		stream.Results = tfiter.Concat(resultStream, func(yield func(list.ListResult) bool) {
			var hasError bool
			for v := range innerResultStream {
				if v.Diagnostics.HasError() {
					hasError = true
				}
				if !yield(v) {
					return
				}
			}

			// All other interceptors are run last to first.
			if hasError {
				opts.when = OnError
			} else {
				opts.when = After
			}
			for v := range tfslices.BackwardValues(interceptors) {
				diags := v(ctx, opts)
				if len(diags) > 0 {
					if !yield(list.ListResult{Diagnostics: diags}) {
						return
					}
				}
			}

			opts.when = Finally
			for v := range tfslices.BackwardValues(interceptors) {
				diags := v(ctx, opts)
				if len(diags) > 0 {
					if !yield(list.ListResult{Diagnostics: diags}) {
						return
					}
				}
			}
		})
	}
}
