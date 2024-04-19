// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	appconfigtypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	smithy "github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func TestAddIsErrorRetryables(t *testing.T) {
	t.Parallel()

	f := func(err error) aws.Ternary {
		if errs.Contains(err, "testing") {
			return aws.TrueTernary
		}
		return aws.UnknownTernary
	}
	testCases := []struct {
		name     string
		err      error
		f        retry.IsErrorRetryableFunc
		expected bool
	}{
		{
			name: "no error",
			f:    f,
		},
		{
			name: "non-retryable",
			err:  errors.New(`this is not retryable`),
			f:    f,
		},
		{
			name:     "retryable",
			err:      errors.New(`this is testing`),
			f:        f,
			expected: true,
		},
		{
			// https://github.com/hashicorp/terraform-provider-aws/issues/36975.
			name: "appconfig ConflictException",
			err: &smithy.OperationError{
				ServiceID:     "AppConfig",
				OperationName: "StartDeployment",
				Err: &awshttp.ResponseError{
					ResponseError: &smithyhttp.ResponseError{
						Response: &smithyhttp.Response{
							Response: &http.Response{
								StatusCode: http.StatusConflict,
							},
						},
						Err: &appconfigtypes.ConflictException{
							Message: aws.String("Deployment number 1 already exists"),
						},
					},
					RequestID: "43e844da-818b-458e-aae2-553960ccc4d6",
				},
			},
			f: func(err error) aws.Ternary {
				if err, ok := errs.As[*smithy.OperationError](err); ok {
					switch err.OperationName {
					case "StartDeployment":
						if errs.IsA[*appconfigtypes.ConflictException](err) {
							return aws.TrueTernary
						}
					}
				}
				return aws.UnknownTernary
			},
			expected: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := AddIsErrorRetryables(retry.NewStandard(), testCase.f).IsErrorRetryable(testCase.err)
			if got, want := got, testCase.expected; got != want {
				t.Errorf("IsErrorRetryable(%q) = %v, want %v", testCase.err, got, want)
			}
		})
	}
}
