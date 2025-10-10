// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"
	"errors"
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

func TestInterceptedCRUDHandler(t *testing.T) {
	t.Parallel()

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	contextFunc := func(ctx context.Context, _ getAttributeFunc, _ getProviderMetaFunc, meta any) (context.Context, error) {
		return ctx, nil
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
					errs.NewErrorDiagnostic("First interceptor Before error", "An error occurred in the first interceptor Before handler"),
				},
			},
			expectedFirstCalls: []when{Before},
			expectedInnerCalls: 0,
			expectedDiags: diag.Diagnostics{
				errs.NewErrorDiagnostic("First interceptor Before error", "An error occurred in the first interceptor Before handler"),
			},
		},

		"Second has Before error": {
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					errs.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
				},
			},
			expectedFirstCalls:  []when{Before},
			expectedSecondCalls: []when{Before},
			expectedInnerCalls:  0,
			expectedDiags: diag.Diagnostics{
				errs.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
			},
		},

		"First has Before warning": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					errs.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				errs.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
			},
		},

		"Second has Before warning": {
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					errs.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				},
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				errs.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
			},
		},

		"First has Before warning Second has Before error": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					errs.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
			},
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					errs.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
				},
			},
			expectedFirstCalls:  []when{Before},
			expectedSecondCalls: []when{Before},
			expectedInnerCalls:  0,
			expectedDiags: diag.Diagnostics{
				errs.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				errs.NewErrorDiagnostic("Second interceptor Before error", "An error occurred in the second interceptor Before handler"),
			},
		},

		"Inner has error": {
			innerFuncDiags: diag.Diagnostics{
				errs.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, OnError, Finally},
			expectedSecondCalls: []when{Before, OnError, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				errs.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
		},

		"Inner has warning": {
			innerFuncDiags: diag.Diagnostics{
				errs.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				errs.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
			},
		},

		"Inner has error First has Before warning": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					errs.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
			},
			innerFuncDiags: diag.Diagnostics{
				errs.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, OnError, Finally},
			expectedSecondCalls: []when{Before, OnError, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				errs.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				errs.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
		},

		"All have warnings": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					errs.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
				After: {
					errs.NewWarningDiagnostic("First interceptor After warning", "A warning occurred in the first interceptor After handler"),
				},
				Finally: {
					errs.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
				},
			},
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					errs.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				},
				After: {
					errs.NewWarningDiagnostic("Second interceptor After warning", "A warning occurred in the second interceptor After handler"),
				},
				Finally: {
					errs.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
				},
			},
			innerFuncDiags: diag.Diagnostics{
				errs.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				errs.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				errs.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				errs.NewWarningDiagnostic("Inner function warning", "A warning occurred in the inner function"),
				errs.NewWarningDiagnostic("Second interceptor After warning", "A warning occurred in the second interceptor After handler"),
				errs.NewWarningDiagnostic("First interceptor After warning", "A warning occurred in the first interceptor After handler"),
				errs.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
				errs.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
			},
		},

		"Inner has error Handlers have warnings": {
			firstInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					errs.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				},
				OnError: {
					errs.NewWarningDiagnostic("First interceptor OnError warning", "A warning occurred in the first interceptor OnError handler"),
				},
				Finally: {
					errs.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
				},
			},
			secondInterceptorDiags: map[when]diag.Diagnostics{
				Before: {
					errs.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				},
				OnError: {
					errs.NewWarningDiagnostic("Second interceptor OnError warning", "A warning occurred in the second interceptor OnError handler"),
				},
				Finally: {
					errs.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
				},
			},
			innerFuncDiags: diag.Diagnostics{
				errs.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
			},
			expectedFirstCalls:  []when{Before, OnError, Finally},
			expectedSecondCalls: []when{Before, OnError, Finally},
			expectedInnerCalls:  1,
			expectedDiags: diag.Diagnostics{
				errs.NewWarningDiagnostic("First interceptor Before warning", "A warning occurred in the first interceptor Before handler"),
				errs.NewWarningDiagnostic("Second interceptor Before warning", "A warning occurred in the second interceptor Before handler"),
				errs.NewErrorDiagnostic("Inner function error", "An error occurred in the inner function"),
				errs.NewWarningDiagnostic("Second interceptor OnError warning", "A warning occurred in the second interceptor OnError handler"),
				errs.NewWarningDiagnostic("First interceptor OnError warning", "A warning occurred in the first interceptor OnError handler"),
				errs.NewWarningDiagnostic("Second interceptor Finally warning", "A warning occurred in the second interceptor Finally handler"),
				errs.NewWarningDiagnostic("First interceptor Finally warning", "A warning occurred in the first interceptor Finally handler"),
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			first := newMockCRUDInterceptor(tc.firstInterceptorDiags)
			second := newMockCRUDInterceptor(tc.secondInterceptorDiags)
			interceptors := append(
				first.Invocations(),
				second.Invocations()...,
			)

			f := newMockInnerCRUDFunc(tc.innerFuncDiags)

			handler := interceptedCRUDHandler(contextFunc, interceptors, f.Call, Create)

			ctx := t.Context()
			diags := handler(ctx, nil, client)

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

type mockCRUDInterceptor struct {
	diags  map[when]diag.Diagnostics
	called []when
}

func newMockCRUDInterceptor(diags map[when]diag.Diagnostics) *mockCRUDInterceptor {
	if diags == nil {
		diags = make(map[when]diag.Diagnostics)
	}
	return &mockCRUDInterceptor{
		diags: diags,
	}
}

func (m *mockCRUDInterceptor) Invocations() interceptorInvocations {
	return interceptorInvocations{
		{
			why:         AllCRUDOps,
			when:        Before | After | OnError | Finally,
			interceptor: m,
		},
	}
}

func (m *mockCRUDInterceptor) run(_ context.Context, opts crudInterceptorOptions) diag.Diagnostics {
	m.called = append(m.called, opts.when)
	return m.diags[opts.when]
}

type mockInnerCRUDFunc struct {
	diags diag.Diagnostics
	count int
}

func newMockInnerCRUDFunc(diags diag.Diagnostics) mockInnerCRUDFunc {
	return mockInnerCRUDFunc{
		diags: diags,
	}
}

func (m *mockInnerCRUDFunc) Call(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	m.count++
	return m.diags
}

func TestInterceptedCustomizeDiffHandler(t *testing.T) {
	t.Parallel()

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	contextFunc := func(ctx context.Context, _ getAttributeFunc, _ getProviderMetaFunc, meta any) (context.Context, error) {
		return ctx, nil
	}

	testcases := map[string]struct {
		firstInterceptorErrors  map[when]error
		secondInterceptorErrors map[when]error
		innerFuncError          error
		expectedFirstCalls      []when
		expectedSecondCalls     []when
		expectedInnerCalls      int
		expectedError           error
	}{
		"First has Before error": {
			firstInterceptorErrors: map[when]error{
				Before: errors.New("First interceptor Before error"),
			},
			expectedFirstCalls: []when{Before},
			expectedInnerCalls: 0,
			expectedError:      errors.New("First interceptor Before error"),
		},

		"Second has Before error": {
			secondInterceptorErrors: map[when]error{
				Before: errors.New("Second interceptor Before error"),
			},
			expectedFirstCalls:  []when{Before},
			expectedSecondCalls: []when{Before},
			expectedInnerCalls:  0,
			expectedError:       errors.New("Second interceptor Before error"),
		},

		"Inner has error": {
			innerFuncError:      errors.New("Inner function error"),
			expectedFirstCalls:  []when{Before, OnError, Finally},
			expectedSecondCalls: []when{Before, OnError, Finally},
			expectedInnerCalls:  1,
			expectedError: errors.Join(
				errors.New("Inner function error"),
			),
		},

		"All have errors": {
			firstInterceptorErrors: map[when]error{
				OnError: errors.New("First interceptor OnError error"),
				Finally: errors.New("First interceptor Finally error"),
			},
			secondInterceptorErrors: map[when]error{
				OnError: errors.New("Second interceptor OnError error"),
				Finally: errors.New("Second interceptor Finally error"),
			},
			innerFuncError:      errors.New("Inner function error"),
			expectedFirstCalls:  []when{Before, OnError, Finally},
			expectedSecondCalls: []when{Before, OnError, Finally},
			expectedInnerCalls:  1,
			expectedError: errors.Join(
				errors.New("Inner function error"),
				errors.New("Second interceptor OnError error"),
				errors.New("First interceptor OnError error"),
				errors.New("Second interceptor Finally error"),
				errors.New("First interceptor Finally error"),
			),
		},

		"Handlers have errors": {
			firstInterceptorErrors: map[when]error{
				After:   errors.New("First interceptor After error"),
				Finally: errors.New("First interceptor Finally error"),
			},
			secondInterceptorErrors: map[when]error{
				After:   errors.New("Second interceptor After error"),
				Finally: errors.New("Second interceptor Finally error"),
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedError: errors.Join(
				errors.New("Second interceptor After error"),
				errors.New("First interceptor After error"),
				errors.New("Second interceptor Finally error"),
				errors.New("First interceptor Finally error"),
			),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			first := newMockCustomizeDiffInterceptor(tc.firstInterceptorErrors)
			second := newMockCustomizeDiffInterceptor(tc.secondInterceptorErrors)
			interceptors := append(
				first.Invocations(),
				second.Invocations()...,
			)

			f := newMockInnerCustomizeDiffFunc(tc.innerFuncError)

			handler := interceptedCustomizeDiffHandler(contextFunc, interceptors, f.Call)

			ctx := t.Context()
			err := handler(ctx, nil, client)

			if diff := cmp.Diff(err, tc.expectedError, cmp.Comparer(func(x, y error) bool {
				return x.Error() == y.Error()
			})); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
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

type mockCustomizeDiffInterceptor struct {
	errors map[when]error
	called []when
}

func newMockCustomizeDiffInterceptor(errors map[when]error) *mockCustomizeDiffInterceptor {
	if errors == nil {
		errors = make(map[when]error)
	}
	return &mockCustomizeDiffInterceptor{
		errors: errors,
	}
}

func (m *mockCustomizeDiffInterceptor) Invocations() interceptorInvocations {
	return interceptorInvocations{
		{
			why:         CustomizeDiff,
			when:        Before | After | OnError | Finally,
			interceptor: m,
		},
	}
}

func (m *mockCustomizeDiffInterceptor) run(_ context.Context, opts customizeDiffInterceptorOptions) error {
	m.called = append(m.called, opts.when)
	return m.errors[opts.when]
}

type mockInnerCustomizeDiffFunc struct {
	err   error
	count int
}

func newMockInnerCustomizeDiffFunc(err error) mockInnerCustomizeDiffFunc {
	return mockInnerCustomizeDiffFunc{
		err: err,
	}
}

func (m *mockInnerCustomizeDiffFunc) Call(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	m.count++
	return m.err
}

func TestInterceptedImportHandler(t *testing.T) {
	t.Parallel()

	client := mockClient{
		accountID: "123456789012",
		region:    "us-west-2", //lintignore:AWSAT003
	}

	contextFunc := func(ctx context.Context, _ getAttributeFunc, _ getProviderMetaFunc, meta any) (context.Context, error) {
		return ctx, nil
	}

	testcases := map[string]struct {
		firstInterceptorErrors  map[when]error
		secondInterceptorErrors map[when]error
		innerFuncError          error
		expectedFirstCalls      []when
		expectedSecondCalls     []when
		expectedInnerCalls      int
		expectedError           error
	}{
		"First has Before error": {
			firstInterceptorErrors: map[when]error{
				Before: errors.New("First interceptor Before error"),
			},
			expectedFirstCalls: []when{Before},
			expectedInnerCalls: 0,
			expectedError:      errors.New("First interceptor Before error"),
		},

		"Second has Before error": {
			secondInterceptorErrors: map[when]error{
				Before: errors.New("Second interceptor Before error"),
			},
			expectedFirstCalls:  []when{Before},
			expectedSecondCalls: []when{Before},
			expectedInnerCalls:  0,
			expectedError:       errors.New("Second interceptor Before error"),
		},

		"Inner has error": {
			innerFuncError:      errors.New("Inner function error"),
			expectedFirstCalls:  []when{Before, OnError, Finally},
			expectedSecondCalls: []when{Before, OnError, Finally},
			expectedInnerCalls:  1,
			expectedError: errors.Join(
				errors.New("Inner function error"),
			),
		},

		"All have errors": {
			firstInterceptorErrors: map[when]error{
				OnError: errors.New("First interceptor OnError error"),
				Finally: errors.New("First interceptor Finally error"),
			},
			secondInterceptorErrors: map[when]error{
				OnError: errors.New("Second interceptor OnError error"),
				Finally: errors.New("Second interceptor Finally error"),
			},
			innerFuncError:      errors.New("Inner function error"),
			expectedFirstCalls:  []when{Before, OnError, Finally},
			expectedSecondCalls: []when{Before, OnError, Finally},
			expectedInnerCalls:  1,
			expectedError: errors.Join(
				errors.New("Inner function error"),
				errors.New("Second interceptor OnError error"),
				errors.New("First interceptor OnError error"),
				errors.New("Second interceptor Finally error"),
				errors.New("First interceptor Finally error"),
			),
		},

		"Handlers have errors": {
			firstInterceptorErrors: map[when]error{
				After:   errors.New("First interceptor After error"),
				Finally: errors.New("First interceptor Finally error"),
			},
			secondInterceptorErrors: map[when]error{
				After:   errors.New("Second interceptor After error"),
				Finally: errors.New("Second interceptor Finally error"),
			},
			expectedFirstCalls:  []when{Before, After, Finally},
			expectedSecondCalls: []when{Before, After, Finally},
			expectedInnerCalls:  1,
			expectedError: errors.Join(
				errors.New("Second interceptor After error"),
				errors.New("First interceptor After error"),
				errors.New("Second interceptor Finally error"),
				errors.New("First interceptor Finally error"),
			),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			first := newMockImportInterceptor(tc.firstInterceptorErrors)
			second := newMockImportInterceptor(tc.secondInterceptorErrors)
			interceptors := append(
				first.Invocations(),
				second.Invocations()...,
			)

			f := newMockInnerImportFunc(tc.innerFuncError)

			handler := interceptedImportHandler(contextFunc, interceptors, f.Call)

			ctx := t.Context()
			_, err := handler(ctx, nil, client)

			if diff := cmp.Diff(err, tc.expectedError, cmp.Comparer(func(x, y error) bool {
				return x.Error() == y.Error()
			})); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
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

type mockImportInterceptor struct {
	errors map[when]error
	called []when
}

func newMockImportInterceptor(errors map[when]error) *mockImportInterceptor {
	if errors == nil {
		errors = make(map[when]error)
	}
	return &mockImportInterceptor{
		errors: errors,
	}
}

func (m *mockImportInterceptor) Invocations() interceptorInvocations {
	return interceptorInvocations{
		{
			why:         Import,
			when:        Before | After | OnError | Finally,
			interceptor: m,
		},
	}
}

func (m *mockImportInterceptor) run(_ context.Context, opts importInterceptorOptions) error {
	m.called = append(m.called, opts.when)
	return m.errors[opts.when]
}

type mockInnerImportFunc struct {
	err   error
	count int
}

func newMockInnerImportFunc(err error) mockInnerImportFunc {
	return mockInnerImportFunc{
		err: err,
	}
}

func (m *mockInnerImportFunc) Call(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	m.count++
	return nil, m.err
}
