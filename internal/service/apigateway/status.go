// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

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
