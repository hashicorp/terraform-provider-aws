// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package retry

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/backoff"
)

type opFunc[T any] func(context.Context) (T, error)
type predicateFunc[T any] func(T, error) (bool, error)
type runFunc[T any] func(context.Context, time.Duration, ...backoff.Option) (T, error)

// Op returns a new wrapper on top of the specified function.
func Op[T any](op func(context.Context) (T, error)) opFunc[T] {
	return op
}

// UntilFoundN retries an operation if it returns a retry.NotFoundError.
func (op opFunc[T]) UntilFoundN(continuousTargetOccurence int) runFunc[T] {
	if continuousTargetOccurence < 1 {
		continuousTargetOccurence = 1
	}

	targetOccurence := 0

	predicate := func(_ T, err error) (bool, error) {
		if err == nil {
			targetOccurence++

			if continuousTargetOccurence == targetOccurence {
				return false, nil
			}

			return true, nil
		}

		if NotFound(err) { // nosemgrep:ci.semgrep.errors.notfound-without-err-checks
			targetOccurence = 0

			return true, err
		}

		return false, err
	}

	return op.If(predicate)
}

func (op opFunc[T]) UntilNotFound() runFunc[T] {
	predicate := func(_ T, err error) (bool, error) {
		if err == nil {
			return true, nil
		}

		if NotFound(err) { // nosemgrep:ci.semgrep.errors.notfound-without-err-checks
			return false, nil
		}

		return false, err
	}

	return func(ctx context.Context, timeout time.Duration, opts ...backoff.Option) (T, error) {
		t, err := op.If(predicate)(ctx, timeout, opts...)

		if TimedOut(err) {
			return t, ErrFoundResource
		}

		return t, err
	}
}

func (op opFunc[T]) If(predicate predicateFunc[T]) runFunc[T] {
	// The default predicate short-circuits a retry loop if the operation returns any error.
	if predicate == nil {
		predicate = func(_ T, err error) (bool, error) {
			return err != nil, err
		}
	}

	return func(ctx context.Context, timeout time.Duration, opts ...backoff.Option) (T, error) {
		// We explicitly don't set a deadline on the context here to maintain compatibility
		// with the Plugin SDKv2 implementation. A parent context may have set a deadline.
		var (
			l   *backoff.Loop
			t   T
			err error
		)
		for l = backoff.NewLoopWithOptions(timeout, opts...); l.Continue(ctx); {
			t, err = op(ctx)

			var retry bool
			if retry, err = predicate(t, err); !retry {
				return t, err
			}
		}

		if err == nil {
			if l.Remaining() == 0 || errors.Is(err, context.Cause(ctx)) {
				err = &TimeoutError{
					// LastError must be nil for `TimedOut` to return true.
					// LastError:     err,
					LastState:     "retryableerror",
					Timeout:       timeout,
					ExpectedState: []string{"success"},
				}
			}
		}

		return t, err
	}
}
