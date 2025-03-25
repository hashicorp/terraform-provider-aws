// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// resourceNotFoundExceptionResourceIDEquals returns true if the error matches all these conditions:
//   - err is of type networkmanager.ResourceNotFoundException
//   - ResourceNotFoundException.ResourceId equals resourceID
func resourceNotFoundExceptionResourceIDEquals(err error, resourceID string) bool {
	if err, ok := errs.As[*awstypes.ResourceNotFoundException](err); ok && aws.ToString(err.ResourceId) == resourceID {
		return true
	}

	return false
}

// validationExceptionMessageContains returns true if the error matches all these conditions:
//   - err is of type networkmanager.ValidationException
//   - ValidationException.Reason equals reason
//   - ValidationException.Message contains message
func validationExceptionMessageContains(err error, reason awstypes.ValidationExceptionReason, message string) bool {
	if err, ok := errs.As[*awstypes.ValidationException](err); ok && err.Reason == reason {
		return strings.Contains(aws.ToString(err.Message), message)
	}

	return false
}

// validationExceptionFieldsMessageContains returns true if the error matches all these conditions:
//   - err is of type awstypes.ValidationException
//   - ValidationException.Reason equals reason
//   - ValidationException.Fields.Message contains message
func validationExceptionFieldsMessageContains(err error, reason awstypes.ValidationExceptionReason, message string) bool {
	if err, ok := errs.As[*awstypes.ValidationException](err); ok && err.Reason == reason && slices.ContainsFunc(err.Fields, func(v awstypes.ValidationExceptionField) bool {
		return strings.Contains(aws.ToString(v.Message), message)
	}) {
		return true
	}

	return false
}

func attachmentError(apiObject *awstypes.AttachmentError) error {
	if apiObject == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", aws.ToString(apiObject.ResourceArn), fmt.Errorf("%s: %s", apiObject.Code, aws.ToString(apiObject.Message)))
}

func attachmentsError(apiObjects []awstypes.AttachmentError) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if err := attachmentError(&apiObject); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", aws.ToString(apiObject.ResourceArn), err))
		}
	}

	return errors.Join(errs...)
}

func connectPeerError(apiObject *awstypes.ConnectPeerError) error {
	if apiObject == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", aws.ToString(apiObject.ResourceArn), fmt.Errorf("%s: %s", apiObject.Code, aws.ToString(apiObject.Message)))
}

func connectPeersError(apiObjects []awstypes.ConnectPeerError) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if err := connectPeerError(&apiObject); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", aws.ToString(apiObject.ResourceArn), err))
		}
	}

	return errors.Join(errs...)
}

func peeringError(apiObject *awstypes.PeeringError) error {
	if apiObject == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", aws.ToString(apiObject.ResourceArn), fmt.Errorf("%s: %s", apiObject.Code, aws.ToString(apiObject.Message)))
}

func peeringsError(apiObjects []awstypes.PeeringError) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if err := peeringError(&apiObject); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", aws.ToString(apiObject.ResourceArn), err))
		}
	}

	return errors.Join(errs...)
}
