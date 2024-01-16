// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func vpcLinkStatus(ctx context.Context, conn *apigateway.APIGateway, vpcLinkId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetVpcLinkWithContext(ctx, &apigateway.GetVpcLinkInput{
			VpcLinkId: aws.String(vpcLinkId),
		})
		if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		// Error messages can also be contained in the response with FAILED status
		if aws.StringValue(output.Status) == apigateway.VpcLinkStatusFailed {
			return output, apigateway.VpcLinkStatusFailed, fmt.Errorf("%s: %s", apigateway.VpcLinkStatusFailed, aws.StringValue(output.StatusMessage))
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func stageCacheStatus(ctx context.Context, conn *apigateway.APIGateway, restApiId, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStageByTwoPartKey(ctx, conn, restApiId, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.CacheClusterStatus), nil
	}
}
