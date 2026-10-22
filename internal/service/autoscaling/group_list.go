// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_autoscaling_group")
func newGroupResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceGroup(), framework.SDKListSpec[awstypes.AutoScalingGroup, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.AutoScalingGroup, error], diag.Diagnostics) {
			var input autoscaling.DescribeAutoScalingGroupsInput

			return listGroups(ctx, c.AutoScalingClient(ctx), &input), nil
		},
		ID: func(g awstypes.AutoScalingGroup, _ framework.RegionQuery) string {
			return aws.ToString(g.AutoScalingGroupName)
		},
		Flatten: resourceGroupFlatten,
	})
}

func listGroups(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeAutoScalingGroupsInput) iter.Seq2[awstypes.AutoScalingGroup, error] {
	return func(yield func(awstypes.AutoScalingGroup, error) bool) {
		pages := autoscaling.NewDescribeAutoScalingGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.AutoScalingGroup{}, fmt.Errorf("listing Auto Scaling Groups: %w", err))
				return
			}

			for _, g := range page.AutoScalingGroups {
				if !yield(g, nil) {
					return
				}
			}
		}
	}
}
