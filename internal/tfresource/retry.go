package tfresource

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Retryable is a function that is used to decide if a function's error is retryable or not.
// The error argument can be `nil`.
// If the error is retryable, returns a bool value of `true` and an error (not necessarily the error passed as the argument).
// If the error is not retryable, returns a bool value of `false` and either no error (success state) or an error (not necessarily the error passed as the argument).
type Retryable func(error) (bool, error)

// RetryWhenContext retries the function `f` when the error it returns satisfies `predicate`.
// `f` is retried until `timeout` expires.
func RetryWhenContext(ctx context.Context, timeout time.Duration, f func() (interface{}, error), retryable Retryable) (interface{}, error) {
	var output interface{}

	err := resource.Retry(timeout, func() *resource.RetryError { // nosemgrep: helper-schema-resource-Retry-without-TimeoutError-check
		var err error
		var retry bool

		output, err = f()
		retry, err = retryable(err)

		if retry {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
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

// RetryWhen retries the function `f` when the error it returns satisfies `predicate`.
// `f` is retried until `timeout` expires.
func RetryWhen(timeout time.Duration, f func() (interface{}, error), retryable Retryable) (interface{}, error) {
	return RetryWhenContext(context.Background(), timeout, f, retryable)
}

// RetryWhenAWSErrCodeEqualsContext retries the specified function when it returns one of the specified AWS error code.
func RetryWhenAWSErrCodeEqualsContext(ctx context.Context, timeout time.Duration, f func() (interface{}, error), codes ...string) (interface{}, error) { // nosemgrep:aws-in-func-name
	return RetryWhenContext(ctx, timeout, f, func(err error) (bool, error) {
		if tfawserr.ErrCodeEquals(err, codes...) {
			return true, err
		}

		return false, err
	})
}

// RetryWhenAWSErrCodeEquals retries the specified function when it returns one of the specified AWS error code.
func RetryWhenAWSErrCodeEquals(timeout time.Duration, f func() (interface{}, error), codes ...string) (interface{}, error) { // nosemgrep:aws-in-func-name
	return RetryWhenAWSErrCodeEqualsContext(context.Background(), timeout, f, codes...)
}

// RetryWhenAWSErrMessageContainsContext retries the specified function when it returns an AWS error containing the specified message.
func RetryWhenAWSErrMessageContainsContext(ctx context.Context, timeout time.Duration, f func() (interface{}, error), code, message string) (interface{}, error) { // nosemgrep:aws-in-func-name
	return RetryWhenContext(ctx, timeout, f, func(err error) (bool, error) {
		if tfawserr.ErrMessageContains(err, code, message) {
			return true, err
		}

		return false, err
	})
}

// RetryWhenAWSErrMessageContains retries the specified function when it returns an AWS error containing the specified message.
func RetryWhenAWSErrMessageContains(timeout time.Duration, f func() (interface{}, error), code, message string) (interface{}, error) { // nosemgrep:aws-in-func-name
	return RetryWhenAWSErrMessageContainsContext(context.Background(), timeout, f, code, message)
}

var resourceFoundError = errors.New(`found resource`)

// RetryUntilNotFoundContext retries the specified function until it returns a resource.NotFoundError.
func RetryUntilNotFoundContext(ctx context.Context, timeout time.Duration, f func() (interface{}, error)) (interface{}, error) {
	return RetryWhenContext(ctx, timeout, f, func(err error) (bool, error) {
		if NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return true, resourceFoundError
	})
}

// RetryUntilNotFound retries the specified function until it returns a resource.NotFoundError.
func RetryUntilNotFound(timeout time.Duration, f func() (interface{}, error)) (interface{}, error) {
	return RetryUntilNotFoundContext(context.Background(), timeout, f)
}

// RetryWhenNotFoundContext retries the specified function when it returns a resource.NotFoundError.
func RetryWhenNotFoundContext(ctx context.Context, timeout time.Duration, f func() (interface{}, error)) (interface{}, error) {
	return RetryWhenContext(ctx, timeout, f, func(err error) (bool, error) {
		if NotFound(err) {
			return true, err
		}

		return false, err
	})
}

// RetryWhenNotFound retries the specified function when it returns a resource.NotFoundError.
func RetryWhenNotFound(timeout time.Duration, f func() (interface{}, error)) (interface{}, error) {
	return RetryWhenNotFoundContext(context.Background(), timeout, f)
}

// RetryWhenNewResourceNotFoundContext retries the specified function when it returns a resource.NotFoundError and `isNewResource` is true.
func RetryWhenNewResourceNotFoundContext(ctx context.Context, timeout time.Duration, f func() (interface{}, error), isNewResource bool) (interface{}, error) {
	return RetryWhenContext(ctx, timeout, f, func(err error) (bool, error) {
		if isNewResource && NotFound(err) {
			return true, err
		}

		return false, err
	})
}

// RetryWhenNewResourceNotFound retries the specified function when it returns a resource.NotFoundError and `isNewResource` is true.
func RetryWhenNewResourceNotFound(timeout time.Duration, f func() (interface{}, error), isNewResource bool) (interface{}, error) {
	return RetryWhenNewResourceNotFoundContext(context.Background(), timeout, f, isNewResource)
}

// RetryConfigContext allows configuration of StateChangeConf's various time arguments.
// This is especially useful for AWS services that are prone to throttling, such as Route53, where
// the default durations cause problems. To not use a StateChangeConf argument and revert to the
// default, pass in a zero value (i.e., 0*time.Second).
func RetryConfigContext(ctx context.Context, delay time.Duration, delayRand time.Duration, minTimeout time.Duration, pollInterval time.Duration, timeout time.Duration, f resource.RetryFunc) error {
	// These are used to pull the error out of the function; need a mutex to
	// avoid a data race.
	var resultErr error
	var resultErrMu sync.Mutex

	c := &resource.StateChangeConf{
		Pending: []string{"retryableerror"},
		Target:  []string{"success"},
		Timeout: timeout,
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

	if delay.Milliseconds() > 0 {
		c.Delay = delay
	}

	if delayRand.Milliseconds() > 0 {
		// Hitting the API at exactly the same time on each iteration of the retry is more likely to
		// cause Throttling problems. We introduce randomness in order to help AWS be happier.
		rand.Seed(time.Now().UTC().UnixNano())

		c.Delay = time.Duration(rand.Int63n(delayRand.Milliseconds())) * time.Millisecond
	}

	if minTimeout.Milliseconds() > 0 {
		c.MinTimeout = minTimeout
	}

	if pollInterval.Milliseconds() > 0 {
		c.PollInterval = pollInterval
	}

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
