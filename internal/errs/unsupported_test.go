// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package errs_test

import (
	"testing"

	"github.com/aws/smithy-go"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func TestIsUnsupportedOperationInPartitionError(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		partition string
		err       error
		want      bool
	}{
		"AccessDenied authorization failure": {
			partition: endpoints.AwsUsGovPartitionID,
			err: &smithy.GenericAPIError{
				Code:    "AccessDenied",
				Message: "User: arn:aws:iam::123456789012:user/example is not authorized to perform: iam:TagRole",
			},
			want: false,
		},
		"AccessDenied unsupported-API signal": {
			partition: endpoints.AwsUsGovPartitionID,
			err: &smithy.GenericAPIError{
				Code:    "AccessDenied",
				Message: "Tagging is not supported in this partition",
			},
			want: true,
		},
		"AccessDeniedException authorization failure": {
			partition: endpoints.AwsUsGovPartitionID,
			err: &smithy.GenericAPIError{
				Code:    "AccessDeniedException",
				Message: "User: arn:aws:iam::123456789012:user/example is not authorized to perform: iam:TagRole",
			},
			want: false,
		},
		"AccessDeniedException unsupported-API signal": {
			partition: endpoints.AwsUsGovPartitionID,
			err: &smithy.GenericAPIError{
				Code:    "AccessDeniedException",
				Message: "Tagging is not supported in this partition",
			},
			want: true,
		},
		"AuthorizationError authorization failure": {
			partition: endpoints.AwsUsGovPartitionID,
			err: &smithy.GenericAPIError{
				Code:    "AuthorizationError",
				Message: "User: arn:aws:iam::123456789012:user/example is not authorized to perform: sns:Subscribe",
			},
			want: false,
		},
		"AuthorizationError unsupported-API signal": {
			partition: endpoints.AwsUsGovPartitionID,
			err: &smithy.GenericAPIError{
				Code:    "AuthorizationError",
				Message: "The operation is not permitted in this partition",
			},
			want: true,
		},
		"InvalidAction": {
			partition: endpoints.AwsIsoPartitionID,
			err: &smithy.GenericAPIError{
				Code:    "InvalidAction",
				Message: "The action is not valid for this service",
			},
			want: true,
		},
		"UnknownOperationException": {
			partition: endpoints.AwsIsoPartitionID,
			err: &smithy.GenericAPIError{
				Code:    "UnknownOperationException",
				Message: "The operation is unknown",
			},
			want: true,
		},
		"ValidationError not support tagging": {
			partition: endpoints.AwsIsoPartitionID,
			err: &smithy.GenericAPIError{
				Code:    "ValidationError",
				Message: "The service does not support tagging",
			},
			want: true,
		},
		"commercial aws partition": {
			partition: endpoints.AwsPartitionID,
			err: &smithy.GenericAPIError{
				Code:    "AccessDenied",
				Message: "User: arn:aws:iam::123456789012:user/example is not authorized to perform: iam:TagRole",
			},
			want: false,
		},
		"nil error": {
			partition: endpoints.AwsUsGovPartitionID,
			err:       nil,
			want:      false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := errs.IsUnsupportedOperationInPartitionError(testcase.partition, testcase.err); got != testcase.want {
				t.Errorf("IsUnsupportedOperationInPartitionError(%q, %v) = %v, want %v", testcase.partition, testcase.err, got, testcase.want)
			}
		})
	}
}
