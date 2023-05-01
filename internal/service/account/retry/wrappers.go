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

// defaultPredicate short-circuits a retry loop if the operation returns any error.
func defaultPredicate[T any](t T, err error) (bool, error) {
	return err != nil, err
}

type operation[T any] struct {
	op        Op[T]
	predicate Predicate[T]
}

// Operation returns a new wrapper on top of the specified function.
func Operation[T any](op OpFunc[T]) operation[T] {
	return operation[T]{op: op, predicate: PredicateFunc[T](defaultPredicate[T])}
}

func (o operation[T]) If(predicate PredicateFunc[T]) operation[T] {
	return operation[T]{op: o.op, predicate: predicate}
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

		if tfresource.NotFound(err) {
			targetOccurence = 0

			return true, err
		}

		return false, err
	}

	return operation[T]{op: o.op, predicate: PredicateFunc[T](predicate)}
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
	return t, ctx.Err()
}

type transformErrorOperation[T any] struct {
	inner     operation[T]
	transform func(error) error
}

func (o transformErrorOperation[T]) Run(ctx context.Context, timeout time.Duration) (T, error) {
	t, err := o.inner.Run(ctx, timeout)

	return t, o.transform(err)
}

func (o operation[T]) UntilNotFound() transformErrorOperation[T] {
	predicate := func(_ T, err error) (bool, error) {
		if err == nil {
			return true, nil
		}

		if tfresource.NotFound(err) {
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

	return transformErrorOperation[T]{inner: operation[T]{op: o.op, predicate: PredicateFunc[T](predicate)}, transform: transform}
}
