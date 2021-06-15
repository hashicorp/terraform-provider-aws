package tfresource

import (
	"time"

	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// RetryUntilFound retries the specified function until the underlying resource is found.
// The function returns a resource.NotFoundError to indicate that the underlying resource does not exist.
// If the retries time out, the function is called one last time.
func RetryUntilFound(timeout time.Duration, f func() (interface{}, error)) (interface{}, error) {
	var output interface{}

	err := resource.Retry(timeout, func() *resource.RetryError {
		var err error

		output, err = f()

		if NotFound(err) {
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

	return output, err
}

// RetryWhenAwsErrCodeEquals retries the specified function when it returns one of the specified AWS error code.
func RetryWhenAwsErrCodeEquals(timeout time.Duration, f func() (interface{}, error), codes ...string) (interface{}, error) {
	var output interface{}

	err := resource.Retry(timeout, func() *resource.RetryError {
		var err error

		output, err = f()

		for _, code := range codes {
			if tfawserr.ErrCodeEquals(err, code) {
				return resource.RetryableError(err)
			}
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
