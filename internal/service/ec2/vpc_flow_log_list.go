// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_flow_log")
func newFlowLogResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceFlowLog(), framework.SDKListSpec[awstypes.FlowLog, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.FlowLog, error], diag.Diagnostics) {
			var input ec2.DescribeFlowLogsInput

			return listFlowLogsPages(ctx, c.EC2Client(ctx), &input), nil
		},
		ID: func(f awstypes.FlowLog, _ framework.RegionQuery) string {
			return aws.ToString(f.FlowLogId)
		},
		Flatten: resourceFlowLogFlatten,
	})
}

func listFlowLogsPages(ctx context.Context, conn *ec2.Client, input *ec2.DescribeFlowLogsInput) iter.Seq2[awstypes.FlowLog, error] {
	return func(yield func(awstypes.FlowLog, error) bool) {
		pages := ec2.NewDescribeFlowLogsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.FlowLog{}, fmt.Errorf("listing Flow Logs: %w", err))
				return
			}

			for _, item := range page.FlowLogs {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
