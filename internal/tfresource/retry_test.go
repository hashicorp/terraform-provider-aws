// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package tfresource_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

//nolint:tparallel
func TestRetryWhenAWSErrCodeEquals(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	ctx := t.Context()
	testCases := []struct {
		Name        string
		F           func(context.Context) (any, error)
		ExpectError bool
	}{
		{
			Name: "no error",
			F: func(context.Context) (any, error) {
				return nil, nil
			},
		},
		{
			Name: "non-retryable other error",
			F: func(context.Context) (any, error) {
				return nil, errors.New("TestCode")
			},
			ExpectError: true,
		},
	}

	for _, testCase := range testCases { //nolint:paralleltest
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 5*time.Second, testCase.F, "TestCode1", "TestCode2")

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

//nolint:tparallel
func TestRetryWhenAWSErrMessageContains(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	ctx := t.Context()
	testCases := []struct {
		Name        string
		F           func(context.Context) (any, error)
		ExpectError bool
	}{
		{
			Name: "no error",
			F: func(context.Context) (any, error) {
				return nil, nil
			},
		},
		{
			Name: "non-retryable other error",
			F: func(context.Context) (any, error) {
				return nil, errors.New("TestCode")
			},
			ExpectError: true,
		},
	}

	for _, testCase := range testCases { //nolint:paralleltest
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 5*time.Second, testCase.F, "TestCode1", "TestMessage1")

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

//nolint:tparallel
func TestRetryWhenNewResourceNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var retryCount int32
	testCases := []struct {
		Name        string
		F           func(context.Context) (any, error)
		NewResource bool
		ExpectError bool
	}{
		{
			Name: "no error",
			F: func(context.Context) (any, error) {
				return nil, nil
			},
		},
		{
			Name: "no error new resource",
			F: func(context.Context) (any, error) {
				return nil, nil
			},
			NewResource: true,
		},
		{
			Name: "non-retryable other error",
			F: func(context.Context) (any, error) {
				return nil, errors.New("TestCode")
			},
			ExpectError: true,
		},
		{
			Name: "non-retryable other error new resource",
			F: func(context.Context) (any, error) {
				return nil, errors.New("TestCode")
			},
			NewResource: true,
			ExpectError: true,
		},
		{
			Name: "retryable NotFoundError not new resource",
			F: func(context.Context) (any, error) {
				return nil, &retry.NotFoundError{}
			},
			ExpectError: true,
		},
		{
			Name: "retryable NotFoundError new resource timeout",
			F: func(context.Context) (any, error) {
				return nil, &retry.NotFoundError{}
			},
			NewResource: true,
			ExpectError: true,
		},
		{
			Name: "retryable NotFoundError success new resource",
			F: func(context.Context) (any, error) {
				if atomic.CompareAndSwapInt32(&retryCount, 0, 1) {
					return nil, &retry.NotFoundError{}
				}

				return nil, nil
			},
			NewResource: true,
		},
	}

	for _, testCase := range testCases { //nolint:paralleltest
		t.Run(testCase.Name, func(t *testing.T) {
			retryCount = 0

			_, err := tfresource.RetryWhenNewResourceNotFound(ctx, 5*time.Second, testCase.F, testCase.NewResource)

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

//nolint:tparallel
func TestRetryWhenNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var retryCount int32
	testCases := []struct {
		Name        string
		F           func(context.Context) (any, error)
		ExpectError bool
	}{
		{
			Name: "no error",
			F: func(ctx context.Context) (any, error) {
				return nil, nil
			},
		},
		{
			Name: "non-retryable other error",
			F: func(ctx context.Context) (any, error) {
				return nil, errors.New("TestCode")
			},
			ExpectError: true,
		},
		{
			Name: "retryable NotFoundError timeout",
			F: func(ctx context.Context) (any, error) {
				return nil, &retry.NotFoundError{}
			},
			ExpectError: true,
		},
		{
			Name: "retryable NotFoundError success",
			F: func(ctx context.Context) (any, error) {
				if atomic.CompareAndSwapInt32(&retryCount, 0, 1) {
					return nil, &retry.NotFoundError{}
				}

				return nil, nil
			},
		},
	}

	for _, testCase := range testCases { //nolint:paralleltest
		t.Run(testCase.Name, func(t *testing.T) {
			retryCount = 0

			_, err := tfresource.RetryWhenNotFound(ctx, 5*time.Second, testCase.F)

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

//nolint:tparallel
func TestRetryUntilEqual(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var retryCount int32
	target := 42
	testCases := []struct {
		Name        string
		F           func(context.Context) (int, error)
		ExpectError bool
	}{
		{
			Name: "return error",
			F: func(context.Context) (int, error) {
				return 0, errors.New("TestCode")
			},
			ExpectError: true,
		},
		{
			Name: "equal immediately",
			F: func(context.Context) (int, error) {
				return target, nil
			},
		},
		{
			Name: "equal eventually",
			F: func(context.Context) (int, error) {
				if atomic.CompareAndSwapInt32(&retryCount, 0, 1) {
					return target, nil
				}

				return 0, nil
			},
		},
		{
			Name: "equal never",
			F: func(context.Context) (int, error) {
				return 0, nil
			},
			ExpectError: true,
		},
	}

	for _, testCase := range testCases { //nolint:paralleltest
		t.Run(testCase.Name, func(t *testing.T) {
			retryCount = 0

			_, err := tfresource.RetryUntilEqual(ctx, 5*time.Second, target, testCase.F)

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

//nolint:tparallel
func TestRetryUntilNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var retryCount int32
	testCases := []struct {
		Name        string
		F           func(context.Context) (any, error)
		ExpectError bool
	}{
		{
			Name: "no error",
			F: func(context.Context) (any, error) {
				return nil, nil
			},
			ExpectError: true,
		},
		{
			Name: "other error",
			F: func(context.Context) (any, error) {
				return nil, errors.New("TestCode")
			},
			ExpectError: true,
		},
		{
			Name: "NotFoundError",
			F: func(context.Context) (any, error) {
				return nil, &retry.NotFoundError{}
			},
		},
		{
			Name: "retryable NotFoundError",
			F: func(context.Context) (any, error) {
				if atomic.CompareAndSwapInt32(&retryCount, 0, 1) {
					return nil, nil
				}

				return nil, &retry.NotFoundError{}
			},
		},
	}

	for _, testCase := range testCases { //nolint:paralleltest
		t.Run(testCase.Name, func(t *testing.T) {
			retryCount = 0

			_, err := tfresource.RetryUntilNotFound(ctx, 5*time.Second, testCase.F)

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

func TestRetryContext_nil(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var expected error
	f := func(context.Context) *tfresource.RetryError {
		return nil
	}

	errCh := make(chan error)
	go func() {
		errCh <- tfresource.Retry(ctx, 1*time.Second, f)
	}()

	select {
	case err := <-errCh:
		if err != expected { //nolint:errorlint // We are actually comparing equality
			t.Fatalf("bad: %#v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func TestRetryContext_nonRetryableError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	expected := fmt.Errorf("nope")
	f := func(context.Context) *tfresource.RetryError {
		return tfresource.NonRetryableError(expected)
	}

	errCh := make(chan error)
	go func() {
		errCh <- tfresource.Retry(ctx, 1*time.Second, f)
	}()

	select {
	case err := <-errCh:
		if err != expected { //nolint:errorlint // We are actually comparing equality
			t.Fatalf("bad: %#v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func TestRetryContext_retryableError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	expected := fmt.Errorf("nope")
	f := func(context.Context) *tfresource.RetryError {
		return tfresource.RetryableError(expected)
	}

	errCh := make(chan error)
	go func() {
		errCh <- tfresource.Retry(ctx, 1*time.Second, f)
	}()

	select {
	case err := <-errCh:
		if err != expected { //nolint:errorlint // We are actually comparing equality
			t.Fatalf("bad: %#v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func TestOptionsApply(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		options  tfresource.Options
		expected retry.StateChangeConf
	}{
		"Nothing": {
			options:  tfresource.Options{},
			expected: retry.StateChangeConf{},
		},
		"Delay": {
			options: tfresource.Options{
				Delay: 1 * time.Minute,
			},
			expected: retry.StateChangeConf{
				Delay: 1 * time.Minute,
			},
		},
		"MinPollInterval": {
			options: tfresource.Options{
				MinPollInterval: 1 * time.Minute,
			},
			expected: retry.StateChangeConf{
				MinTimeout: 1 * time.Minute,
			},
		},
		"PollInterval": {
			options: tfresource.Options{
				PollInterval: 1 * time.Minute,
			},
			expected: retry.StateChangeConf{
				PollInterval: 1 * time.Minute,
			},
		},
		"NotFoundChecks": {
			options: tfresource.Options{
				NotFoundChecks: 10,
			},
			expected: retry.StateChangeConf{
				NotFoundChecks: 10,
			},
		},
		"ContinuousTargetOccurence": {
			options: tfresource.Options{
				ContinuousTargetOccurence: 3,
			},
			expected: retry.StateChangeConf{
				ContinuousTargetOccurence: 3,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			conf := retry.StateChangeConf{}

			testCase.options.Apply(&conf)

			if a, e := conf.Delay, testCase.expected.Delay; a != e {
				t.Errorf("Delay: expected %s, got %s", e, a)
			}
			if a, e := conf.MinTimeout, testCase.expected.MinTimeout; a != e {
				t.Errorf("MinTimeout: expected %s, got %s", e, a)
			}
			if a, e := conf.PollInterval, testCase.expected.PollInterval; a != e {
				t.Errorf("PollInterval: expected %s, got %s", e, a)
			}
			if a, e := conf.NotFoundChecks, testCase.expected.NotFoundChecks; a != e {
				t.Errorf("NotFoundChecks: expected %d, got %d", e, a)
			}
			if a, e := conf.ContinuousTargetOccurence, testCase.expected.ContinuousTargetOccurence; a != e {
				t.Errorf("ContinuousTargetOccurence: expected %d, got %d", e, a)
			}
		})
	}
}

// Distinct fake error types used by the OneOf*ErrorMessageContains tests below — each
// implements errs.ErrorWithErrorMessage so the generic dispatch can disambiguate by type.
type retryTestErrA struct{ msg string }

func (e *retryTestErrA) Error() string        { return e.msg }
func (e *retryTestErrA) ErrorMessage() string { return e.msg }

type retryTestErrB struct{ msg string }

func (e *retryTestErrB) Error() string        { return e.msg }
func (e *retryTestErrB) ErrorMessage() string { return e.msg }

type retryTestErrC struct{ msg string }

func (e *retryTestErrC) Error() string        { return e.msg }
func (e *retryTestErrC) ErrorMessage() string { return e.msg }

func TestRetryWhenIsOneOf2ErrorMessageContains(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	t.Run("no error", func(t *testing.T) {
		t.Parallel()
		_, err := tfresource.RetryWhenIsOneOf2ErrorMessageContains[any, *retryTestErrA, *retryTestErrB](ctx, 5*time.Second,
			func(context.Context) (any, error) { return nil, nil },
			"needle-a", "needle-b",
		)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("non-retryable other error", func(t *testing.T) {
		t.Parallel()
		_, err := tfresource.RetryWhenIsOneOf2ErrorMessageContains[any, *retryTestErrA, *retryTestErrB](ctx, 5*time.Second,
			func(context.Context) (any, error) { return nil, errors.New("unrelated") },
			"needle-a", "needle-b",
		)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("retries first error type then succeeds", func(t *testing.T) {
		t.Parallel()
		var count int32
		_, err := tfresource.RetryWhenIsOneOf2ErrorMessageContains[any, *retryTestErrA, *retryTestErrB](ctx, 5*time.Second,
			func(context.Context) (any, error) {
				if atomic.AddInt32(&count, 1) == 1 {
					return nil, &retryTestErrA{msg: "wraps needle-a inside"}
				}
				return nil, nil
			},
			"needle-a", "needle-b",
		)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if got := atomic.LoadInt32(&count); got < 2 {
			t.Errorf("expected at least 2 attempts, got %d", got)
		}
	})

	t.Run("retries second error type then succeeds", func(t *testing.T) {
		t.Parallel()
		var count int32
		_, err := tfresource.RetryWhenIsOneOf2ErrorMessageContains[any, *retryTestErrA, *retryTestErrB](ctx, 5*time.Second,
			func(context.Context) (any, error) {
				if atomic.AddInt32(&count, 1) == 1 {
					return nil, &retryTestErrB{msg: "needle-b appears"}
				}
				return nil, nil
			},
			"needle-a", "needle-b",
		)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("type matches but needle does not", func(t *testing.T) {
		t.Parallel()
		_, err := tfresource.RetryWhenIsOneOf2ErrorMessageContains[any, *retryTestErrA, *retryTestErrB](ctx, 5*time.Second,
			func(context.Context) (any, error) { return nil, &retryTestErrA{msg: "wrong message"} },
			"needle-a", "needle-b",
		)
		if err == nil {
			t.Fatal("expected error: type matches but needle does not")
		}
	})

	t.Run("first type with second needle does not retry", func(t *testing.T) {
		// Pairing must be by index: E1 retries on needle1, E2 on needle2.
		// An error of type E1 carrying needle2's text must NOT trigger a retry.
		t.Parallel()
		_, err := tfresource.RetryWhenIsOneOf2ErrorMessageContains[any, *retryTestErrA, *retryTestErrB](ctx, 5*time.Second,
			func(context.Context) (any, error) { return nil, &retryTestErrA{msg: "needle-b but type A"} },
			"needle-a", "needle-b",
		)
		if err == nil {
			t.Fatal("expected error: needle is paired with its type by index, not interchangeable")
		}
	})
}

func TestRetryWhenIsOneOf3ErrorMessageContains(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	t.Run("no error", func(t *testing.T) {
		t.Parallel()
		_, err := tfresource.RetryWhenIsOneOf3ErrorMessageContains[any, *retryTestErrA, *retryTestErrB, *retryTestErrC](ctx, 5*time.Second,
			func(context.Context) (any, error) { return nil, nil },
			"needle-a", "needle-b", "needle-c",
		)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("non-retryable other error", func(t *testing.T) {
		t.Parallel()
		_, err := tfresource.RetryWhenIsOneOf3ErrorMessageContains[any, *retryTestErrA, *retryTestErrB, *retryTestErrC](ctx, 5*time.Second,
			func(context.Context) (any, error) { return nil, errors.New("unrelated") },
			"needle-a", "needle-b", "needle-c",
		)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("retries each of the three types then succeeds", func(t *testing.T) {
		t.Parallel()
		// Cycles through type A, B, C on consecutive failures, then returns success.
		var count int32
		_, err := tfresource.RetryWhenIsOneOf3ErrorMessageContains[any, *retryTestErrA, *retryTestErrB, *retryTestErrC](ctx, 5*time.Second,
			func(context.Context) (any, error) {
				switch atomic.AddInt32(&count, 1) {
				case 1:
					return nil, &retryTestErrA{msg: "before needle-a after"}
				case 2:
					return nil, &retryTestErrB{msg: "needle-b mid"}
				case 3:
					return nil, &retryTestErrC{msg: "trailing needle-c"}
				default:
					return nil, nil
				}
			},
			"needle-a", "needle-b", "needle-c",
		)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if got := atomic.LoadInt32(&count); got < 4 {
			t.Errorf("expected at least 4 attempts (A, B, C, success), got %d", got)
		}
	})

	t.Run("type matches but needle does not", func(t *testing.T) {
		t.Parallel()
		_, err := tfresource.RetryWhenIsOneOf3ErrorMessageContains[any, *retryTestErrA, *retryTestErrB, *retryTestErrC](ctx, 5*time.Second,
			func(context.Context) (any, error) { return nil, &retryTestErrB{msg: "wrong message"} },
			"needle-a", "needle-b", "needle-c",
		)
		if err == nil {
			t.Fatal("expected error: type matches but needle does not")
		}
	})

	t.Run("third type with first needle does not retry", func(t *testing.T) {
		t.Parallel()
		_, err := tfresource.RetryWhenIsOneOf3ErrorMessageContains[any, *retryTestErrA, *retryTestErrB, *retryTestErrC](ctx, 5*time.Second,
			func(context.Context) (any, error) { return nil, &retryTestErrC{msg: "needle-a but type C"} },
			"needle-a", "needle-b", "needle-c",
		)
		if err == nil {
			t.Fatal("expected error: needle is paired with its type by index, not interchangeable")
		}
	})
}
