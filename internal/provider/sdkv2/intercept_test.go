// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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

func TestInterceptedHandler_Diags_FirstHasBeforeError(t *testing.T) {
	t.Parallel()

	expectedDiags := diag.Diagnostics{
		errs.NewErrorDiagnostic("First interceptor Before error", "An error occurred in the first interceptor Before handler"),
	}

	first := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			errs.NewErrorDiagnostic("First interceptor Before error", "An error occurred in the first interceptor Before handler"),
		},
	})
	second := newMockInterceptor(map[when]diag.Diagnostics{})
	interceptors := append(
		first.Invocations(),
		second.Invocations()...,
	)

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	contextFunc := func(ctx context.Context, _ getAttributeFunc, meta any) (context.Context, error) {
		return ctx, nil
	}

	var f mockInnerFunc
	handler := interceptedCRUDHandler(contextFunc, interceptors, f.Call, Create)

	ctx := t.Context()
	diags := handler(ctx, nil, client)

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics difference: %s", diff)
	}

	if !slices.Equal(first.called, []when{Before}) {
		t.Errorf("expected first interceptor to be called once, got %v", first.called)
	}
	if !slices.Equal(second.called, []when{}) {
		t.Errorf("expected second interceptor to not be called, got %v", second.called)
	}
	if f.count != 0 {
		t.Errorf("expected inner function to not be called, got %d", f.count)
	}
}

func TestInterceptedHandler_Diags_SecondHasBeforeError(t *testing.T) {
	t.Parallel()

	expectedDiags := diag.Diagnostics{
		errs.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
	}

	first := newMockInterceptor(map[when]diag.Diagnostics{})
	second := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			errs.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
		},
	})
	interceptors := append(
		first.Invocations(),
		second.Invocations()...,
	)

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	contextFunc := func(ctx context.Context, _ getAttributeFunc, meta any) (context.Context, error) {
		return ctx, nil
	}

	var f mockInnerFunc
	handler := interceptedCRUDHandler(contextFunc, interceptors, f.Call, Create)

	ctx := t.Context()
	diags := handler(ctx, nil, client)

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics difference: %s", diff)
	}

	if !slices.Equal(first.called, []when{Before}) {
		t.Errorf("expected first interceptor to be called once, got %v", first.called)
	}
	if !slices.Equal(second.called, []when{Before}) {
		t.Errorf("expected second interceptor to be called once, got %v", second.called)
	}
	if f.count != 0 {
		t.Errorf("expected inner function to not be called, got %d", f.count)
	}
}

func TestInterceptedHandler_Diags_FirstHasBeforeWarning_SecondHasBeforeError(t *testing.T) {
	t.Parallel()

	expectedDiags := diag.Diagnostics{
		errs.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
		errs.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
	}

	first := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			errs.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
		},
	})
	second := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			errs.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
		},
	})
	interceptors := append(
		first.Invocations(),
		second.Invocations()...,
	)

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	contextFunc := func(ctx context.Context, _ getAttributeFunc, meta any) (context.Context, error) {
		return ctx, nil
	}

	var f mockInnerFunc
	handler := interceptedCRUDHandler(contextFunc, interceptors, f.Call, Create)

	ctx := t.Context()
	diags := handler(ctx, nil, client)

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics difference: %s", diff)
	}

	if !slices.Equal(first.called, []when{Before}) {
		t.Errorf("expected first interceptor to be called once, got %v", first.called)
	}
	if !slices.Equal(second.called, []when{Before}) {
		t.Errorf("expected second interceptor to be called once, got %v", second.called)
	}
	if f.count != 0 {
		t.Errorf("expected inner function to not be called, got %d", f.count)
	}
}

type mockInterceptor struct {
	diags  map[when]diag.Diagnostics
	called []when
}

func newMockInterceptor(diags map[when]diag.Diagnostics) *mockInterceptor {
	return &mockInterceptor{
		diags: diags,
	}
}

func (m *mockInterceptor) Invocations() interceptorInvocations {
	return interceptorInvocations{
		{
			why:         AllCRUDOps,
			when:        Before | After | OnError | Finally,
			interceptor: m.interceptor(),
		},
	}
}

func (m *mockInterceptor) interceptor() crudInterceptor {
	return crudInterceptorFunc(func(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
		m.called = append(m.called, opts.when)
		return m.diags[opts.when]
	})
}

type mockInnerFunc struct {
	diags diag.Diagnostics
	count int
}

func (m *mockInnerFunc) Call(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	m.count++
	return m.diags
}
