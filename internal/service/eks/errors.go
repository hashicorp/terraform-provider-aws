// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go/aws/awserr"
	multierror "github.com/hashicorp/go-multierror"
)

func AddonIssueError(apiObject *types.AddonIssue) error {
	if apiObject == nil {
		return nil
	}

	return awserr.New(string(apiObject.Code), aws.ToString(apiObject.Message), nil)
}

func AddonIssuesError(apiObjects []types.AddonIssue) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if &apiObject == nil {
			continue
		}

		err := AddonIssueError(&apiObject)

		if err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: %w", strings.Join(apiObject.ResourceIds, ", "), err))
		}
	}

	return errors.ErrorOrNil()
}

func ErrorDetailError(apiObject types.ErrorDetail) error {
	if &apiObject == nil {
		return nil
	}

	return awserr.New(string(apiObject.ErrorCode), aws.ToString(apiObject.ErrorMessage), nil)
}

func ErrorDetailsError(apiObjects []types.ErrorDetail) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if &apiObject == nil {
			continue
		}

		err := ErrorDetailError(apiObject)

		if err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: %w", strings.Join(apiObject.ResourceIds, ", "), err))
		}
	}

	return errors.ErrorOrNil()
}

func IssueError(apiObject *types.Issue) error {
	if apiObject == nil {
		return nil
	}

	return awserr.New(string(apiObject.Code), aws.ToString(apiObject.Message), nil)
}

func IssuesError(apiObjects []types.Issue) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if &apiObject == nil {
			continue
		}

		err := IssueError(&apiObject)

		if err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: %w", strings.Join(apiObject.ResourceIds, ", "), err))
		}
	}

	return errors.ErrorOrNil()
}
