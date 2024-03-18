// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func vpcLinkStatus(ctx context.Context, conn *apigateway.Client, vpcLinkId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetVpcLink(ctx, &apigateway.GetVpcLinkInput{
			VpcLinkId: aws.String(vpcLinkId),
		})
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		// Error messages can also be contained in the response with FAILED status
		if output.Status == awstypes.VpcLinkStatusFailed {
			return output, string(awstypes.VpcLinkStatusFailed), fmt.Errorf("%s: %s", string(awstypes.VpcLinkStatusFailed), aws.ToString(output.StatusMessage))
		}

		return output, string(output.Status), nil
	}
}

func stageCacheStatus(ctx context.Context, conn *apigateway.Client, restApiId, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStageByTwoPartKey(ctx, conn, restApiId, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, string(output.CacheClusterStatus), nil
	}
}
