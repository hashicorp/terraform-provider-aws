// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindLambdaFunctionAssociationByARNWithContext(ctx context.Context, conn *connect.Connect, instanceID string, functionArn string) (string, error) {
	var result string

	input := &connect.ListLambdaFunctionsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListLambdaFunctionsMaxResults),
	}

	err := conn.ListLambdaFunctionsPagesWithContext(ctx, input, func(page *connect.ListLambdaFunctionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cf := range page.LambdaFunctions {
			if cf == nil {
				continue
			}

			if aws.StringValue(cf) == functionArn {
				result = functionArn
				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		return "", &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return "", err
	}

	if result == "" {
		return "", tfresource.NewEmptyResultError(input)
	}

	return result, nil
}
