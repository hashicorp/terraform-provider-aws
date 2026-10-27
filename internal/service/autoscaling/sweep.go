// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_autoscaling_group", sweepGroups)
	awsv2.Register("aws_launch_configuration", sweepLaunchConfigurations, "aws_autoscaling_group")
}

func sweepGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AutoScalingClient(ctx)
	var input autoscaling.DescribeAutoScalingGroupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := autoscaling.NewDescribeAutoScalingGroupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.AutoScalingGroups {
			asgName := aws.ToString(v.AutoScalingGroupName)

			// "ValidationError: You can't force delete this Auto Scaling group because the TerminateHookAbandon retention trigger is set to retain. Change the retention trigger to terminate and try again".
			if v.InstanceLifecyclePolicy != nil && v.InstanceLifecyclePolicy.RetentionTriggers != nil && v.InstanceLifecyclePolicy.RetentionTriggers.TerminateHookAbandon == awstypes.RetentionActionRetain {
				input := autoscaling.UpdateAutoScalingGroupInput{
					AutoScalingGroupName: aws.String(asgName),
					InstanceLifecyclePolicy: &awstypes.InstanceLifecyclePolicy{
						RetentionTriggers: &awstypes.RetentionTriggers{
							TerminateHookAbandon: awstypes.RetentionActionTerminate,
						},
					},
				}
				_, err := conn.UpdateAutoScalingGroup(ctx, &input)
				if err != nil {
					return nil, err
				}
			}

			r := resourceGroup()
			d := r.Data(nil)
			d.SetId(asgName)
			d.Set(names.AttrForceDelete, true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepLaunchConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AutoScalingClient(ctx)
	var input autoscaling.DescribeLaunchConfigurationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := autoscaling.NewDescribeLaunchConfigurationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.LaunchConfigurations {
			r := resourceLaunchConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.LaunchConfigurationName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
