// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

var (
	errCodeResourceNotFound = (*awstypes.ResourceNotFound)(nil).ErrorCode()
)

func partialFailuresError(apiObjects []awstypes.PartialFailure) error {
	var errs []error

	for _, apiObject := range apiObjects {
		errs = append(errs, fmt.Errorf("%s %s: %w", aws.ToString(apiObject.ExceptionType), aws.ToString(apiObject.FailureResource), partialFailureError(apiObject)))
	}

	return errors.Join(errs...)
}

func partialFailureError(apiObject awstypes.PartialFailure) error {
	return errs.APIError(aws.ToString(apiObject.FailureCode), aws.ToString(apiObject.FailureDescription))
}
