// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_cloudwatch_log_group")
func newLogGroupResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceGroup(), framework.SDKListSpec[awstypes.LogGroup, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.LogGroup, error], diag.Diagnostics) {
			var input cloudwatchlogs.DescribeLogGroupsInput

			return listLogGroups(ctx, c.LogsClient(ctx), &input), nil
		},
		ID: func(lg awstypes.LogGroup, _ framework.RegionQuery) string {
			return aws.ToString(lg.LogGroupName)
		},
		Flatten: resourceGroupFlatten,
	})
}

func listLogGroups(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) iter.Seq2[awstypes.LogGroup, error] {
	return tfiter.ConcatValuesWithError(listLogGroupPages(ctx, conn, input, optFns...))
}
