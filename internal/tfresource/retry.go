// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfresource

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	tfawserr_sdkv2 "github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// Retryable is a function that is used to decide if a function's error is retryable or not.
// The error argument can be `nil`.
// If the error is retryable, returns a bool value of `true` and an error (not necessarily the error passed as the argument).
// If the error is not retryable, returns a bool value of `false` and either no error (success state) or an error (not necessarily the error passed as the argument).
type Retryable func(error) (bool, error)

// RetryWhen retries the function `f` when the error it returns satisfies `retryable`.
// `f` is retried until `timeout` expires.
func RetryWhen(ctx context.Context, timeout time.Duration, f func() (interface{}, error), retryable Retryable) (interface{}, error) {
	var output interface{}

	err := Retry(ctx, timeout, func() *retry.RetryError {
		var err error
		var again bool

		output, err = f()
		again, err = retryable(err)

		if again {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if TimedOut(err) {
		output, err = f()
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

// RetryWhenAWSErrCodeEquals retries the specified function when it returns one of the specified AWS error code.
func RetryWhenAWSErrCodeEquals(ctx context.Context, timeout time.Duration, f func() (interface{}, error), codes ...string) (interface{}, error) { // nosemgrep:ci.aws-in-func-name
	return RetryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if tfawserr.ErrCodeEquals(err, codes...) || tfawserr_sdkv2.ErrCodeEquals(err, codes...) {
			return true, err
		}

		return false, err
	})
}

// RetryWhenAWSErrMessageContains retries the specified function when it returns an AWS error containing the specified message.
func RetryWhenAWSErrMessageContains(ctx context.Context, timeout time.Duration, f func() (interface{}, error), code, message string) (interface{}, error) { // nosemgrep:ci.aws-in-func-name
	return RetryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if tfawserr.ErrMessageContains(err, code, message) || tfawserr_sdkv2.ErrMessageContains(err, code, message) {
			return true, err
		}

		return false, err
	})
}

func RetryWhenIsA[T error](ctx context.Context, timeout time.Duration, f func() (interface{}, error)) (interface{}, error) {
	return RetryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if errs.IsA[T](err) {
			return true, err
		}

		return false, err
	})
}

func RetryWhenIsAErrorMessageContains[T errs.ErrorWithErrorMessage](ctx context.Context, timeout time.Duration, f func() (interface{}, error), needle string) (interface{}, error) {
	return RetryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if errs.IsAErrorMessageContains[T](err, needle) {
			return true, err
		}

		return false, err
	})
}

var ErrFoundResource = errors.New(`found resource`)

// RetryUntilNotFound retries the specified function until it returns a retry.NotFoundError.
func RetryUntilNotFound(ctx context.Context, timeout time.Duration, f func() (interface{}, error)) (interface{}, error) {
	return RetryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return true, ErrFoundResource
	})
}

// RetryWhenNotFound retries the specified function when it returns a retry.NotFoundError.
func RetryWhenNotFound(ctx context.Context, timeout time.Duration, f func() (interface{}, error)) (interface{}, error) {
	return RetryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if NotFound(err) {
			return true, err
		}

		return false, err
	})
}

// RetryWhenNewResourceNotFound retries the specified function when it returns a retry.NotFoundError and `isNewResource` is true.
func RetryWhenNewResourceNotFound(ctx context.Context, timeout time.Duration, f func() (interface{}, error), isNewResource bool) (interface{}, error) {
	return RetryWhen(ctx, timeout, f, func(err error) (bool, error) {
		if isNewResource && NotFound(err) {
			return true, err
		}

		return false, err
	})
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
func Retry(ctx context.Context, timeout time.Duration, f retry.RetryFunc, optFns ...OptionsFunc) error {
	// These are used to pull the error out of the function; need a mutex to
	// avoid a data race.
	var resultErr error
	var resultErrMu sync.Mutex

	options := Options{}
	for _, fn := range optFns {
		fn(&options)
	}

	c := &retry.StateChangeConf{
		Pending:    []string{"retryableerror"},
		Target:     []string{"success"},
		Timeout:    timeout,
		MinTimeout: 500 * time.Millisecond,
		Refresh: func() (interface{}, string, error) {
			rerr := f()

			resultErrMu.Lock()
			defer resultErrMu.Unlock()

			if rerr == nil {
				resultErr = nil
				return 42, "success", nil
			}

			resultErr = rerr.Err

			if rerr.Retryable {
				return 42, "retryableerror", nil
			}

			return nil, "quit", rerr.Err
		},
	}

	options.Apply(c)

	_, waitErr := c.WaitForStateContext(ctx)

	// Need to acquire the lock here to be able to avoid race using resultErr as
	// the return value
	resultErrMu.Lock()
	defer resultErrMu.Unlock()

	// resultErr may be nil because the wait timed out and resultErr was never
	// set; this is still an error
	if resultErr == nil {
		return waitErr
	}
	// resultErr takes precedence over waitErr if both are set because it is
	// more likely to be useful
	return resultErr
}

type deadline time.Time

func NewDeadline(duration time.Duration) deadline {
	return deadline(time.Now().Add(duration))
}

func (d deadline) Remaining() time.Duration {
	if v := time.Until(time.Time(d)); v < 0 {
		return 0
	} else {
		return v
	}
}
