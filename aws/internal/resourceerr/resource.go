package resourceerr

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// NotFoundError returns true if the error is of type resource.NotFoundError.
func NotFoundError(err error) bool {
	_, ok := err.(*resource.NotFoundError)
	return ok
}

// TimeoutError returns true if the error matches all these conditions:
//  * err is of type resource.TimeoutError
//  * TimeoutError.LastError is nil
func TimeoutError(err error) bool {
	timeoutErr, ok := err.(*resource.TimeoutError)
	return ok && timeoutErr.LastError == nil
}
