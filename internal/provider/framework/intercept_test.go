// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestInterceptedHandler(t *testing.T) {
	t.Parallel()

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	testcases := map[string]struct {
		firstInterceptorDiags  map[when]diag.Diagnostics
		secondInterceptorDiags map[when]diag.Diagnostics
		innerFuncDiags         diag.Diagnostics
		expectedFirstCalls     []when
		expectedSecondCalls    []when
		expectedInnerCalls     int
		expectedDiags          diag.Diagnostics
	}{
		"First has Before error": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewErrorDiagnostic("First interceptor Before error", "An error occurred in the first interceptor Before handler"),
				},
			},
			expectedFirstCalls: []when{Before},
			expectedInnerCalls: 0,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("First interceptor Before error", "An error occurred in the first interceptor Before handler"),
			},
		},

		"Second has Before error": {
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
				},
			},
			expectedFirstCalls:  []when{Before},
			expectedSecondCalls: []when{Before},
			expectedInnerCalls:  0,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
			},
		},

		"First has Before warning": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
			},
		},

		"Second has Before warning": {
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				},
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
			},
		},

		"First has Before warning Second has Before error": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
			},
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
				},
			},
			expectedFirstCalls:  []when{Before},
			expectedSecondCalls: []when{Before},
			expectedInnerCalls:  0,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				diag.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
			},
		},

		"Inner has error": {
			innerFuncDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, OnError, Finally},
			expectedSecondCalls: []when{Before, OnError, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
		},

		"Inner has warning": {
			innerFuncDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
			},
		},

		"Inner has error First has Before warning": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
			},
			innerFuncDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, OnError, Finally},
			expectedSecondCalls: []when{Before, OnError, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
		},

		"All have warnings": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
				After: {
					diag.NewWarningDiagnostic("First interceptor After warning", "A warning occurred in the first interceptor After handler"),
				},
				Finally: {
					diag.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
				},
			},
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				},
				After: {
					diag.NewWarningDiagnostic("Second interceptor After warning", "A warning occurred in the second interceptor After handler"),
				},
				Finally: {
					diag.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
				},
			},
			innerFuncDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				diag.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
				diag.NewWarningDiagnostic("Second interceptor After warning", "A warning occurred in the second interceptor After handler"),
				diag.NewWarningDiagnostic("First interceptor After warning", "A warning occurred in the first interceptor After handler"),
				diag.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
				diag.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
			},
		},

		"Inner has error Handlers have warnings": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
				OnError: {
					diag.NewWarningDiagnostic("First interceptor OnError warning", "A warning occurred in the first interceptor OnError handler"),
				},
				Finally: {
					diag.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
				},
			},
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				},
				OnError: {
					diag.NewWarningDiagnostic("Second interceptor OnError warning", "A warning occurred in the second interceptor OnError handler"),
				},
				Finally: {
					diag.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
				},
			},
			innerFuncDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, OnError, Finally},
			expectedSecondCalls: []when{Before, OnError, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
				diag.NewWarningDiagnostic("Second interceptor OnError warning", "A warning occurred in the second interceptor OnError handler"),
				diag.NewWarningDiagnostic("First interceptor OnError warning", "A warning occurred in the first interceptor OnError handler"),
				diag.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
				diag.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			first := newMockInterceptor(tc.firstInterceptorDiags)
			second := newMockInterceptor(tc.secondInterceptorDiags)
			interceptors := []interceptorFunc[resource.SchemaRequest, resource.SchemaResponse]{
				first.Intercept,
				second.Intercept,
			}

			f := newMockInnerFunc(tc.innerFuncDiags)

			handler := interceptedHandler(interceptors, f.Call, resourceSchemaHasError, client)

			ctx := t.Context()
			var request resource.SchemaRequest
			response := resource.SchemaResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
				},
			}
			tc.expectedDiags = slices.Insert(tc.expectedDiags, 0, diag.Diagnostic(diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors")))

			handler(ctx, request, &response)

			if diff := cmp.Diff(response.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			if diff := cmp.Diff(first.called, tc.expectedFirstCalls); diff != "" {
				t.Errorf("unexpected first interceptor calls difference: %s", diff)
			}
			if diff := cmp.Diff(second.called, tc.expectedSecondCalls); diff != "" {
				t.Errorf("unexpected second interceptor calls difference: %s", diff)
			}
			if tc.expectedInnerCalls == 0 {
				if f.count != 0 {
					t.Errorf("expected inner function to not be called, got %d", f.count)
				}
			} else {
				if f.count != tc.expectedInnerCalls {
					t.Errorf("expected inner function to be called %d times, got %d", tc.expectedInnerCalls, f.count)
				}
			}
		})
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

func (m *mockInterceptor) Intercept(ctx context.Context, opts interceptorOptions[resource.SchemaRequest, resource.SchemaResponse]) {
	m.called = append(m.called, opts.when)
	opts.response.Diagnostics.Append(m.diags[opts.when]...)
}

type mockInnerFunc struct {
	diags diag.Diagnostics
	count int
}

func newMockInnerFunc(diags diag.Diagnostics) mockInnerFunc {
	return mockInnerFunc{
		diags: diags,
	}
}

func (m *mockInnerFunc) Call(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	m.count++
	response.Diagnostics.Append(m.diags...)
}

func TestInterceptedListHandler(t *testing.T) {
	t.Parallel()

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	testcases := map[string]struct {
		firstInterceptorDiags  map[when]diag.Diagnostics
		secondInterceptorDiags map[when]diag.Diagnostics
		innerFuncDiags         diag.Diagnostics
		expectedFirstCalls     []when
		expectedSecondCalls    []when
		expectedInnerCalls     int
		expectedDiags          diag.Diagnostics
	}{
		"First has Before error": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewErrorDiagnostic("First interceptor Before error", "An error occurred in the first interceptor Before handler"),
				},
			},
			expectedFirstCalls: []when{Before},
			expectedInnerCalls: 0,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("First interceptor Before error", "An error occurred in the first interceptor Before handler"),
			},
		},

		"Second has Before error": {
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
				},
			},
			expectedFirstCalls:  []when{Before},
			expectedSecondCalls: []when{Before},
			expectedInnerCalls:  0,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
			},
		},

		"First has Before warning": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
			},
		},

		"Second has Before warning": {
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				},
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
			},
		},

		"First has Before warning Second has Before error": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
			},
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
				},
			},
			expectedFirstCalls:  []when{Before},
			expectedSecondCalls: []when{Before},
			expectedInnerCalls:  0,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				diag.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
			},
		},

		"Inner has error": {
			innerFuncDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, OnError, Finally},
			expectedSecondCalls: []when{Before, OnError, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
		},

		"Inner has warning": {
			innerFuncDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
			},
		},

		"Inner has error First has Before warning": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
			},
			innerFuncDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, OnError, Finally},
			expectedSecondCalls: []when{Before, OnError, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
		},

		"All have warnings": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
				After: {
					diag.NewWarningDiagnostic("First interceptor After warning", "A warning occurred in the first interceptor After handler"),
				},
				Finally: {
					diag.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
				},
			},
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				},
				After: {
					diag.NewWarningDiagnostic("Second interceptor After warning", "A warning occurred in the second interceptor After handler"),
				},
				Finally: {
					diag.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
				},
			},
			innerFuncDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				diag.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
				diag.NewWarningDiagnostic("Second interceptor After warning", "A warning occurred in the second interceptor After handler"),
				diag.NewWarningDiagnostic("First interceptor After warning", "A warning occurred in the first interceptor After handler"),
				diag.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
				diag.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
			},
		},

		"Inner has error Handlers have warnings": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
				OnError: {
					diag.NewWarningDiagnostic("First interceptor OnError warning", "A warning occurred in the first interceptor OnError handler"),
				},
				Finally: {
					diag.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
				},
			},
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				},
				OnError: {
					diag.NewWarningDiagnostic("Second interceptor OnError warning", "A warning occurred in the second interceptor OnError handler"),
				},
				Finally: {
					diag.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
				},
			},
			innerFuncDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, OnError, Finally},
			expectedSecondCalls: []when{Before, OnError, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				diag.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				diag.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
				diag.NewWarningDiagnostic("Second interceptor OnError warning", "A warning occurred in the second interceptor OnError handler"),
				diag.NewWarningDiagnostic("First interceptor OnError warning", "A warning occurred in the first interceptor OnError handler"),
				diag.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
				diag.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			first := newMockListInterceptor(tc.firstInterceptorDiags)
			second := newMockListInterceptor(tc.secondInterceptorDiags)
			interceptors := []listInterceptorFunc[list.ListRequest, list.ListResultsStream]{
				first.Intercept,
				second.Intercept,
			}

			f := newMockInnerListFunc(tc.innerFuncDiags)

			handler := interceptedListHandler(interceptors, f.Call, client)

			ctx := t.Context()
			var request list.ListRequest
			response := list.ListResultsStream{
				Results: list.ListResultsStreamDiagnostics(diag.Diagnostics{
					diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors"),
				}),
			}
			tc.expectedDiags = slices.Insert(tc.expectedDiags, 0, diag.Diagnostic(diag.NewWarningDiagnostic("Pre-existing warning", "This is a pre-existing warning that should not be affected by the interceptors")))

			handler(ctx, request, &response)

			var diags diag.Diagnostics
			for d := range response.Results {
				if len(d.Diagnostics) > 0 {
					diags = append(diags, d.Diagnostics...)
				}
			}

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			if diff := cmp.Diff(first.called, tc.expectedFirstCalls); diff != "" {
				t.Errorf("unexpected first interceptor calls difference: %s", diff)
			}
			if diff := cmp.Diff(second.called, tc.expectedSecondCalls); diff != "" {
				t.Errorf("unexpected second interceptor calls difference: %s", diff)
			}
			if tc.expectedInnerCalls == 0 {
				if f.count != 0 {
					t.Errorf("expected inner function to not be called, got %d", f.count)
				}
			} else {
				if f.count != tc.expectedInnerCalls {
					t.Errorf("expected inner function to be called %d times, got %d", tc.expectedInnerCalls, f.count)
				}
			}
		})
	}
}

type mockListInterceptor struct {
	diags  map[when]diag.Diagnostics
	called []when
}

func newMockListInterceptor(diags map[when]diag.Diagnostics) *mockListInterceptor {
	return &mockListInterceptor{
		diags: diags,
	}
}

func (m *mockListInterceptor) Intercept(ctx context.Context, opts interceptorOptions[list.ListRequest, list.ListResultsStream]) diag.Diagnostics {
	m.called = append(m.called, opts.when)
	return m.diags[opts.when]
}

type mockInnerListFunc struct {
	diags diag.Diagnostics
	count int
}

func newMockInnerListFunc(diags diag.Diagnostics) mockInnerListFunc {
	return mockInnerListFunc{
		diags: diags,
	}
}

func (m *mockInnerListFunc) Call(ctx context.Context, request list.ListRequest, response *list.ListResultsStream) {
	m.count++
	if len(m.diags) > 0 {
		response.Results = list.ListResultsStreamDiagnostics(m.diags)
	} else {
		response.Results = list.NoListResults
	}
}
