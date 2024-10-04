// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_sns_platform_application", &resource.Sweeper{
		Name: "aws_sns_platform_application",
		F:    sweepPlatformApplications,
	})

	resource.AddTestSweepers("aws_sns_topic", &resource.Sweeper{
		Name: "aws_sns_topic",
		F:    sweepTopics,
		Dependencies: []string{
			"aws_autoscaling_group",
			"aws_backup_vault_notifications",
			"aws_budgets_budget",
			"aws_config_delivery_channel",
			"aws_dax_cluster",
			"aws_db_event_subscription",
			"aws_elasticache_cluster",
			"aws_elasticache_replication_group",
			"aws_glacier_vault",
			"aws_iot_topic_rule",
			"aws_neptune_event_subscription",
			"aws_redshift_event_subscription",
			"aws_s3_bucket",
			"aws_ses_configuration_set",
			"aws_ses_domain_identity",
			"aws_ses_email_identity",
			"aws_ses_receipt_rule_set",
			"aws_sns_platform_application",
		},
	})

	resource.AddTestSweepers("aws_sns_topic_subscription", &resource.Sweeper{
		Name: "aws_sns_topic_subscription",
		F:    sweepTopicSubscriptions,
	})
}

func sweepPlatformApplications(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &sns.ListPlatformApplicationsInput{}
	conn := client.SNSClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sns.NewListPlatformApplicationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SNS Platform Application sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing SNS Platform Applications (%s): %w", region, err)
		}

		for _, v := range page.PlatformApplications {
			r := resourcePlatformApplication()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.PlatformApplicationArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SNS Platform Applications (%s): %w", region, err)
	}

	return nil
}

func sweepTopics(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &sns.ListTopicsInput{}
	conn := client.SNSClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sns.NewListTopicsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SNS Topic sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing SNS Topics (%s): %w", region, err)
		}

		for _, v := range page.Topics {
			r := resourceTopic()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TopicArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SNS Topics (%s): %w", region, err)
	}

	return nil
}

func sweepTopicSubscriptions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &sns.ListSubscriptionsInput{}
	conn := client.SNSClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sns.NewListSubscriptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SNS Topic Subscription sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing SNS Topic Subscriptions (%s): %w", region, err)
		}

		for _, v := range page.Subscriptions {
			arn := aws.ToString(v.SubscriptionArn)

			if arn == "PendingConfirmation" {
				continue
			}

			r := resourceTopicSubscription()
			d := r.Data(nil)
			d.SetId(arn)
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SNS Topic Subscriptions (%s): %w", region, err)
	}

	return nil
}
