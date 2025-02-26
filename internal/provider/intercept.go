// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// schemaResourceData is an interface that implements a subset of schema.ResourceData's public methods.
type schemaResourceData interface {
	sdkv2.ResourceDiffer
	Set(string, any) error
}

type interceptorOptions struct {
	c    *conns.AWSClient
	d    schemaResourceData
	when when
	why  why
}

// An interceptor is functionality invoked during the CRUD request lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type interceptor interface {
	run(context.Context, interceptorOptions) diag.Diagnostics
}

type interceptorFunc func(context.Context, interceptorOptions) diag.Diagnostics

func (f interceptorFunc) run(ctx context.Context, opts interceptorOptions) diag.Diagnostics {
	return f(ctx, opts)
}

// interceptorItem represents a single interceptor invocation.
type interceptorItem struct {
	when        when
	why         why
	interceptor interceptor
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

// why represents the CRUD operation(s) that an interceptor is run.
// Multiple values can be ORed together.
type why uint16

const (
	Create why = 1 << iota // Interceptor is invoked for a Create call
	Read                   // Interceptor is invoked for a Read call
	Update                 // Interceptor is invoked for an Update call
	Delete                 // Interceptor is invoked for a Delete call

	AllOps = Create | Read | Update | Delete // Interceptor is invoked for all calls
)

type interceptorItems []interceptorItem

// why returns a slice of interceptors that run for the specified CRUD operation.
func (s interceptorItems) why(why why) interceptorItems {
	return slices.Filter(s, func(e interceptorItem) bool {
		return e.why&why != 0
	})
}

// interceptedHandler returns a handler that invokes the specified CRUD handler, running any interceptors.
func interceptedHandler[F ~func(context.Context, *schema.ResourceData, any) diag.Diagnostics](bootstrapContext contextFunc, interceptors interceptorItems, f F, why why) F {
	return func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		ctx, diags := bootstrapContext(ctx, d.GetOk, meta)
		if diags.HasError() {
			return diags
		}

		// Before interceptors are run first to last.
		forward := interceptors.why(why)

		when := Before
		for _, v := range forward {
			if v.when&when != 0 {
				opts := interceptorOptions{
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
		reverse := slices.Reverse(forward)
		diags = f(ctx, d, meta)

		if diags.HasError() {
			when = OnError
		} else {
			when = After
		}
		for _, v := range reverse {
			if v.when&when != 0 {
				opts := interceptorOptions{
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
				opts := interceptorOptions{
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
