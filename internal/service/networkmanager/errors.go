// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
)

// resourceNotFoundExceptionResourceIDEquals returns true if the error matches all these conditions:
//   - err is of type networkmanager.ResourceNotFoundException
//   - ResourceNotFoundException.ResourceId equals resourceID
func resourceNotFoundExceptionResourceIDEquals(err error, resourceID string) bool {
	var resourceNotFoundException *networkmanager.ResourceNotFoundException

	if errors.As(err, &resourceNotFoundException) && aws.StringValue(resourceNotFoundException.ResourceId) == resourceID {
		return true
	}

	return false
}

// validationExceptionFieldsMessageContains returns true if the error matches all these conditions:
//   - err is of type networkmanager.ValidationException
//   - ValidationException.Reason equals reason
//   - ValidationException.Fields.Message contains message
func validationExceptionFieldsMessageContains(err error, reason string, message string) bool {
	var validationException *networkmanager.ValidationException

	if errors.As(err, &validationException) && aws.StringValue(validationException.Reason) == reason {
		for _, v := range validationException.Fields {
			if strings.Contains(aws.StringValue(v.Message), message) {
				return true
			}
		}
	}

	return false
}
