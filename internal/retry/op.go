// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package retry

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/backoff"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type opFunc[T any] func(context.Context) (T, error)
type predicateFunc[T any] func(T, error) (bool, error)
type runFunc[T any] func(context.Context, time.Duration, ...backoff.Option) (T, error)

// Op returns a new wrapper on top of the specified function.
func Op[T any](op func(context.Context) (T, error)) opFunc[T] {
	return op
}

// UntilFoundN retries an operation if it returns a retry.NotFoundError.
func (o opFunc[T]) UntilFoundN(continuousTargetOccurence int) runFunc[T] {
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

	return o.If(predicate)
}

func (o opFunc[T]) UntilNotFound() runFunc[T] {
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
		t, err := o.If(predicate)(ctx, timeout, opts...)

		if TimedOut(err) {
			return t, ErrFoundResource
		}

		return t, err
	}
}

func (o opFunc[T]) If(predicate predicateFunc[T]) runFunc[T] {
	// The default predicate short-circuits a retry loop if the operation returns any error.
	if predicate == nil {
		predicate = func(_ T, err error) (bool, error) {
			return err != nil, err
		}
	}

	return func(ctx context.Context, timeout time.Duration, opts ...backoff.Option) (T, error) {
		// We explicitly don't set a deadline on the context here to maintain compatibility
		// with the Plugin SDKv2 implementation. A parent context may have set a deadline.
		var l *backoff.Loop
		for l = backoff.NewLoopWithOptions(timeout, opts...); l.Continue(ctx); {
			t, err := o(ctx)

			if retry, err := predicate(t, err); !retry {
				return t, err
			}
		}

		var err error
		if l.Remaining() == 0 {
			err = inttypes.ErrDeadlineExceeded
		} else {
			err = context.Cause(ctx)
		}

		if errors.Is(err, inttypes.ErrDeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			err = &TimeoutError{
				LastError:     err,
				LastState:     "retryableerror",
				Timeout:       timeout,
				ExpectedState: []string{"success"},
			}
		}

		var zero T
		return zero, err
	}
}
