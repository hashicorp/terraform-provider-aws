// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// schemaResourceData is an interface that implements a subset of schema.ResourceData's public methods.
type schemaResourceData interface {
	sdkv2.ResourceDiffer
	Set(string, any) error
}

type interceptorOptions[D any] struct {
	c    *conns.AWSClient
	d    D
	when when
	why  why
}

type (
	crudInterceptorOptions          = interceptorOptions[schemaResourceData]
	customizeDiffInterceptorOptions = interceptorOptions[*schema.ResourceDiff]
)

// An interceptor is functionality invoked during a request's lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type interceptor[D any, E any] interface {
	run(context.Context, interceptorOptions[D]) E
}

type (
	// crudInterceptor is functionality invoked during a CRUD request lifecycle.
	crudInterceptor = interceptor[schemaResourceData, diag.Diagnostics]
	// customizeDiffInterceptor is functionality invoked during a CustomizeDiff request lifecycle.
	customizeDiffInterceptor = interceptor[*schema.ResourceDiff, error]
)

type interceptorFunc[D any, E any] func(context.Context, interceptorOptions[D]) E

func (f interceptorFunc[D, E]) run(ctx context.Context, opts interceptorOptions[D]) E {
	return f(ctx, opts)
}

type (
	crudInterceptorFunc = interceptorFunc[schemaResourceData, diag.Diagnostics]
)

// interceptorInvocation represents a single interceptor invocation.
type interceptorInvocation struct {
	when        when
	why         why
	interceptor any
}

type typedInterceptorInvocation[D any, E any] struct {
	when        when
	why         why
	interceptor interceptor[D, E]
}

type (
	crudInterceptorInvocation          = typedInterceptorInvocation[schemaResourceData, diag.Diagnostics]
	customizeDiffInterceptorInvocation = typedInterceptorInvocation[*schema.ResourceDiff, error]
)

// when represents the point in the request lifecycle that an interceptor is run.
// Multiple values can be ORed together.
type when uint16

const (
	Before  when = 1 << iota // Interceptor is invoked before call to method in schema
	After                    // Interceptor is invoked after successful call to method in schema
	OnError                  // Interceptor is invoked after unsuccessful call to method in schema
	Finally                  // Interceptor is invoked after After or OnError
)

// why represents the operation(s) that an interceptor is run.
// Multiple values can be ORed together.
type why uint16

const (
	Create        why = 1 << iota // Interceptor is invoked for a Create call
	Read                          // Interceptor is invoked for a Read call
	Update                        // Interceptor is invoked for an Update call
	Delete                        // Interceptor is invoked for a Delete call
	CustomizeDiff                 // Interceptor is invoked for a CustomizeDiff call

	AllCRUDOps = Create | Read | Update | Delete // Interceptor is invoked for all CRUD calls
)

type interceptorInvocations []interceptorInvocation

func (s interceptorInvocations) why(why why) interceptorInvocations {
	return tfslices.Filter(s, func(e interceptorInvocation) bool {
		return e.why&why != 0
	})
}

// interceptedCRUDHandler returns a handler that invokes the specified CRUD handler, running any interceptors.
func interceptedCRUDHandler[F ~func(context.Context, *schema.ResourceData, any) diag.Diagnostics](bootstrapContext contextFunc, interceptorInvocations interceptorInvocations, f F, why why) F {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		ctx, diags := bootstrapContext(ctx, d.GetOk, meta)
		if diags.HasError() {
			return diags
		}

		// Before interceptors are run first to last.
		forward := make([]crudInterceptorInvocation, 0)
		for _, v := range interceptorInvocations.why(why) {
			if interceptor, ok := v.interceptor.(crudInterceptor); ok {
				forward = append(forward, crudInterceptorInvocation{
					when:        v.when,
					why:         v.why,
					interceptor: interceptor,
				})
			}
		}

		when := Before
		for _, v := range forward {
			if v.when&when != 0 {
				opts := crudInterceptorOptions{
					c:    meta.(*conns.AWSClient),
					d:    d,
					when: when,
					why:  why,
				}
				diags = append(diags, v.interceptor.run(ctx, opts)...)

				// Short circuit if any Before interceptor errors.
				if diags.HasError() {
					return diags
				}
			}
		}

		// All other interceptors are run last to first.
		reverse := tfslices.Reverse(forward)
		diags = f(ctx, d, meta)

		if diags.HasError() {
			when = OnError
		} else {
			when = After
		}
		for _, v := range reverse {
			if v.when&when != 0 {
				opts := crudInterceptorOptions{
					c:    meta.(*conns.AWSClient),
					d:    d,
					when: when,
					why:  why,
				}
				diags = append(diags, v.interceptor.run(ctx, opts)...)
			}
		}

		when = Finally
		for _, v := range reverse {
			if v.when&when != 0 {
				opts := crudInterceptorOptions{
					c:    meta.(*conns.AWSClient),
					d:    d,
					when: when,
					why:  why,
				}
				diags = append(diags, v.interceptor.run(ctx, opts)...)
			}
		}

		return diags
	}
}

// interceptedCustomizeDiffHandler returns a handler that invokes the specified CustomizeDiff handler, running any interceptors.
func interceptedCustomizeDiffHandler(bootstrapContext contextFunc, interceptorInvocations interceptorInvocations, f schema.CustomizeDiffFunc) schema.CustomizeDiffFunc {
	return func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
		ctx, diags := bootstrapContext(ctx, d.GetOk, meta)
		if diags.HasError() {
			return sdkdiag.DiagnosticsError(diags)
		}

		why := CustomizeDiff

		// Before interceptors are run first to last.
		forward := make([]customizeDiffInterceptorInvocation, 0)
		for _, v := range interceptorInvocations.why(why) {
			if interceptor, ok := v.interceptor.(customizeDiffInterceptor); ok {
				forward = append(forward, customizeDiffInterceptorInvocation{
					when:        v.when,
					why:         v.why,
					interceptor: interceptor,
				})
			}
		}

		when := Before
		for _, v := range forward {
			if v.when&when != 0 {
				opts := customizeDiffInterceptorOptions{
					c:    meta.(*conns.AWSClient),
					d:    d,
					when: when,
					why:  why,
				}
				// Short circuit if any Before interceptor errors.
				if err := v.interceptor.run(ctx, opts); err != nil {
					return err
				}
			}
		}

		// All other interceptors are run last to first.
		reverse := tfslices.Reverse(forward)
		var errs []error

		if err := f(ctx, d, meta); err != nil {
			when = OnError
			errs = append(errs, err)
		} else {
			when = After
		}
		for _, v := range reverse {
			if v.when&when != 0 {
				opts := customizeDiffInterceptorOptions{
					c:    meta.(*conns.AWSClient),
					d:    d,
					when: when,
					why:  why,
				}
				if err := v.interceptor.run(ctx, opts); err != nil {
					errs = append(errs, err)
				}
			}
		}

		when = Finally
		for _, v := range reverse {
			if v.when&when != 0 {
				opts := customizeDiffInterceptorOptions{
					c:    meta.(*conns.AWSClient),
					d:    d,
					when: when,
					why:  why,
				}
				if err := v.interceptor.run(ctx, opts); err != nil {
					errs = append(errs, err)
				}
			}
		}

		return errors.Join(errs...)
	}
}
