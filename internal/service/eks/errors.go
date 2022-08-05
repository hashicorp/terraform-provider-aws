package eks

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/eks"
	multierror "github.com/hashicorp/go-multierror"
)

func AddonIssueError(apiObject *eks.AddonIssue) error {
	if apiObject == nil {
		return nil
	}

	return awserr.New(aws.StringValue(apiObject.Code), aws.StringValue(apiObject.Message), nil)
}

func AddonIssuesError(apiObjects []*eks.AddonIssue) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		err := AddonIssueError(apiObject)

		if err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: %w", strings.Join(aws.StringValueSlice(apiObject.ResourceIds), ", "), err))
		}
	}

	return errors.ErrorOrNil()
}

func ErrorDetailError(apiObject *eks.ErrorDetail) error {
	if apiObject == nil {
		return nil
	}

	return awserr.New(aws.StringValue(apiObject.ErrorCode), aws.StringValue(apiObject.ErrorMessage), nil)
}

func ErrorDetailsError(apiObjects []*eks.ErrorDetail) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		err := ErrorDetailError(apiObject)

		if err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: %w", strings.Join(aws.StringValueSlice(apiObject.ResourceIds), ", "), err))
		}
	}

	return errors.ErrorOrNil()
}

func IssueError(apiObject *eks.Issue) error {
	if apiObject == nil {
		return nil
	}

	return awserr.New(aws.StringValue(apiObject.Code), aws.StringValue(apiObject.Message), nil)
}

func IssuesError(apiObjects []*eks.Issue) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		err := IssueError(apiObject)

		if err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: %w", strings.Join(aws.StringValueSlice(apiObject.ResourceIds), ", "), err))
		}
	}

	return errors.ErrorOrNil()
}
