// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package retry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

type Op[T any] interface {
	Invoke(context.Context) (T, error)
}

type OpFunc[T any] func(context.Context) (T, error)

func (f OpFunc[T]) Invoke(ctx context.Context) (T, error) {
	return f(ctx)
}

type Predicate[T any] interface {
	Invoke(T, error) (bool, error)
}

type PredicateFunc[T any] func(T, error) (bool, error)

func (f PredicateFunc[T]) Invoke(t T, err error) (bool, error) {
	return f(t, err)
}

type operation[T any] struct {
	op                Op[T]
	predicate         Predicate[T]
	transformRunError func(error) error
}

// Operation returns a new wrapper on top of the specified function.
func Operation[T any](op OpFunc[T]) operation[T] {
	return operation[T]{
		op: op,
		// The default predicate short-circuits a retry loop if the operation returns any error.
		predicate: PredicateFunc[T](func(t T, err error) (bool, error) {
			return err != nil, err
		}),
		// The default error transformer does nothing.
		transformRunError: func(err error) error { return err },
	}
}

func (o operation[T]) withPredicate(predicate Predicate[T]) operation[T] {
	return operation[T]{op: o.op, predicate: predicate, transformRunError: o.transformRunError}
}

func (o operation[T]) withTransformRunError(f func(error) error) operation[T] {
	return operation[T]{op: o.op, predicate: o.predicate, transformRunError: f}
}

func (o operation[T]) If(predicate PredicateFunc[T]) operation[T] {
	return o.withPredicate(predicate)
}

// UntilFoundN retries an operation if it returns a retry.NotFoundError.
func (o operation[T]) UntilFoundN(continuousTargetOccurence int) operation[T] {
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

		if tfresource.NotFound(err) { // nosemgrep:ci.semgrep.errors.notfound-without-err-checks
			targetOccurence = 0

			return true, err
		}

		return false, err
	}

	return o.If(predicate)
}

func (o operation[T]) UntilNotFound() operation[T] {
	predicate := func(_ T, err error) (bool, error) {
		if err == nil {
			return true, nil
		}

		if tfresource.NotFound(err) { // nosemgrep:ci.semgrep.errors.notfound-without-err-checks
			return false, nil
		}

		return false, err
	}

	transform := func(err error) error {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("found resource: %w", err)
		}

		return err
	}

	return o.If(predicate).withTransformRunError(transform)
}

// Run retries an operation until the timeout elapses or predicate indicates otherwise.
func (o operation[T]) Run(ctx context.Context, timeout time.Duration) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for r := Begin(); r.Continue(ctx); {
		t, err := o.op.Invoke(ctx)

		if retry, err := o.predicate.Invoke(t, err); !retry {
			return t, err
		}
	}

	var t T
	return t, o.transformRunError(ctx.Err())
}
