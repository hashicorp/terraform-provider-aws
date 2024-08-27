// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
)

// resourceNotFoundExceptionResourceIDEquals returns true if the error matches all these conditions:
//   - err is of type networkmanager.ResourceNotFoundException
//   - ResourceNotFoundException.ResourceId equals resourceID
func resourceNotFoundExceptionResourceIDEquals(err error, resourceID string) bool {
	var resourceNotFoundException *awstypes.ResourceNotFoundException

	if errors.As(err, &resourceNotFoundException) && aws.ToString(resourceNotFoundException.ResourceId) == resourceID {
		return true
	}

	return false
}

// validationExceptionFieldsMessageContains returns true if the error matches all these conditions:
//   - err is of type awstypes.ValidationException
//   - ValidationException.Reason equals reason
//   - ValidationException.Fields.Message contains message
func validationExceptionFieldsMessageContains(err error, reason awstypes.ValidationExceptionReason, message string) bool {
	var validationException *awstypes.ValidationException

	if errors.As(err, &validationException) && validationException.Reason == reason {
		for _, v := range validationException.Fields {
			if strings.Contains(aws.ToString(v.Message), message) {
				return true
			}
		}
	}

	return false
}
