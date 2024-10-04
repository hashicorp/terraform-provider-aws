// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package errs

import (
	smithy "github.com/aws/smithy-go"
)

// APIError returns a new error suitable for checking via aws-sdk-go-base/tfawserr.
func APIError[T ~string](code T, message string) smithy.APIError {
	return &smithy.GenericAPIError{
		Code:    string(code),
		Message: message,
	}
}
