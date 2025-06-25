// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

type (
	crudInterceptorFunc = interceptorFunc1[schemaResourceData, diag.Diagnostics]
)

func TestInterceptorsWhy(t *testing.T) {
	t.Parallel()

	var interceptors interceptorInvocations
	interceptors = append(interceptors, interceptorInvocation{
		when: Before,
		why:  Create,
		interceptor: crudInterceptorFunc(func(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
			var diags diag.Diagnostics
			return diags
		}),
	})
	interceptors = append(interceptors, interceptorInvocation{
		when: After,
		why:  Delete,
		interceptor: crudInterceptorFunc(func(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
			var diags diag.Diagnostics
			return diags
		}),
	})
	interceptors = append(interceptors, interceptorInvocation{
		when: Before,
		why:  Create,
		interceptor: crudInterceptorFunc(func(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
			var diags diag.Diagnostics
			return diags
		}),
	})

	if got, want := len(interceptors.why(Create)), 2; got != want {
		t.Errorf("length of interceptors.Why(Create) = %v, want %v", got, want)
	}
	if got, want := len(interceptors.why(Read)), 0; got != want {
		t.Errorf("length of interceptors.Why(Read) = %v, want %v", got, want)
	}
	if got, want := len(interceptors.why(Update)), 0; got != want {
		t.Errorf("length of interceptors.Why(Update) = %v, want %v", got, want)
	}
	if got, want := len(interceptors.why(Delete)), 1; got != want {
		t.Errorf("length of interceptors.Why(Delete) = %v, want %v", got, want)
	}
}

func TestInterceptedHandler(t *testing.T) {
	t.Parallel()

	var interceptors interceptorInvocations

	interceptors = append(interceptors, interceptorInvocation{
		when: Before,
		why:  Create,
		interceptor: crudInterceptorFunc(func(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
			var diags diag.Diagnostics
			return diags
		}),
	})
	interceptors = append(interceptors, interceptorInvocation{
		when: After,
		why:  Delete,
		interceptor: crudInterceptorFunc(func(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
			var diags diag.Diagnostics
			return diags
		}),
	})
	interceptors = append(interceptors, interceptorInvocation{
		when: Before,
		why:  Create,
		interceptor: crudInterceptorFunc(func(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
			var diags diag.Diagnostics
			return diags
		}),
	})

	var read schema.ReadContextFunc = func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
		var diags diag.Diagnostics
		return sdkdiag.AppendErrorf(diags, "read error")
	}
	bootstrapContext := func(ctx context.Context, _ getAttributeFunc, meta any) (context.Context, error) {
		return ctx, nil
	}

	diags := interceptedCRUDHandler(bootstrapContext, interceptors, read, Read)(context.Background(), nil, 42)
	if got, want := len(diags), 1; got != want {
		t.Errorf("length of diags = %v, want %v", got, want)
	}
}
