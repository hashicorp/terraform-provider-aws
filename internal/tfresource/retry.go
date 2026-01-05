// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package tfresource

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/internal/backoff"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

var ErrFoundResource = retry.ErrFoundResource

// Retryable is a function that is used to decide if a function's error is retryable or not.
// The error argument can be `nil`.
// If the error is retryable, returns a bool value of `true` and an error (not necessarily the error passed as the argument).
// If the error is not retryable, returns a bool value of `false` and either no error (success state) or an error (not necessarily the error passed as the argument).
type Retryable func(error) (bool, error)

// RetryWhen retries the function `f` when the error it returns satisfies `retryable`.
// `f` is retried until `timeout` expires.
func RetryWhen[T any](ctx context.Context, timeout time.Duration, f func(context.Context) (T, error), retryable Retryable) (T, error) {
	return retryWhen(ctx, timeout, f, retryable)
}

// RetryWhenAWSErrCodeEquals retries the specified function when it returns one of the specified AWS error codes.
func RetryWhenAWSErrCodeEquals[T any](ctx context.Context, timeout time.Duration, f func(context.Context) (T, error), codes ...string) (T, error) { // nosemgrep:ci.aws-in-func-name
	return retryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if tfawserr.ErrCodeEquals(err, codes...) {
			return true, err
		}

		return false, err
	})
}

// RetryWhenAWSErrCodeContains retries the specified function when it returns an AWS error containing the specified code.
func RetryWhenAWSErrCodeContains[T any](ctx context.Context, timeout time.Duration, f func(context.Context) (T, error), code string) (T, error) { // nosemgrep:ci.aws-in-func-name
	return retryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if tfawserr.ErrCodeContains(err, code) {
			return true, err
		}

		return false, err
	})
}

// RetryWhenAWSErrMessageContains retries the specified function when it returns an AWS error containing the specified message.
func RetryWhenAWSErrMessageContains[T any](ctx context.Context, timeout time.Duration, f func(context.Context) (T, error), code, message string) (T, error) { // nosemgrep:ci.aws-in-func-name
	return retryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if tfawserr.ErrMessageContains(err, code, message) {
			return true, err
		}

		return false, err
	})
}

func RetryWhenIsA[T any, E error](ctx context.Context, timeout time.Duration, f func(context.Context) (T, error)) (T, error) {
	return retryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if errs.IsA[E](err) {
			return true, err
		}

		return false, err
	})
}

func RetryWhenIsOneOf2[T any, E1, E2 error](ctx context.Context, timeout time.Duration, f func(context.Context) (T, error)) (T, error) {
	return retryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if errs.IsA[E1](err) || errs.IsA[E2](err) {
			return true, err
		}

		return false, err
	})
}

func RetryWhenIsOneOf3[T any, E1, E2, E3 error](ctx context.Context, timeout time.Duration, f func(context.Context) (T, error)) (T, error) {
	return retryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if errs.IsA[E1](err) || errs.IsA[E2](err) || errs.IsA[E3](err) {
			return true, err
		}

		return false, err
	})
}

func RetryWhenIsOneOf4[T any, E1, E2, E3, E4 error](ctx context.Context, timeout time.Duration, f func(context.Context) (T, error)) (T, error) {
	return retryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if errs.IsA[E1](err) || errs.IsA[E2](err) || errs.IsA[E3](err) || errs.IsA[E4](err) {
			return true, err
		}

		return false, err
	})
}

func RetryWhenIsAErrorMessageContains[T any, E errs.ErrorWithErrorMessage](ctx context.Context, timeout time.Duration, f func(context.Context) (T, error), needle string) (T, error) {
	return retryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if errs.IsAErrorMessageContains[E](err, needle) {
			return true, err
		}

		return false, err
	})
}

// RetryUntilEqual retries the specified function until it returns a value equal to `target`.
func RetryUntilEqual[T comparable](ctx context.Context, timeout time.Duration, target T, f func(context.Context) (T, error), opts ...backoff.Option) (T, error) {
	t, err := retry.Op(f).If(func(t T, err error) (bool, error) {
		if err != nil {
			return false, err
		}
		return t != target, nil
	})(ctx, timeout, opts...)

	if TimedOut(err) {
		err = fmt.Errorf("output = %v, want %v", t, target)
	}

	return t, err
}

// RetryUntilNotFound retries the specified function until it returns a retry.NotFoundError.
func RetryUntilNotFound(ctx context.Context, timeout time.Duration, f func(context.Context) (any, error)) (any, error) {
	return retry.Op(f).UntilNotFound()(ctx, timeout)
}

// RetryWhenNotFound retries the specified function when it returns a retry.NotFoundError.
func RetryWhenNotFound[T any](ctx context.Context, timeout time.Duration, f func(context.Context) (T, error)) (T, error) {
	return retry.Op(f).UntilFoundN(1)(ctx, timeout)
}

// RetryWhenNewResourceNotFound retries the specified function when it returns a retry.NotFoundError and `isNewResource` is true.
func RetryWhenNewResourceNotFound[T any](ctx context.Context, timeout time.Duration, f func(context.Context) (T, error), isNewResource bool) (T, error) {
	return retry.Op(f).If(func(_ T, err error) (bool, error) {
		if isNewResource && retry.NotFound(err) {
			return true, err
		}

		return false, err
	})(ctx, timeout)
}

type Options struct {
	Delay                     time.Duration // Wait this time before starting checks
	MinPollInterval           time.Duration // Smallest time to wait before refreshes (MinTimeout in retry.StateChangeConf)
	PollInterval              time.Duration // Override MinPollInterval/backoff and only poll this often
	NotFoundChecks            int           // Number of times to allow not found (nil result from Refresh)
	ContinuousTargetOccurence int           // Number of times the Target state has to occur continuously
}

func (o Options) Apply(c *retry.StateChangeConf) {
	if o.Delay > 0 {
		c.Delay = o.Delay
	}

	if o.MinPollInterval > 0 {
		c.MinTimeout = o.MinPollInterval
	}

	if o.PollInterval > 0 {
		c.PollInterval = o.PollInterval
	}

	if o.NotFoundChecks > 0 {
		c.NotFoundChecks = o.NotFoundChecks
	}

	if o.ContinuousTargetOccurence > 0 {
		c.ContinuousTargetOccurence = o.ContinuousTargetOccurence
	}
}

type OptionsFunc func(*Options)

func WithDelay(delay time.Duration) OptionsFunc {
	return func(o *Options) {
		o.Delay = delay
	}
}

// WithDelayRand sets the delay to a value between 0s and the passed duration
func WithDelayRand(delayRand time.Duration) OptionsFunc {
	return func(o *Options) {
		o.Delay = time.Duration(rand.Int63n(delayRand.Milliseconds())) * time.Millisecond
	}
}

func WithMinPollInterval(minPollInterval time.Duration) OptionsFunc {
	return func(o *Options) {
		o.MinPollInterval = minPollInterval
	}
}

func WithPollInterval(pollInterval time.Duration) OptionsFunc {
	return func(o *Options) {
		o.PollInterval = pollInterval
	}
}

func WithNotFoundChecks(notFoundChecks int) OptionsFunc {
	return func(o *Options) {
		o.NotFoundChecks = notFoundChecks
	}
}

func WithContinuousTargetOccurence(continuousTargetOccurence int) OptionsFunc {
	return func(o *Options) {
		o.ContinuousTargetOccurence = continuousTargetOccurence
	}
}

// Retry allows configuration of StateChangeConf's various time arguments.
// This is especially useful for AWS services that are prone to throttling, such as Route53, where
// the default durations cause problems.
func Retry(ctx context.Context, timeout time.Duration, f func(context.Context) *RetryError, optFns ...OptionsFunc) error {
	options := Options{
		MinPollInterval: 500 * time.Millisecond, //nolint:mnd // 500ms is the Plugin SDKv2 default
	}
	for _, fn := range optFns {
		fn(&options)
	}

	_, err := retryWhen(ctx, timeout,
		func(ctx context.Context) (any, error) {
			return nil, f(ctx)
		},
		func(err error) (bool, error) {
			if err, ok := errs.As[*RetryError](err); ok {
				if err != nil {
					return err.isRetryable, err.err
				}
				return false, nil
			}

			return false, err
		},
		backoff.WithDelay(backoff.SDKv2HelperRetryCompatibleDelay(options.Delay, options.PollInterval, options.MinPollInterval)),
	)

	return err
}

func retryWhen[T any](ctx context.Context, timeout time.Duration, f func(context.Context) (T, error), retryable Retryable, opts ...backoff.Option) (T, error) {
	return retry.Op(f).If(func(_ T, err error) (bool, error) {
		return retryable(err)
	})(ctx, timeout, opts...)
}
