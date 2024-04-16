// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

const (
	errCodeValidationError = "ValidationError"
)

func stackSetOperationError(apiObjects []awstypes.StackSetOperationResultSummary) error {
	var errs []error

	for _, apiObject := range apiObjects {
		errs = append(errs, fmt.Errorf("Account (%s), Region (%s), %s: %s",
			aws.ToString(apiObject.Account),
			aws.ToString(apiObject.Region),
			string(apiObject.Status),
			aws.ToString(apiObject.StatusReason),
		))
	}

	return errors.Join(errs...)
}
