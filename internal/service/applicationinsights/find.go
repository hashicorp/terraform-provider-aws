// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationinsights

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationinsights"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindApplicationByName(ctx context.Context, conn *applicationinsights.ApplicationInsights, name string) (*applicationinsights.ApplicationInfo, error) {
	input := applicationinsights.DescribeApplicationInput{
		ResourceGroupName: aws.String(name),
	}

	output, err := conn.DescribeApplicationWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, applicationinsights.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApplicationInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ApplicationInfo, nil
}
