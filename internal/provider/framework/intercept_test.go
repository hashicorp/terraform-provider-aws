// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestInterceptedHandler_Diags_FirstHasBeforeError(t *testing.T) {
	t.Parallel()

	expectedDiags := diag.Diagnostics{
		diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		diag.NewErrorDiagnostic("First interceptor Before error", "An error occurred in the first interceptor Before handler"),
	}

	first := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			diag.NewErrorDiagnostic("First interceptor Before error", "An error occurred in the first interceptor Before handler"),
		},
	})
	second := newMockInterceptor(map[when]diag.Diagnostics{})
	interceptors := []interceptorFunc[resource.SchemaRequest, resource.SchemaResponse]{
		first.Intercept,
		second.Intercept,
	}

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	var f mockInnerFunc
	handler := interceptedHandler(interceptors, f.Call, client)

	ctx := t.Context()
	var request resource.SchemaRequest
	response := resource.SchemaResponse{
		Diagnostics: diag.Diagnostics{
			diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		},
	}

	response.Diagnostics.Append(handler(ctx, &request, &response)...)

	if diff := cmp.Diff(response.Diagnostics, expectedDiags); diff != "" {
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
		diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		diag.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
	}

	first := newMockInterceptor(map[when]diag.Diagnostics{})
	second := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			diag.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
		},
	})
	interceptors := []interceptorFunc[resource.SchemaRequest, resource.SchemaResponse]{
		first.Intercept,
		second.Intercept,
	}

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	var f mockInnerFunc
	handler := interceptedHandler(interceptors, f.Call, client)

	ctx := t.Context()
	var request resource.SchemaRequest
	response := resource.SchemaResponse{
		Diagnostics: diag.Diagnostics{
			diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		},
	}

	response.Diagnostics.Append(handler(ctx, &request, &response)...)

	if diff := cmp.Diff(response.Diagnostics, expectedDiags); diff != "" {
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

func TestInterceptedHandler_Diags_FirstHasBeforeWarning(t *testing.T) {
	t.Parallel()

	expectedDiags := diag.Diagnostics{
		diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
	}

	first := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
		},
	})
	second := newMockInterceptor(map[when]diag.Diagnostics{})
	interceptors := []interceptorFunc[resource.SchemaRequest, resource.SchemaResponse]{
		first.Intercept,
		second.Intercept,
	}

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	var f mockInnerFunc
	handler := interceptedHandler(interceptors, f.Call, client)

	ctx := t.Context()
	var request resource.SchemaRequest
	response := resource.SchemaResponse{
		Diagnostics: diag.Diagnostics{
			diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		},
	}

	response.Diagnostics.Append(handler(ctx, &request, &response)...)

	if diff := cmp.Diff(response.Diagnostics, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics difference: %s", diff)
	}

	if !slices.Equal(first.called, []when{Before, After, Finally}) {
		t.Errorf("expected first interceptor to be called three times, got %v", first.called)
	}
	if !slices.Equal(second.called, []when{Before, After, Finally}) {
		t.Errorf("expected second interceptor to be called three times, got %v", second.called)
	}
	if f.count != 1 {
		t.Errorf("expected inner function to be called once, got %d", f.count)
	}
}

func TestInterceptedHandler_Diags_SecondHasBeforeWarning(t *testing.T) {
	t.Parallel()

	expectedDiags := diag.Diagnostics{
		diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
	}

	first := newMockInterceptor(map[when]diag.Diagnostics{})
	second := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
		},
	})
	interceptors := []interceptorFunc[resource.SchemaRequest, resource.SchemaResponse]{
		first.Intercept,
		second.Intercept,
	}

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	var f mockInnerFunc
	handler := interceptedHandler(interceptors, f.Call, client)

	ctx := t.Context()
	var request resource.SchemaRequest
	response := resource.SchemaResponse{
		Diagnostics: diag.Diagnostics{
			diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		},
	}

	response.Diagnostics.Append(handler(ctx, &request, &response)...)

	if diff := cmp.Diff(response.Diagnostics, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics difference: %s", diff)
	}

	if !slices.Equal(first.called, []when{Before, After, Finally}) {
		t.Errorf("expected first interceptor to be called three times, got %v", first.called)
	}
	if !slices.Equal(second.called, []when{Before, After, Finally}) {
		t.Errorf("expected second interceptor to be called three times, got %v", second.called)
	}
	if f.count != 1 {
		t.Errorf("expected inner function to be called once, got %d", f.count)
	}
}

func TestInterceptedHandler_Diags_FirstHasBeforeWarning_SecondHasBeforeError(t *testing.T) {
	t.Parallel()

	expectedDiags := diag.Diagnostics{
		diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
		diag.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
	}

	first := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
		},
	})
	second := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			diag.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
		},
	})
	interceptors := []interceptorFunc[resource.SchemaRequest, resource.SchemaResponse]{
		first.Intercept,
		second.Intercept,
	}

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	var f mockInnerFunc
	handler := interceptedHandler(interceptors, f.Call, client)

	ctx := t.Context()
	var request resource.SchemaRequest
	response := resource.SchemaResponse{
		Diagnostics: diag.Diagnostics{
			diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		},
	}

	response.Diagnostics.Append(handler(ctx, &request, &response)...)

	if diff := cmp.Diff(response.Diagnostics, expectedDiags); diff != "" {
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

func TestInterceptedHandler_Diags_InnerHasError(t *testing.T) {
	t.Parallel()

	expectedDiags := diag.Diagnostics{
		diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
	}

	first := mockInterceptor{}
	second := mockInterceptor{}
	interceptors := []interceptorFunc[resource.SchemaRequest, resource.SchemaResponse]{
		first.Intercept,
		second.Intercept,
	}

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	var f mockInnerFunc
	f.diags = diag.Diagnostics{
		diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
	}
	handler := interceptedHandler(interceptors, f.Call, client)

	ctx := t.Context()
	var request resource.SchemaRequest
	response := resource.SchemaResponse{
		Diagnostics: diag.Diagnostics{
			diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		},
	}

	response.Diagnostics.Append(handler(ctx, &request, &response)...)

	if diff := cmp.Diff(response.Diagnostics, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics difference: %s", diff)
	}

	if !slices.Equal(first.called, []when{Before, OnError, Finally}) {
		t.Errorf("expected first interceptor to be called three times, got %v", first.called)
	}
	if !slices.Equal(second.called, []when{Before, OnError, Finally}) {
		t.Errorf("expected second interceptor to be called three times, got %v", second.called)
	}
	if f.count != 1 {
		t.Errorf("expected inner function to be called once, got %d", f.count)
	}
}

func TestInterceptedHandler_Diags_InnerHasWarning(t *testing.T) {
	t.Parallel()

	expectedDiags := diag.Diagnostics{
		diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		diag.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
	}

	first := mockInterceptor{}
	second := mockInterceptor{}
	interceptors := []interceptorFunc[resource.SchemaRequest, resource.SchemaResponse]{
		first.Intercept,
		second.Intercept,
	}

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	var f mockInnerFunc
	f.diags = diag.Diagnostics{
		diag.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
	}
	handler := interceptedHandler(interceptors, f.Call, client)

	ctx := t.Context()
	var request resource.SchemaRequest
	response := resource.SchemaResponse{
		Diagnostics: diag.Diagnostics{
			diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		},
	}

	response.Diagnostics.Append(handler(ctx, &request, &response)...)

	if diff := cmp.Diff(response.Diagnostics, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics difference: %s", diff)
	}

	if !slices.Equal(first.called, []when{Before, After, Finally}) {
		t.Errorf("expected first interceptor to be called three times, got %v", first.called)
	}
	if !slices.Equal(second.called, []when{Before, After, Finally}) {
		t.Errorf("expected second interceptor to be called three times, got %v", second.called)
	}
	if f.count != 1 {
		t.Errorf("expected inner function to be called once, got %d", f.count)
	}
}

func TestInterceptedHandler_Diags_InnerHasError_FirstHasBeforeWarning(t *testing.T) {
	t.Parallel()

	expectedDiags := diag.Diagnostics{
		diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
		diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
	}

	first := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
		},
	})
	second := newMockInterceptor(map[when]diag.Diagnostics{})

	interceptors := []interceptorFunc[resource.SchemaRequest, resource.SchemaResponse]{
		first.Intercept,
		second.Intercept,
	}

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	var f mockInnerFunc
	f.diags = diag.Diagnostics{
		diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
	}
	handler := interceptedHandler(interceptors, f.Call, client)

	ctx := t.Context()
	var request resource.SchemaRequest
	response := resource.SchemaResponse{
		Diagnostics: diag.Diagnostics{
			diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		},
	}

	response.Diagnostics.Append(handler(ctx, &request, &response)...)

	if diff := cmp.Diff(response.Diagnostics, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics difference: %s", diff)
	}

	if !slices.Equal(first.called, []when{Before, OnError, Finally}) {
		t.Errorf("expected first interceptor to be called three times, got %v", first.called)
	}
	if !slices.Equal(second.called, []when{Before, OnError, Finally}) {
		t.Errorf("expected second interceptor to be called three times, got %v", second.called)
	}
	if f.count != 1 {
		t.Errorf("expected inner function to be called once, got %d", f.count)
	}
}

func TestInterceptedHandler_Diags_AllHaveWarnings(t *testing.T) {
	t.Parallel()

	expectedDiags := diag.Diagnostics{
		diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
		diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
		diag.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
		diag.NewWarningDiagnostic("Second interceptor After warning", "A warning occurred in the second interceptor After handler"),
		diag.NewWarningDiagnostic("First interceptor After warning", "A warning occurred in the first interceptor After handler"),
		diag.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
		diag.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
	}

	first := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
		},
		After: {
			diag.NewWarningDiagnostic("First interceptor After warning", "A warning occurred in the first interceptor After handler"),
		},
		Finally: {
			diag.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
		},
	})
	second := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
		},
		After: {
			diag.NewWarningDiagnostic("Second interceptor After warning", "A warning occurred in the second interceptor After handler"),
		},
		Finally: {
			diag.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
		},
	})
	interceptors := []interceptorFunc[resource.SchemaRequest, resource.SchemaResponse]{
		first.Intercept,
		second.Intercept,
	}

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	var f mockInnerFunc
	f.diags = diag.Diagnostics{
		diag.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
	}
	handler := interceptedHandler(interceptors, f.Call, client)

	ctx := t.Context()
	var request resource.SchemaRequest
	response := resource.SchemaResponse{
		Diagnostics: diag.Diagnostics{
			diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		},
	}

	response.Diagnostics.Append(handler(ctx, &request, &response)...)

	if diff := cmp.Diff(response.Diagnostics, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics difference: %s", diff)
	}

	if !slices.Equal(first.called, []when{Before, After, Finally}) {
		t.Errorf("expected first interceptor to be called three times, got %v", first.called)
	}
	if !slices.Equal(second.called, []when{Before, After, Finally}) {
		t.Errorf("expected second interceptor to be called three times, got %v", second.called)
	}
	if f.count != 1 {
		t.Errorf("expected inner function to be called once, got %d", f.count)
	}
}

func TestInterceptedHandler_Diags_InnerHasError_HandlersHaveWarnings(t *testing.T) {
	t.Parallel()

	expectedDiags := diag.Diagnostics{
		diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
		diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
		diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
		diag.NewWarningDiagnostic("Second interceptor OnError warning", "A warning occurred in the second interceptor OnError handler"),
		diag.NewWarningDiagnostic("First interceptor OnError warning", "A warning occurred in the first interceptor OnError handler"),
		diag.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
		diag.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
	}

	first := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
		},
		OnError: {
			diag.NewWarningDiagnostic("First interceptor OnError warning", "A warning occurred in the first interceptor OnError handler"),
		},
		Finally: {
			diag.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
		},
	})
	second := newMockInterceptor(map[when]diag.Diagnostics{
		Before: {
			diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
		},
		OnError: {
			diag.NewWarningDiagnostic("Second interceptor OnError warning", "A warning occurred in the second interceptor OnError handler"),
		},
		Finally: {
			diag.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
		},
	})
	interceptors := []interceptorFunc[resource.SchemaRequest, resource.SchemaResponse]{
		first.Intercept,
		second.Intercept,
	}

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	var f mockInnerFunc
	f.diags = diag.Diagnostics{
		diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
	}
	handler := interceptedHandler(interceptors, f.Call, client)

	ctx := t.Context()
	var request resource.SchemaRequest
	response := resource.SchemaResponse{
		Diagnostics: diag.Diagnostics{
			diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
		},
	}

	response.Diagnostics.Append(handler(ctx, &request, &response)...)

	if diff := cmp.Diff(response.Diagnostics, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics difference: %s", diff)
	}

	if !slices.Equal(first.called, []when{Before, OnError, Finally}) {
		t.Errorf("expected first interceptor to be called three times, got %v", first.called)
	}
	if !slices.Equal(second.called, []when{Before, OnError, Finally}) {
		t.Errorf("expected second interceptor to be called three times, got %v", second.called)
	}
	if f.count != 1 {
		t.Errorf("expected inner function to be called once, got %d", f.count)
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

func (m *mockInterceptor) Intercept(ctx context.Context, opts interceptorOptions[resource.SchemaRequest, resource.SchemaResponse]) diag.Diagnostics {
	m.called = append(m.called, opts.when)
	return m.diags[opts.when]
}

type mockInnerFunc struct {
	diags diag.Diagnostics
	count int
}

func (m *mockInnerFunc) Call(ctx context.Context, request *resource.SchemaRequest, response *resource.SchemaResponse) diag.Diagnostics {
	m.count++
	response.Diagnostics.Append(m.diags...)
	return response.Diagnostics
}
