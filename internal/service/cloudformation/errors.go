// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	multierror "github.com/hashicorp/go-multierror"
)

const (
	errCodeValidationError = "ValidationError"
)

func StackSetOperationError(apiObjects []*cloudformation.StackSetOperationResultSummary) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		errors = multierror.Append(errors, fmt.Errorf("Account (%s) Region (%s) Status (%s) Status Reason: %s",
			aws.StringValue(apiObject.Account),
			aws.StringValue(apiObject.Region),
			aws.StringValue(apiObject.Status),
			aws.StringValue(apiObject.StatusReason),
		))
	}

	return errors.ErrorOrNil()
}
