// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func TestInterceptorsWhy(t *testing.T) {
	t.Parallel()

	var interceptors interceptorItems

	interceptors = append(interceptors, interceptorItem{
		when: Before,
		why:  Create,
		interceptor: interceptorFunc(func(ctx context.Context, d schemaResourceData, meta any, when when, why why, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
			return ctx, diags
		}),
	})
	interceptors = append(interceptors, interceptorItem{
		when: After,
		why:  Delete,
		interceptor: interceptorFunc(func(ctx context.Context, d schemaResourceData, meta any, when when, why why, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
			return ctx, diags
		}),
	})
	interceptors = append(interceptors, interceptorItem{
		when: Before,
		why:  Create,
		interceptor: interceptorFunc(func(ctx context.Context, d schemaResourceData, meta any, when when, why why, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
			return ctx, diags
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

	var interceptors interceptorItems

	interceptors = append(interceptors, interceptorItem{
		when: Before,
		why:  Create,
		interceptor: interceptorFunc(func(ctx context.Context, d schemaResourceData, meta any, when when, why why, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
			return ctx, diags
		}),
	})
	interceptors = append(interceptors, interceptorItem{
		when: After,
		why:  Delete,
		interceptor: interceptorFunc(func(ctx context.Context, d schemaResourceData, meta any, when when, why why, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
			return ctx, diags
		}),
	})
	interceptors = append(interceptors, interceptorItem{
		when: Before,
		why:  Create,
		interceptor: interceptorFunc(func(ctx context.Context, d schemaResourceData, meta any, when when, why why, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
			return ctx, diags
		}),
	})

	var read schema.ReadContextFunc = func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		var diags diag.Diagnostics
		return sdkdiag.AppendErrorf(diags, "read error")
	}
	bootstrapContext := func(ctx context.Context, meta any) context.Context {
		return ctx
	}

	diags := interceptedHandler(bootstrapContext, interceptors, read, Read)(context.Background(), nil, 42)
	if got, want := len(diags), 1; got != want {
		t.Errorf("length of diags = %v, want %v", got, want)
	}
}
