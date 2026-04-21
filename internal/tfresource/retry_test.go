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
