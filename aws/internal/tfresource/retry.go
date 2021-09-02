package tfresource

import (
	"time"

	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

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
