// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

const (
	errCodeValidationError = "ValidationError"
)

func stackSetOperationError(apiObjects []*cloudformation.StackSetOperationResultSummary) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		errs = append(errs, fmt.Errorf("Account (%s), Region (%s), %s: %s",
			aws.StringValue(apiObject.Account),
			aws.StringValue(apiObject.Region),
			aws.StringValue(apiObject.Status),
			aws.StringValue(apiObject.StatusReason),
		))
	}

	return errors.Join(errs...)
}
