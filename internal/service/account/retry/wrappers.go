package retry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// If retries an operation until the timeout elapses or predicate indicates otherwise.
func If[T any](ctx context.Context, timeout time.Duration, op func() (T, error), predicate func(T, error) (bool, error)) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for r := Begin(); r.Continue(ctx); {
		t, err := op()
		if retry, err := predicate(t, err); !retry {
			return t, err
		}
	}

	var t T
	return t, ctx.Err()
}

// UntilFoundN retries an operation if it returns a retry.NotFoundError.
func UntilFoundN[T any](ctx context.Context, timeout time.Duration, op func() (T, error), continuousTargetOccurence int) (T, error) {
	if continuousTargetOccurence < 1 {
		continuousTargetOccurence = 1
	}

	targetOccurence := 0

	t, err := If(ctx, timeout, op, func(_ T, err error) (bool, error) {
		if err == nil {
			targetOccurence++

			if continuousTargetOccurence == targetOccurence {
				return false, nil
			}

			return true, nil
		}

		if tfresource.NotFound(err) {
			targetOccurence = 0

			return true, err
		}

		return false, err
	})

	return t, err
}

// UntilNotFound retries an operation until it returns a retry.NotFoundError.
func UntilNotFound[T any](ctx context.Context, timeout time.Duration, op func() (T, error)) (T, error) {
	t, err := If(ctx, timeout, op, func(_ T, err error) (bool, error) {
		if err == nil {
			return true, nil
		}

		if tfresource.NotFound(err) {
			return false, nil
		}

		return false, err
	})

	if errors.Is(err, context.DeadlineExceeded) {
		return t, fmt.Errorf("found resource: %w", err)
	}

	return t, err
}
