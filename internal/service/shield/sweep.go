// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_shield_drt_access_log_bucket_association", &resource.Sweeper{
		Name: "aws_shield_drt_access_log_bucket_association",
		F:    sweepDRTAccessLogBucketAssociations,
	})

	resource.AddTestSweepers("aws_shield_drt_access_role_arn_association", &resource.Sweeper{
		Name: "aws_shield_drt_access_role_arn_association",
		F:    sweepDRTAccessRoleARNAssociations,
		Dependencies: []string{
			"aws_shield_drt_access_log_bucket_association",
		},
	})

	resource.AddTestSweepers("aws_shield_proactive_engagement", &resource.Sweeper{
		Name: "aws_shield_proactive_engagement",
		F:    sweepProactiveEngagements,
		Dependencies: []string{
			"aws_shield_drt_access_role_arn_association",
		},
	})
}

func sweepDRTAccessLogBucketAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &shield.DescribeDRTAccessInput{}
	conn := client.ShieldClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeDRTAccess(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Shield DRT Log Bucket Association sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Shield DRT Log Bucket Associations (%s): %w", region, err)
	}

	for _, v := range output.LogBucketList {
		log.Printf("[INFO] Deleting Shield DRT Log Bucket Association: %s", v)
		sweepResources = append(sweepResources, framework.NewSweepResource(newDRTAccessLogBucketAssociationResource, client,
			framework.NewAttribute(names.AttrID, v),
		))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Shield DRT Log Bucket Associations (%s): %w", region, err)
	}

	return nil
}

func sweepDRTAccessRoleARNAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &shield.DescribeDRTAccessInput{}
	conn := client.ShieldClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeDRTAccess(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Shield DRT Role ARN Association sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Shield DRT Role ARN Associations (%s): %w", region, err)
	}

	if v := aws.ToString(output.RoleArn); v != "" {
		log.Printf("[INFO] Deleting Shield DRT Role ARN Association: %s", v)
		sweepResources = append(sweepResources, framework.NewSweepResource(newDRTAccessRoleARNAssociationResource, client,
			framework.NewAttribute(names.AttrID, client.AccountID),
		))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Shield DRT Role ARN Associations (%s): %w", region, err)
	}

	return nil
}

func sweepProactiveEngagements(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &shield.DescribeSubscriptionInput{}
	conn := client.ShieldClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeSubscription(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Shield Proactive Engagement sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Shield Proactive Engagements (%s): %w", region, err)
	}

	if output.Subscription.ProactiveEngagementStatus != "" {
		log.Printf("[INFO] Deleting Shield Proactive Engagement")
		sweepResources = append(sweepResources, framework.NewSweepResource(newProactiveEngagementResource, client,
			framework.NewAttribute(names.AttrID, client.AccountID),
		))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Shield Proactive Engagements (%s): %w", region, err)
	}

	return nil
}
