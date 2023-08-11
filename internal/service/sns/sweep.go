// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package sns

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	conn := client.SNSConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListPlatformApplicationsPagesWithContext(ctx, input, func(page *sns.ListPlatformApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.PlatformApplications {
			r := ResourcePlatformApplication()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.PlatformApplicationArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SNS Platform Applications sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing SNS Platform Applications: %w", err)
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
	conn := client.SNSConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListTopicsPagesWithContext(ctx, input, func(page *sns.ListTopicsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Topics {
			r := ResourceTopic()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.TopicArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SNS Topics sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing SNS Topics: %w", err)
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
	conn := client.SNSConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListSubscriptionsPagesWithContext(ctx, input, func(page *sns.ListSubscriptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Subscriptions {
			arn := aws.StringValue(v.SubscriptionArn)

			if arn == "PendingConfirmation" {
				continue
			}

			r := ResourceTopicSubscription()
			d := r.Data(nil)
			d.SetId(arn)
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SNS Topic Subscriptions sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing SNS Topic Subscriptions: %w", err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SNS Topic Subscriptions (%s): %w", region, err)
	}

	return nil
}
